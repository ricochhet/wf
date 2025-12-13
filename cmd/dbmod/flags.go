package main

import (
	"flag"

	"github.com/ricochhet/dbmod/config"
	"github.com/ricochhet/pkg/flagutil"
)

var Flag = NewFlags()

// NewFlags creates an empty Flags.
func NewFlags() *config.Flags {
	return &config.Flags{}
}

var dotfileFlag = flagutil.NewOptional("d", "Dotfile")

//nolint:gochecknoinits // wontfix
func init() {
	registerFlags(flag.CommandLine, Flag)
	flag.Parse()
}

// registerFlags registers all flags to the flagset.
func registerFlags(fs *flag.FlagSet, f *config.Flags) {
	fs.StringVar(&f.Dotfile, "dotfile", ".dbmod.cue", "dotfile")
	fs.BoolVar(
		&f.DryRun,
		"dry-run",
		false,
		"commit changes locally, without modifying the database",
	)
	fs.StringVar(&f.MongoURI, "uri", "mongodb://localhost:27017", "MongoDB URI")
	fs.StringVar(&f.Database, "db-name", "openWF", "MongoDB database name")
	fs.StringVar(&f.DBData, "db-data", "dbdata", "MongoDB collection backup and dry-run path")
	fs.StringVar(&f.WFData, "wf-data", "assets", "Warframe public export path")
	fs.IntVar(&f.Index, "i", 0, "index of the document to modify")
	fs.StringVar(&f.Mode, "m", "cheat", "mode to use for editing (c | cheat, p | patch)")
	fs.StringVar(&f.Global, "g", flagutil.Set("", dotfileFlag), "use global dotfile")
	fs.BoolVar(&f.LogTime, "logtime", true, "show timestamp in log")
	fs.BoolVar(&f.Debug, "debug", false, "enable debug mode")
	fs.BoolVar(&f.QuickEdit, "quick-edit", false, "enable quick edit mode")
}
