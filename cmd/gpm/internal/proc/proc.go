package proc

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/ricochhet/gpm/config"
	"github.com/ricochhet/gpm/internal/custom"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/fsutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/ricochhet/pkg/maputil"
)

type Context struct {
	Mu                *sync.Mutex
	Flags             *config.Flags
	SharedProc        *config.ProcManager
	StoredProc        *config.ProcManager
	Builtins          *custom.Builtins
	MaxProcNameLength int
}

// command: Runas. execute a task as a mapped executable name.
func (ctx *Context) Runas(runas []config.Runas) (bool, error) {
	if flag.NArg() != 0 && ctx.Flags.Args[0] != "runas" {
		return false, nil
	}

	var (
		runner bool
		name   string
	)

	if flag.NArg() == 0 {
		exePath, err := os.Executable()
		if err != nil {
			return false, errutil.WithFrame(err)
		}

		name = filepath.Base(exePath)
	} else {
		name = ctx.Flags.Args[1]
	}

	for _, run := range runas {
		if !strings.EqualFold(name, run.Name) &&
			!slices.Contains(run.Aliases, name) {
			continue
		}

		runner = true

		start := []string{"start"}
		if !run.Start {
			start = []string{}
		}

		newCfg := run.Flags
		if run.Port != 0 {
			newCfg.Port = run.Port
		} else {
			newCfg.Port = ctx.Flags.Port
		}

		ctx.Flags = maputil.Merge(ctx.Flags, &newCfg, "json", true)
		ctx.Flags.Args = slices.Concat(start, run.Tasks)
	}

	logutil.Debugf(os.Stdout, "starting as '%s'", name)

	return runner, nil
}

// command: Start. spawn procs.
func (ctx *Context) Start(rpcCtx context.Context, sig <-chan os.Signal, cfg *config.Flags) error {
	rpcCtx, cancel := context.WithCancel(rpcCtx)
	// Cancel the RPC server when procs have returned/errored, cancel the
	// context anyway in case of early return.
	defer cancel()

	if len(cfg.Args) <= 1 {
		return errors.New("no task specified")
	}

	tmp := make([]*config.ProcInfo, 0, len(cfg.Args[1:]))
	ctx.MaxProcNameLength = 0

	for _, v := range cfg.Args[1:] {
		proc := ctx.FindProc(v)
		if proc == nil {
			return errors.New("unknown proc: " + v)
		}

		tmp = append(tmp, proc)

		if len(v) > ctx.MaxProcNameLength {
			ctx.MaxProcNameLength = len(v)
		}
	}

	ctx.Mu.Lock()
	ctx.SharedProc.SetAll(tmp)
	ctx.Mu.Unlock()

	if len(cfg.Envfiles) > 0 {
		var err error

		newEnvfiles := []string{}

		for _, envfile := range cfg.Envfiles {
			if fsutil.Exists(envfile) {
				newEnvfiles = append(newEnvfiles, envfile)
			}
		}

		if cfg.EnvOverload {
			err = godotenv.Overload(newEnvfiles...)
		} else {
			err = godotenv.Load(newEnvfiles...)
		}

		if err != nil && len(newEnvfiles) != 0 {
			return errutil.WithFrame(err)
		}
	}

	rpcChan := make(chan *RPCMessage, 10)

	if cfg.StartRPCServer {
		go func() {
			if err := ctx.StartServer(rpcCtx, rpcChan, cfg.Port); err != nil {
				logutil.Errorf(os.Stderr, "Failed to start RPC server: %v\n", err)
			}
		}()
	}

	//nolint:contextcheck // wontfix
	return ctx.StartProcs(sig, rpcChan, cfg.ExitOnError)
}

