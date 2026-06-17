package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const (
	mainPackageName = "main"
	mainFuncName    = "main"

	osPackagePath = "os"
	osExitFunc    = "Exit"

	errMsg = "direct call to os.Exit in main function is prohibited"
)

var OSExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "reports direct calls to os.Exit inside main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != mainPackageName {
		return nil, nil
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if fd.Recv != nil || fd.Name.Name != mainFuncName {
				continue
			}

			if fd.Body == nil {
				continue
			}

			ast.Inspect(fd.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				id, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				obj := pass.TypesInfo.Uses[id]
				if obj == nil {
					return true
				}

				pkg, ok := obj.(*types.PkgName)
				if !ok {
					return true
				}

				if pkg.Imported().Path() == osPackagePath &&
					sel.Sel.Name == osExitFunc {
					pass.Reportf(call.Pos(), errMsg)
				}

				return true
			})
		}
	}

	return nil, nil
}
