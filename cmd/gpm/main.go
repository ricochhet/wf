package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ricochhet/gpm/config"
	"github.com/ricochhet/gpm/internal/custom"
	"github.com/ricochhet/gpm/internal/proc"
	"github.com/ricochhet/pkg/cmdutil"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/ricochhet/pkg/maputil"
)

// version is the git tag at the time of build and is used to denote the
// binary's current version. This value is supplied as an ldflag at compile
// time by goreleaser (see .goreleaser.yml).
const (
	name     = "gpm"
	dotfile  = ".gpm"
	version  = "0.3.16"
	revision = "HEAD"
)

func usage() {
	fmt.Fprint(os.Stderr, `Tasks:
  gpm console                    # Start a minimal command console
  gpm check                      # Show entries in Taskfile
  gpm help [TASK]                # Show this help
  gpm export [FORMAT] [LOCATION] # Export the apps to another process
                                       (upstart)
  gpm run COMMAND [PROCESS...]   # Run a command
                                       start
                                       stop
                                       stop-all
                                       restart
                                       restart-all
                                       list
                                       status
  gpm start [PROCESS]            # Start the application
  gpm runas [PROCESS]            # Run a runas process
  gpm version                    # Display gpm version

Options:
`)
	flag.PrintDefaults()

	if console {
		return
	}

	os.Exit(0)
}

var (
	mu        sync.Mutex
	ataskfile config.Taskfile
	ctx       proc.Context
	console   = false
)

// showVersion shows the current version of gpm.
func showVersion() {
	logutil.Infof(os.Stdout, "%s\n", version)

	if console {
		return
	}

	os.Exit(0)
}

func main() {
	var err error

	cfg := readConfig()

	logutil.LogTime.Store(cfg.LogTime)
	logutil.MaxProcNameLength.Store(0)

	if cfg.BaseDir != "" {
		err = os.Chdir(cfg.BaseDir)
		if err != nil {
			logutil.Errorf(os.Stderr, "gpm: %v\n", err)
			os.Exit(1)
		}
	}

	ctx = proc.Context{
		Mu:         &mu,
		Flags:      cfg,
		SharedProc: config.NewProcManager(),
		StoredProc: config.NewProcManager(),
		Builtins:   custom.NewDefaultBuiltins(),
	}

	err = readTaskfile()
	exitOnErr(err)

	runner, err := ctx.Runas(ataskfile.Runas)
	exitOnErr(err)

	if (flag.NArg() == 0 && !runner) || len(ctx.Flags.Args) == 0 {
		usage()
	}

	cmd := ctx.Flags.Args[0]
	switch cmd {
	case "console":
		console = true
		fs := flag.NewFlagSet("console", flag.ContinueOnError)

		var f config.Flags

		cmdutil.NewScanner(func(i string) error {
			registerFlags(fs, &f)

			if err := fs.Parse(strings.Fields(i)); err != nil {
				return errutil.New("fs.Parse", err)
			}

			ctx.Flags = maputil.Merge(ctx.Flags, &f, "json", true)
			ctx.Flags.Args = fs.Args()
			cmd := ctx.Flags.Args[0]

			if strings.EqualFold(cmd, "exit") || strings.EqualFold(cmd, "q") {
				os.Exit(0)
				return nil
			}

			ctx.SharedProc.CopyFrom(ctx.StoredProc)

			return commands()
		})
	default:
		err = commands()
	}

	exitOnErr(err)
}

// commands runs functions based on the provided Flags.Args[0].
func commands() error {
	var err error

	cmd := ctx.Flags.Args[0]
	switch cmd {
	case "check":
		err = ctx.Check()
	case "help":
		usage()
	case "run":
		if len(ctx.Flags.Args) >= 2 {
			cmd, args := ctx.Flags.Args[1], ctx.Flags.Args[2:]
			err = proc.Run(cmd, args, ctx.Flags.Port)
		} else {
			usage()
		}
	case "export":
		if len(ctx.Flags.Args) == 3 {
			format, path := ctx.Flags.Args[1], ctx.Flags.Args[2]
			err = export(format, path)
		} else {
			usage()
		}
	case "start":
		nc, stop := proc.NotifyCh()
		defer stop()

		err = ctx.Start(context.Background(), nc, ctx.Flags)
	case "runas":
		_, err = ctx.Runas(ataskfile.Runas) // Returned boolean is unneeded here.
	case "version":
		showVersion()
	default:
		usage()
	}

	return errutil.WithFrame(err)
}

// exitOnErr prints an error message and exits the program.
func exitOnErr(err error) {
	if err != nil {
		logutil.Errorf(os.Stderr, "%s: %v\n", os.Args[0], err.Error())
		os.Exit(1)
	}
}
