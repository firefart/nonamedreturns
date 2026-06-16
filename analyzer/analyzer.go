package analyzer

import (
	"errors"
	"flag"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const FlagReportErrorInDefer = "report-error-in-defer"

var Analyzer = &analysis.Analyzer{
	Name:     "nonamedreturns",
	Doc:      "Reports all named returns",
	Flags:    flags(),
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func flags() flag.FlagSet {
	fs := flag.FlagSet{}
	fs.Bool(FlagReportErrorInDefer, false, "report named error if it is assigned inside defer")
	return fs
}

func run(pass *analysis.Pass) (any, error) {
	reportErrorInDefer := pass.Analyzer.Flags.Lookup(FlagReportErrorInDefer).Value.String() == "true"
	errorType := types.Universe.Lookup("error").Type()

	inspector, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, errors.New("failed to get inspector")
	}

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

		// Function without body, ex: https://github.com/golang/go/blob/master/src/internal/syscall/unix/net.go
		if funcBody == nil {
			return
		}

		// no return values
		if funcResults == nil {
			return
		}

		// deferUsed and assigned are computed lazily and at most once per
		// function: the relatively expensive body walk only happens when there
		// is an error-typed named return whose defer exemption we need to check.
		var deferUsed, assigned map[types.Object]bool

		for _, p := range funcResults.List {
			if len(p.Names) == 0 {
				// all good, the parameter is not named
				continue
			}

			isError := types.Identical(pass.TypesInfo.TypeOf(p.Type), errorType)

			for _, n := range p.Names {
				if n.Name == "_" {
					continue
				}

				// A named error return is allowed when it is referenced inside a defer
				// (e.g. to inspect or modify it before returning) as long as it is also
				// assigned somewhere in the function body. The assignment may happen
				// inside the defer itself or anywhere else in the function.
				if !reportErrorInDefer && isError {
					if deferUsed == nil {
						deferUsed, assigned = collectDeferUsageAndAssignments(funcBody, pass.TypesInfo)
					}

					obj := pass.TypesInfo.ObjectOf(n)
					if deferUsed[obj] && assigned[obj] {
						continue
					}
				}

				pass.Reportf(n.Pos(), "named return %q with type %q found", n.Name, types.ExprString(p.Type))
			}
		}
	})

	return nil, nil // nolint:nilnil
}

// collectDeferUsageAndAssignments walks body a single time and returns:
//   - deferUsed: objects referenced (read or written) inside a deferred func literal
//   - assigned:  objects that appear on the left-hand side of an assignment
//     anywhere in the body, including inside a deferred func literal
//
// Variable shadowing is handled naturally: a shadowed declaration introduces a
// distinct types.Object, so it does not match the outer named return.
func collectDeferUsageAndAssignments(body *ast.BlockStmt, info *types.Info) (map[types.Object]bool, map[types.Object]bool) {
	deferUsed := make(map[types.Object]bool)
	assigned := make(map[types.Object]bool)

	ast.Inspect(body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.AssignStmt:
			for _, lh := range n.Lhs {
				if i, ok := lh.(*ast.Ident); ok {
					if obj := info.ObjectOf(i); obj != nil {
						assigned[obj] = true
					}
				}
			}
		case *ast.DeferStmt:
			if fn, ok := n.Call.Fun.(*ast.FuncLit); ok {
				ast.Inspect(fn.Body, func(inner ast.Node) bool {
					if i, ok := inner.(*ast.Ident); ok {
						if obj := info.ObjectOf(i); obj != nil {
							deferUsed[obj] = true
						}
					}
					return true
				})
			}
		}

		// keep descending so assignments (and nested defers) inside a deferred
		// closure are still collected by the outer walk
		return true
	})

	return deferUsed, assigned
}
