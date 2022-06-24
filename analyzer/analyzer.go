package analyzer

import (
	"flag"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const FlagAllowErrorInDefer = "allow-error-in-defer"

var Analyzer = &analysis.Analyzer{
	Name:     "nonamedreturns",
	Doc:      "Reports all named returns",
	Flags:    flags(),
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func flags() flag.FlagSet {
	fs := flag.FlagSet{}
	fs.Bool(FlagAllowErrorInDefer, false, "do not complain about named error, if it is assigned inside defer")
	return fs
}

func run(pass *analysis.Pass) (interface{}, error) {
	allowErrorInDefer := pass.Analyzer.Flags.Lookup(FlagAllowErrorInDefer).Value.String() == "true"

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// only filter function defintions
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		var funcResults *ast.FieldList
		var funcBody *ast.BlockStmt

		switch n := node.(type) {
		case *ast.FuncLit:
			funcResults = n.Type.Results
			funcBody = n.Body
		case *ast.FuncDecl:
			funcResults = n.Type.Results
			funcBody = n.Body
		default:
			return
		}

		// no return values
		if funcResults == nil {
			return
		}

		resultsList := funcResults.List

		for _, p := range resultsList {
			if len(p.Names) == 0 {
				// all good, the parameter is not named
				continue
			}

			for _, n := range p.Names {
				if n.Name == "_" {
					continue
				}

				if allowErrorInDefer {
					if pass.TypesInfo.TypeOf(p.Type).String() == "error" { // with package prefix, so only built-it error fits
						if findDeferWithVariableAssignment(funcBody, pass.TypesInfo, pass.TypesInfo.ObjectOf(n)) {
							continue
						}
					}
				}

				pass.Reportf(node.Pos(), "named return %q with type %q found", n.Name, types.ExprString(p.Type))
			}
		}
	})

	return nil, nil
}

func findDeferWithVariableAssignment(body *ast.BlockStmt, info *types.Info, variable types.Object) bool {
	found := false

	ast.Inspect(body, func(node ast.Node) bool {
		if found {
			return false // stop inspection
		}

		if d, ok := node.(*ast.DeferStmt); ok {
			if fn, ok2 := d.Call.Fun.(*ast.FuncLit); ok2 {
				if findVariableAssignment(fn.Body, info, variable) {
					found = true
					return false
				}
			}
		}

		return true
	})

	return found
}

func findVariableAssignment(body *ast.BlockStmt, info *types.Info, variable types.Object) bool {
	found := false

	ast.Inspect(body, func(node ast.Node) bool {
		if found {
			return false // stop inspection
		}

		if a, ok := node.(*ast.AssignStmt); ok {
			for _, lh := range a.Lhs {
				if i, ok2 := lh.(*ast.Ident); ok2 {
					if info.ObjectOf(i) == variable {
						found = true
						return false
					}
				}
			}
		}

		return true
	})

	return found
}