// SpawnProc starts the specified proc, and returns any error from running it.
func (ctx *Context) SpawnProc(name string, errCh chan<- error) {
	proc := ctx.FindProc(name)
	if proc == nil {
		return
	}

	logger := logutil.CreateLogger(name, proc.ColorIndex)
	cs := slices.Concat(cmdStart, proc.Cmdline)

	if ok, err := ctx.Builtins.Start(logger, cs[2], *ctx.Flags); ok ||
		err != nil {
		ctx.SpawnProcs(logger, proc.Steps, errCh)

		errCh <- err

		return
	}

	for {
		cmd := exec.CommandContext(context.Background(), cs[0], cs[1:]...)
		cmd.Dir = proc.Dir
		cmd.Stdin = nil

		if ctx.Flags.InheritStdin {
			cmd.Stdin = os.Stdin
		}

		if proc.Fork {
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.SysProcAttr = forkProcAttrs
		} else {
			cmd.Stdout = logger
			cmd.Stderr = logger
			cmd.SysProcAttr = procAttrs
		}

		if proc.Silent {
			cmd.Stdout = io.Discard
			cmd.Stderr = io.Discard
		}

		var err error

		// StartPTY sets cmd.Std to the logger, we can optionally set the logger to nil.
		cleanup, err := ctx.startPTY(logger, cmd)
		if err != nil {
			select {
			case errCh <- err:
			default:
			}

			logutil.Infof(logger, "Failed to open pty for %s: %s\n", name, err)
		}
		defer cleanup()

		if proc.SetPort {
			cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", proc.Port))
			logutil.Infof(logger, "Starting %s on port %d\n", name, proc.Port)
		}

		for key, value := range proc.Env {
			cmd.Env = append(cmd.Env, fsutil.CombineEnviron(key, value))
			logutil.Debugf(logger, "added envar: %s=%s\n", key, value)
		}

		logutil.Debugf(logger, "cmd: %s\n", cmd.String())

		if err := cmd.Start(); err != nil {
			select {
			case errCh <- err:
			default:
			}

			logutil.Infof(logger, "Failed to start %s: %s\n", name, err)

			return
		}

		proc.Cmd = cmd
		proc.StoppedBySupervisor = false

		if !proc.Fork {
			proc.Mu.Unlock()

			err = cmd.Wait()

			proc.Mu.Lock()
		}

		proc.Cond.Broadcast()

		if err != nil && !proc.StoppedBySupervisor {
			select {
			case errCh <- err:
			default:
			}
		}

		proc.WaitErr = err
		proc.Cmd = nil

		logutil.Infof(logger, "Terminating %s\n", name)
		ctx.SpawnProcs(logger, proc.Steps, errCh)

		if proc.StoppedBySupervisor || !proc.RestartOnError || err == nil {
			break
		}

		logutil.Infof(logger, "Restarting %s\n", name)
	}
}

// SpawnProcs starts the specified procs, and returns any error from running it.
func (ctx *Context) SpawnProcs(logger *logutil.Logger, names []string, errCh chan<- error) {
	if len(names) == 0 {
		return
	}

	ctx.Flags.StartRPCServer = false // Server already started.

	for _, name := range names {
		if ok, err := ctx.Builtins.Start(logger, name, *ctx.Flags); ok ||
			err != nil {
			errCh <- err

			continue
		}

		ctx.SharedProc.CopyFrom(ctx.StoredProc)

		ctx.Flags.Args = append([]string{""}, name)

		nc, stop := NotifyCh()
		defer stop()

		err := ctx.Start(context.Background(), nc, ctx.Flags)
		errCh <- err
	}
}

// StartProc starts the specified proc, if proc is started already, return nil.
func (ctx *Context) StartProc(name string, wg *sync.WaitGroup, errCh chan<- error) error {
	proc := ctx.FindProc(name)
	if proc == nil {
		return errors.New("unknown name: " + name)
	}

	proc.Mu.Lock()

	if proc.Cmd != nil {
		proc.Mu.Unlock()
		return nil
	}

	if wg != nil {
		wg.Add(1)
	}

	go func() {
		ctx.SpawnProc(name, errCh)

		if wg != nil {
			wg.Done()
		}

		proc.Mu.Unlock()
	}()

	return nil
}

