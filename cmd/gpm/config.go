package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/ricochhet/gpm/config"
	"github.com/ricochhet/pkg/cmdutil"
	"github.com/ricochhet/pkg/cueutil"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/fsutil"
	"github.com/ricochhet/pkg/logutil"
)

// readConfig reads the '.gpm' configuration file and returns config.Flags.
func readConfig() *config.Flags {
	cfg := Flag

	cfg.Args = flag.Args()
	cfg.Envfiles = strings.FieldsFunc(cfg.Envfile, func(c rune) bool {
		return c == ','
	})

	logutil.SetDebug(cfg.Debug)

	if err := cmdutil.QuickEdit(cfg.QuickEdit); err != nil {
		logutil.Errorf(os.Stderr, "Failed to set Quick Edit mode: %v\n", err)
	}

	path, err := maybeGlobalDotfile(cfg)
	if err != nil {
		return cfg
	}

	if _, err := cueutil.NewDefaultUnmarshal[*config.Flags]().File(path, &cfg); err != nil {
		logutil.Errorf(os.Stderr, "Failed to read config: %v\n", err)
		return cfg
	}

	return cfg
}

// maybeGlobalDotfile returns the path of the dotfile file to use.
// Global dotfile is set next to the executable, regardless of the working directory.
func maybeGlobalDotfile(cfg *config.Flags) (string, error) {
	if !dotfileFlag.IsIn(cfg.Global) {
		if !fsutil.Exists(cfg.Dotfile) {
			return cfg.Dotfile, errutil.Newf(
				"fsutil.Exists",
				"path does not exist: %s",
				cfg.Dotfile,
			)
		}

		logutil.Debugf(os.Stdout, "using local dotfile: %s\n", cfg.Dotfile)

		return cfg.Dotfile, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return cfg.Dotfile, errutil.New("os.Executable", err)
	}

	path := filepath.Join(filepath.Dir(exePath), cfg.Dotfile)
	if !fsutil.Exists(path) {
		return cfg.Dotfile, errutil.Newf("fsutil.Exists", "path does not exist: %s", path)
	}

	logutil.Debugf(os.Stdout, "using global dotfile %s\n", path)

	return path, nil
}

// readTaskfile reads the taskfile specified by Flags.Taskfile and
// adds proc info to the SharedProc ProcManager.
func readTaskfile() error {
	var err error

	path, err := maybeGlobalTaskfile(ctx.Flags.Taskfile)
	if err != nil {
		return errutil.WithFrame(err)
	}

	unmarshaler := cueutil.NewDefaultUnmarshal[config.Taskfile]()
	unmarshaler.Compile(*cueutil.NewBuiltins([]string{}).Map())

	ataskfile, _, err = unmarshaler.NewFile(path)
	if err != nil {
		return errutil.New("taskfileUnmarshaler.NewFile", err)
	}

	for _, include := range ataskfile.Includes {
		pathInclude, err := maybeGlobalTaskfile(include)
		if err != nil {
			return errutil.WithFrame(err)
		}

		taskfile, _, err := unmarshaler.NewFile(pathInclude)
		if err != nil {
			return errutil.New("taskfileUnmarshaler.NewFile.include", err)
		}

		ataskfile = ataskfile.Merge(taskfile)
	}

	mu.Lock()
	defer mu.Unlock()

	ctx.Builtins.SetArtifacts(ataskfile.Artifacts)

	index := 0

	for _, task := range ataskfile.Tasks {
		if len(task.Platforms) != 0 && !slices.Contains(task.Platforms, runtime.GOOS) {
			continue
		}

		name := strings.TrimSpace(task.Name)

		proc := &config.ProcInfo{
			Name:           name,
			Desc:           task.Desc,
			Aliases:        task.Aliases,
			Cmdline:        task.Cmd,
			Env:            ataskfile.Env,
			Steps:          task.Steps,
			Dir:            task.Dir,
			Fork:           task.Fork,
			Silent:         task.Silent,
			ColorIndex:     index,
			RestartOnError: ctx.Flags.RestartOnError,
			InheritStdin:   ctx.Flags.InheritStdin,
		}
		if ctx.Flags.SetPorts {
			proc.SetPort = true
			proc.Port = ctx.Flags.BasePort
			ctx.Flags.BasePort += 100
		}

		proc.Cond = sync.NewCond(&proc.Mu)
		ctx.SharedProc.Add(proc)

		if len(name) > ctx.MaxProcNameLength {
			ctx.MaxProcNameLength = len(name)
		}

		index = (index + 1) % len(logutil.Colors)
	}

	if len(ctx.SharedProc.All()) == 0 {
		return errors.New("no valid entry")
	}

	ctx.StoredProc.CopyFrom(ctx.SharedProc)

	return nil
}

// maybeGlobalTaskfile returns the path of the taskfile file to use.
// Global taskfile is set next to the executable, regardless of the working directory.
func maybeGlobalTaskfile(taskfile string) (string, error) {
	if !taskfileFlag.IsIn(ctx.Flags.Global) {
		if !fsutil.Exists(taskfile) {
			return taskfile, errutil.Newf("fsutil.Exists", "path does not exist: %s", taskfile)
		}

		logutil.Debugf(os.Stdout, "using local taskfile: %s\n", taskfile)

		return taskfile, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return taskfile, errutil.New("os.Executable", err)
	}

	path := filepath.Join(filepath.Dir(exePath), taskfile)
	if !fsutil.Exists(path) {
		return taskfile, errutil.Newf("fsutil.Exists", "path does not exist: %s", path)
	}

	logutil.Debugf(os.Stdout, "using global taskfile: %s\n", path)

	return path, nil
}
