package sweep_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const servicePrefix = "terraform-provider-kion/internal/service/"

// TestSweeperRegistration guards the two ways `make sweep` can silently clean
// nothing: a sweeper that is registered but never calls the API, and a sweeper
// that does call it but is never linked into this package.
//
// Both failed before: 29 generated sweep.go files registered a body that
// returned nil, and this entrypoint's import list was hand-maintained and
// missed packages that had a real one.
func TestSweeperRegistration(t *testing.T) {
	t.Parallel()

	registered, noop := scanServiceSweepers(t)
	if len(noop) > 0 {
		t.Errorf("these sweep.go files register a sweeper that never calls the API.\n"+
			"A no-op reporting success is worse than no sweeper: regenerate with `make crud-force`,\n"+
			"which emits a file registering nothing when a resource cannot be enumerated.\n  %s",
			strings.Join(noop, "\n  "))
	}

	imported := entrypointImports(t)
	for _, pkg := range diff(registered, imported) {
		t.Errorf("internal/service/%s registers a sweeper that never runs: add a blank import to internal/sweep/sweep_test.go", pkg)
	}
	for _, pkg := range diff(imported, registered) {
		t.Errorf("internal/sweep/sweep_test.go imports internal/service/%s, which registers no sweeper: drop the import", pkg)
	}
}

// scanServiceSweepers reads every internal/service/*/sweep.go, returning the
// packages that register a sweeper and, separately, any whose registered
// function never reaches the API.
func scanServiceSweepers(t *testing.T) (registered, noop []string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("..", "service", "*", "sweep.go"))
	if err != nil {
		t.Fatalf("globbing sweep files: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no internal/service/*/sweep.go files found")
	}
	for _, p := range paths {
		f, err := parser.ParseFile(token.NewFileSet(), p, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", p, err)
		}
		names := registeredSweeperFuncs(f)
		if len(names) == 0 {
			continue
		}
		registered = append(registered, filepath.Base(filepath.Dir(p)))
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Body == nil || !names[fn.Name.Name] {
				continue
			}
			if !callsAPI(fn.Body) {
				noop = append(noop, p+" ("+fn.Name.Name+")")
			}
		}
	}
	sort.Strings(registered)
	sort.Strings(noop)
	return registered, noop
}

// registeredSweeperFuncs collects the function named by each
// resource.AddTestSweepers(..., &resource.Sweeper{F: <ident>}) call in f.
func registeredSweeperFuncs(f *ast.File) map[string]bool {
	names := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		kv, ok := n.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "F" {
			return true
		}
		if id, ok := kv.Value.(*ast.Ident); ok {
			names[id.Name] = true
		}
		return true
	})
	return names
}

// callsAPI reports whether the body reaches a conn.Client.<Op> method. A body
// that fetches the shared client and returns nil does not.
func callsAPI(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && id.Name == "conn" && sel.Sel.Name == "Client" {
			found = true
		}
		return true
	})
	return found
}

// entrypointImports lists the service packages blank-imported by sweep_test.go.
func entrypointImports(t *testing.T) []string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "sweep_test.go", nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing sweep_test.go: %v", err)
	}
	var out []string
	for _, imp := range f.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if pkg, ok := strings.CutPrefix(path, servicePrefix); ok {
			out = append(out, pkg)
		}
	}
	sort.Strings(out)
	return out
}

// diff returns the elements of a that are absent from b.
func diff(a, b []string) []string {
	have := make(map[string]bool, len(b))
	for _, s := range b {
		have[s] = true
	}
	var out []string
	for _, s := range a {
		if !have[s] {
			out = append(out, s)
		}
	}
	return out
}