// StartProcs starts all procs.
func (ctx *Context) StartProcs(
	sc <-chan os.Signal,
	rpcCh <-chan *RPCMessage,
	exitOnError bool,
) error {
	var wg sync.WaitGroup

	errCh := make(chan error, 1)

	for _, proc := range ctx.SharedProc.All() {
		if err := ctx.StartProc(proc.Name, &wg, errCh); err != nil {
			return errutil.New("StartProc", err)
		}

		if ctx.Flags.Interval > 0 {
			time.Sleep(time.Second * time.Duration(ctx.Flags.Interval))
		}
	}

	allProcsDone := make(chan struct{}, 1)

	if ctx.Flags.ExitOnStop {
		go func() {
			wg.Wait()

			allProcsDone <- struct{}{}
		}()
	}

	for {
		select {
		case rpcMsg := <-rpcCh:
			switch rpcMsg.Msg {
			// TODO: add more events here.
			case "stop":
				for _, proc := range rpcMsg.Args {
					if err := ctx.StopProc(proc, nil); err != nil {
						rpcMsg.ErrCh <- err
						break
					}
				}

				close(rpcMsg.ErrCh)
			default:
				panic("unimplemented rpc message type " + rpcMsg.Msg)
			}
		case err := <-errCh:
			if exitOnError {
				if err := ctx.StopProcs(os.Interrupt); err != nil {
					return errutil.New("StopProcs", err)
				}

				return errutil.WithFrame(err)
			}
		case <-allProcsDone:
			return ctx.StopProcs(os.Interrupt)
		case sig := <-sc:
			return ctx.StopProcs(sig)
		}
	}
}

// RestartProc restarts the proc by name.
func (ctx *Context) RestartProc(name string) error {
	err := ctx.StopProc(name, nil)
	if err != nil {
		return errutil.WithFrame(err)
	}

	return ctx.StartProc(name, nil, nil)
}

// StopProcs attempts to stop every running process and returns any non-nil
// error, if one exists. StopProcs will wait until all procs have had an
// opportunity to stop.
func (ctx *Context) StopProcs(sig os.Signal) error {
	var err error

	if ctx.Flags.ReverseOnStop {
		procs := ctx.SharedProc.All()
		proclen := len(procs)
		reversed := make([]*config.ProcInfo, proclen)

		for i := range proclen {
			reversed[i] = procs[proclen-1-i]
		}

		ctx.SharedProc.SetAll(reversed)
	}

	for _, proc := range ctx.SharedProc.All() {
		stopErr := ctx.StopProc(proc.Name, sig)
		if stopErr != nil {
			err = stopErr
		}

		if ctx.Flags.Interval > 0 {
			time.Sleep(time.Second * time.Duration(ctx.Flags.Interval))
		}
	}

	return errutil.WithFrame(err)
}

// StopProc stops the specified proc, issuing os.Kill if it does not terminate within 10
// seconds. If signal is nil, os.Interrupt is used.
func (ctx *Context) StopProc(name string, signal os.Signal) error {
	if signal == nil {
		signal = os.Interrupt
	}

	proc := ctx.FindProc(name)
	if proc == nil {
		return errors.New("unknown proc: " + name)
	}

	proc.Mu.Lock()
	defer proc.Mu.Unlock()

	if proc.Cmd == nil {
		return nil
	}

	proc.StoppedBySupervisor = true

	err := terminateProc(proc, signal)
	if err != nil {
		return errutil.New("terminateProc", err)
	}

	timeout := time.AfterFunc(10*time.Second, func() {
		proc.Mu.Lock()
		defer proc.Mu.Unlock()

		if proc.Cmd != nil {
			err = killProc(proc.Cmd.Process)
		}
	})

	proc.Cond.Wait()
	timeout.Stop()

	return errutil.WithFrame(err)
}

// command: check. show Taskfile entries.
func (ctx *Context) Check() error {
	ctx.Mu.Lock()
	defer ctx.Mu.Unlock()

	keys := make([]string, len(ctx.SharedProc.All()))
	for i, proc := range ctx.SharedProc.All() {
		if proc.Desc != "" {
			keys[i] = proc.Name + ": " + proc.Desc
		} else {
			keys[i] = proc.Name
		}
	}

	sort.Strings(keys)
	logutil.Infof(os.Stdout, "Valid taskfile detected (%s)\n", strings.Join(keys, ", "))

	return nil
}

// FindProc finds the process in the slice by name.
func (ctx *Context) FindProc(name string) *config.ProcInfo {
	ctx.Mu.Lock()
	defer ctx.Mu.Unlock()

	for _, proc := range ctx.SharedProc.All() {
		if proc.Name == name {
			return proc
		}

		if slices.Contains(proc.Aliases, name) {
			return proc
		}
	}

	return nil
}
