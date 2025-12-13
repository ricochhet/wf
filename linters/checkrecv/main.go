package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

const linterName = "checkrecv"

var Analyzer = &analysis.Analyzer{
	Name: linterName,
	Doc:  "reports whether a method uses a pointer or value receiver",
	Run:  run,
}

func main() {
	singlechecker.Main(Analyzer)
}
