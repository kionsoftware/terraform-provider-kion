package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// run must read the manifest before it answers --list-types, and must answer it
// only once.
//
// --list-types lists what the manifest holds, and --manifest says which manifest
// to read. An earlier arrangement answered --list-types from the embedded
// manifest and returned before the --manifest-aware load further down, so the
// second branch was unreachable and passing both flags silently listed the
// embedded set instead of the file's. Nothing failed; the output was just wrong,
// which is the hardest kind of wrong to notice in a tool whose whole job is to
// tell you what it found.
//
// This is structural rather than behavioral because listTypes prints straight
// to stdout, and the property worth protecting is the ordering, not the text.
func TestRunLoadsManifestBeforeListTypes(t *testing.T) {
	t.Parallel()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", nil, 0)
	require.NoError(t, err)

	var runFn *ast.FuncDecl
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "run" && fn.Recv == nil {
			runFn = fn
			break
		}
	}
	require.NotNil(t, runFn, "run not found in main.go")

	var loadPos, listPos token.Pos
	listCount := 0

	ast.Inspect(runFn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			// kimport.LoadManifest(data)
			if sel, ok := node.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "LoadManifest" {
				if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "kimport" && !loadPos.IsValid() {
					loadPos = node.Pos()
				}
			}
		case *ast.IfStmt:
			if ident, ok := node.Cond.(*ast.Ident); ok && ident.Name == "flagListTypes" {
				listCount++
				if !listPos.IsValid() {
					listPos = node.Pos()
				}
			}
		}
		return true
	})

	require.True(t, listPos.IsValid(), "no `if flagListTypes` branch in run")
	assert.Equal(t, 1, listCount,
		"run has %d `if flagListTypes` branches; the second is unreachable, and the "+
			"two disagreed about whether --manifest applies", listCount)

	require.True(t, loadPos.IsValid(), "run never calls kimport.LoadManifest")
	assert.Less(t, int(loadPos), int(listPos),
		"run answers --list-types before loading the manifest, so --manifest is ignored")
}
