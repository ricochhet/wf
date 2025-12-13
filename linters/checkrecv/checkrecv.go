package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
				continue // Not a method.
			}

			recvField := funcDecl.Recv.List[0]

			var (
				recvTypeName string
				isPointer    bool
			)

			switch expr := recvField.Type.(type) {
			case *ast.StarExpr:
				isPointer = true

				if ident, ok := expr.X.(*ast.Ident); ok {
					recvTypeName = ident.Name
				}
			case *ast.Ident:
				recvTypeName = expr.Name
			default:
				continue // Unknown receiver type.
			}

			methodName := funcDecl.Name.Name
			pos := funcDecl.Name.Pos()

			// Get the full type from the type info.
			obj := pass.TypesInfo.Defs[funcDecl.Name]
			if funcObj, ok := obj.(*types.Func); ok && funcObj.Type() != nil {
				sig, _ := funcObj.Type().(*types.Signature)

				recv := sig.Recv()
				if recv != nil {
					if _, ok := recv.Type().(*types.Pointer); ok {
						isPointer = true
					}
				}
			}

			if isPointer {
				pass.Reportf(
					pos,
					"method %s has a pointer receiver (*%s)",
					methodName,
					recvTypeName,
				)
			} else {
				pass.Reportf(pos, "method %s has a value receiver (%s)", methodName, recvTypeName)
			}
		}
	}

	return nil, nil //nolint:nilnil // wontfix
}
