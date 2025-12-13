package main

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"

	"github.com/ricochhet/pkg/linters"
	"golang.org/x/tools/go/analysis"
)

var check = map[string]bool{
	"fmt.Printf": true,
	// "fmt.Sprintf": true,
	"fmt.Fprintf": true,
}

// run checks if a print statement contains newlines.
func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Get the selector. (fmt.Printf)
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			// Get function object from type info.
			obj := pass.TypesInfo.Uses[sel.Sel]

			funcObj, ok := obj.(*types.Func)
			if !ok || funcObj.Pkg() == nil {
				return true
			}

			fullName := funcObj.Pkg().Path() + "." + funcObj.Name()
			if !check[fullName] {
				return true
			}

			// Determine the format string argument index.
			formatArgIndex := 0
			if funcObj.Name() == "Fprintf" {
				formatArgIndex = 1
			}

			if len(call.Args) <= formatArgIndex {
				return true
			}

			formatArg := call.Args[formatArgIndex]

			// Check if it's a string literal.
			basicLit, ok := formatArg.(*ast.BasicLit)
			if !ok || basicLit.Kind != token.STRING {
				return true
			}

			// Get constant string value.
			val := pass.TypesInfo.Types[formatArg].Value
			if val == nil {
				return true
			}

			str := constant.StringVal(val)

			// Check for newline.
			hasNewline := strings.ContainsRune(str, '\n') || strings.Contains(str, "\n")

			if linters.Directive(pass, call, linterName, "nolint") {
				return true
			}

			if hasNewline && !warnIfMissing {
				pass.Reportf(formatArg.Pos(), "format string contains \\n")
				return true
			}

			if !hasNewline && warnIfMissing {
				pass.Reportf(formatArg.Pos(), "format string is missing \\n")
			}

			return true
		})
	}

	return nil, nil //nolint:nilnil // wontfix
}
