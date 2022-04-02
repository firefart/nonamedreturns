package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "nonamedreturnlint",
	Doc:      "Reports all named returns",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// only filter function defintions
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)

		results := funcDecl.Type.Results
		// no return values
		if results == nil {
			return
		}

		resultsList := results.List

		for _, p := range resultsList {
			if len(p.Names) == 0 {
				// all good, the parameter is not named
				continue
			}

			for _, n := range p.Names {
				pass.Reportf(node.Pos(), "named return %s (%s) found in function %s", n.Name, types.ExprString(p.Type), funcDecl.Name.Name)
			}
		}
	})

	return nil, nil
}
