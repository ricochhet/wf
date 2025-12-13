package linters

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Directive checks if a nolint directive is present.
func Directive(pass *analysis.Pass, node ast.Node, linterName, directive string) bool {
	for _, cg := range pass.Files {
		for _, group := range cg.Comments {
			for _, comment := range group.List {
				if strings.Contains(comment.Text, directive) &&
					strings.Contains(comment.Text, linterName) {
					// Check if it's the same file, and close to the relevant node.
					if group.Pos() <= node.Pos() && node.Pos() <= group.End()+3 {
						return true
					}
				}
			}
		}
	}

	return false
}
