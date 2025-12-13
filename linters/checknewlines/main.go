package main

import (
	"flag"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

const linterName = "checknewlines"

var warnIfMissing bool

var Analyzer = &analysis.Analyzer{
	Name: linterName,
	Doc:  "checks for \\n in format strings of fmt.Printf, fmt.Sprintf, and fmt.Fprintf",
	Run:  run,
	Flags: func() flag.FlagSet {
		var fs flag.FlagSet
		fs.BoolVar(
			&warnIfMissing,
			"warn-if-missing",
			false,
			"warn if \\n is missing instead of present",
		)
		return fs
	}(),
}

func main() {
	singlechecker.Main(Analyzer)
}
