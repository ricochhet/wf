package main

import (
	"flag"

	"github.com/ricochhet/gpm/config"
	"github.com/ricochhet/pkg/flagutil"
)

var Flag = NewFlags()

// NewFlags creates an empty Flags.
func NewFlags() *config.Flags {
	return &config.Flags{}
}

//nolint:gochecknoinits // wontfix
func init() {
	registerFlags(flag.CommandLine, Flag)
	flag.Parse()
}

var (
	taskfileFlag = flagutil.NewOptional("f", "Taskfile")
	dotfileFlag  = flagutil.NewOptional("d", "Dotfile")
)

// registerFlags registers all flags to the flagset.
func registerFlags(fs *flag.FlagSet, f *config.Flags) {
	fs.StringVar(&f.Taskfile, "f", "Taskfile.cue", "task file")
	fs.StringVar(&f.Dotfile, "dotfile", ".gpm.cue", "dotfile")
	fs.StringVar(&f.Envfile, "env", ".env", "env files to load (comma separated)")
	fs.BoolVar(&f.EnvOverload, "env-overload", false, "Overload system env with local env")
	fs.UintVar(&f.Port, "p", config.DefaultPort(), "port")
	fs.BoolVar(
		&f.StartRPCServer,
		"rpc-server",
		true,
		"Start an RPC server listening on "+config.DefaultAddr(),
	)
	fs.StringVar(&f.BaseDir, "basedir", "", "base directory")
	fs.UintVar(&f.BasePort, "b", 5000, "base number of port")
	fs.BoolVar(
		&f.SetPorts,
		"set-ports",
		true,
		"False to avoid setting PORT env var for each subprocess",
	)
	fs.BoolVar(
		&f.RestartOnError,
		"restart-on-error",
		false,
		"Restart subprocess if a subprocess quits with a nonzero return code",
	)
	fs.BoolVar(
		&f.ExitOnError,
		"exit-on-error",
		false,
		"Exit gpm if a subprocess quits with a nonzero return code",
	)
	fs.BoolVar(&f.ExitOnStop, "exit-on-stop", true, "Exit gpm if all subprocesses stop")
	fs.BoolVar(&f.LogTime, "logtime", true, "show timestamp in log")
	fs.BoolVar(&f.Pty, "pty", false, "Use a PTY for all subprocesses (noop on Windows)")
	fs.UintVar(&f.Interval, "interval", 0, "the interval at which to start applications")
	fs.BoolVar(&f.ReverseOnStop, "reverse-on-stop", false, "reverse procs sort when stop")
	fs.BoolVar(&f.InheritStdin, "inherit-stdin", false, "inherit stdin from gpm")
	fs.IntVar(&f.VarPasses, "var-passes", 3, "maximum passes variables will do while parsing")
	fs.StringVar(&f.Global, "g", flagutil.Set("", dotfileFlag), "use global dotfile or taskfile")
	fs.BoolVar(&f.Debug, "debug", false, "enable debug mode")
	fs.BoolVar(&f.QuickEdit, "quick-edit", false, "enable quick edit mode")
	fs.BoolVar(&f.Optionals, "optionals", false, "download optional artifacts")
}
