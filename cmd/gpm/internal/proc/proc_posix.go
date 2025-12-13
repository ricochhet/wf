//go:build !windows
// +build !windows

package proc

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/ricochhet/gpm/config"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/logutil"
	"golang.org/x/sys/unix"
)

const (
	sigint  = unix.SIGINT
	sigterm = unix.SIGTERM
	sighup  = unix.SIGHUP
)

var (
	cmdStart      = []string{"/bin/sh", "-c"}
	procAttrs     = &unix.SysProcAttr{Setpgid: true}
	forkProcAttrs = &syscall.SysProcAttr{
		Setsid: true,
	}
)

// terminateProc terminates the process by sending the signal to the process.
func terminateProc(proc *config.ProcInfo, signal os.Signal) error {
	p := proc.Cmd.Process
	if p == nil {
		return nil
	}

	pgid, err := unix.Getpgid(p.Pid)
	if err != nil {
		return errutil.New("unix.Getpgid", err)
	}

	// Use pgid, ref: http://unix.stackexchange.com/questions/14815/process-descendants
	pid := p.Pid
	if pgid == p.Pid {
		pid = -1 * pid
	}

	target, err := os.FindProcess(pid)
	if err != nil {
		return errutil.New("os.FindProcess", err)
	}

	return target.Signal(signal)
}

// killProc kills the proc with pid, as well as its children.
func killProc(process *os.Process) error {
	return unix.Kill(-1*process.Pid, unix.SIGKILL)
}

// NotifyCh create the terminate/interrupt notifier.
func NotifyCh() (<-chan os.Signal, func()) {
	sc := make(chan os.Signal, 10)
	signal.Notify(sc, sigterm, sigint, sighup)

	return sc, func() {
		signal.Stop(sc)
	}
}

// startPTY starts a PTY (posix only).
func (ctx *Context) startPTY(logger *logutil.Logger, cmd *exec.Cmd) (func(), error) {
	if ctx.Flags.Pty {
		p, t, err := pty.Open()
		if err != nil {
			return nil, errutil.New("pty.Open", err)
		}

		cmd.Stdout = t
		cmd.Stderr = t

		go func() {
			if _, err := io.Copy(logger, p); err != nil && !errors.Is(err, io.EOF) {
				logutil.Errorf(os.Stderr, "io.Copy: %v", err)
			}
		}()

		cleanup := func() {
			if err := p.Close(); err != nil {
				logutil.Errorf(os.Stderr, "p.Close: %v", err)
			}

			if err := t.Close(); err != nil {
				logutil.Errorf(os.Stderr, "t.Close: %v", err)
			}
		}

		return cleanup, nil
	}

	cmd.Stdout = logger
	cmd.Stderr = logger

	return func() {}, nil
}
