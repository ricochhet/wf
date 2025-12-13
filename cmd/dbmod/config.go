package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/ricochhet/dbmod/config"
	"github.com/ricochhet/pkg/cmdutil"
	"github.com/ricochhet/pkg/cueutil"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/fsutil"
	"github.com/ricochhet/pkg/logutil"
)

// readConfig reads the '.dbmod' configuration file and returns config.Flags.
func readConfig() *config.Flags {
	cfg := Flag

	logutil.SetDebug(cfg.Debug)

	cfg.Args = flag.Args()
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
