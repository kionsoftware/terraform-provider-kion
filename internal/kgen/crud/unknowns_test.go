package crud_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every Create and Update must resolve unknowns before it writes state.
//
// An Optional+Computed attribute is unknown at plan time when the config omits
// it, and the read-back after create assigns only what the API returns. A
// write-only input, or any field the API never echoes, therefore reaches state
// still unknown -- which Terraform rejects outright:
//
//	Provider returned invalid result object after apply
//
// The first acceptance-test run of this provider hit that on 17 attributes
// across 8 resources. It is invisible to build, vet, lint and the unit tests,
// because nothing below an actual apply ever inspects a post-apply value. This
// test encodes the fix structurally, so a new archetype template cannot ship
// without it.
func TestCreateAndUpdateResolveUnknowns(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..", "service")
	entries, err := os.ReadDir(root)
	require.NoError(t, err)

	checked := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pkgDir := filepath.Join(root, e.Name())
		files, err := filepath.Glob(filepath.Join(pkgDir, "*.go"))
		require.NoError(t, err)

		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, file, nil, 0)
			require.NoError(t, err)

			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv == nil || fn.Body == nil {
					continue
				}
				if fn.Name.Name != "Create" && fn.Name.Name != "Update" {
					continue
				}
				// Only the methods that build state from the plan; Read sets
				// from prior state, where nothing is ever unknown.
				if !setsStateFromPlan(fn.Body) {
					continue
				}
				checked++
				assert.Truef(t, callsResolveUnknowns(fn.Body),
					"%s: %s does not call flex.ResolveUnknowns before resp.State.Set(ctx, &plan); "+
						"any Optional+Computed attribute the API does not return would stay unknown "+
						"and fail the apply", file, fn.Name.Name)
			}
		}
	}

	// Guards the walk itself: a rename that stopped matching would otherwise
	// leave this test passing over nothing.
	assert.Greater(t, checked, 100, "expected to find the generated Create/Update methods")
}

func setsStateFromPlan(body *ast.BlockStmt) bool {
	return containsCall(body, func(sel *ast.SelectorExpr, call *ast.CallExpr) bool {
		if sel.Sel.Name != "Set" || len(call.Args) != 2 {
			return false
		}
		unary, ok := call.Args[1].(*ast.UnaryExpr)
		if !ok {
			return false
		}
		ident, ok := unary.X.(*ast.Ident)
		return ok && ident.Name == "plan"
	})
}

func callsResolveUnknowns(body *ast.BlockStmt) bool {
	return containsCall(body, func(sel *ast.SelectorExpr, _ *ast.CallExpr) bool {
		pkg, ok := sel.X.(*ast.Ident)
		return ok && pkg.Name == "flex" && sel.Sel.Name == "ResolveUnknowns"
	})
}

func containsCall(body *ast.BlockStmt, match func(*ast.SelectorExpr, *ast.CallExpr) bool) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && match(sel, call) {
			found = true
			return false
		}
		return true
	})
	return found
}
