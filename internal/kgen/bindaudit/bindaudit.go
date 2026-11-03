// Package bindaudit reports schema attributes a practitioner can set that the
// resource's Create or Update never reads.
//
// The gap is structural, and it is the one failure mode a compiler cannot see.
// A request body is assembled by matching SDK body fields to model fields by
// json/tfsdk name (bodyBinds and resolveNested in internal/kgen/crud); a body
// field with no model counterpart is skipped, and a model field no body field
// names is simply never visited. Neither path says anything. The result is a
// struct literal that compiles, type-checks, and lies: the provider accepts
// `default_value_string = "x"` and sends a body that does not carry it, or --
// the limit case -- assembles `&PermissionSchemeUpdateRequest{}` and sends `{}`.
//
// Checking the shape the generator BOUND would only cover resources the
// generator derives. This package reads the generated output instead, so the
// bespoke archetypes (cloud_account, cv_override, singleton) and any
// hand-written body are audited by the same rule. It needs no OpenAPI spec,
// which is gitignored -- the same reason fieldaudit reads output rather than
// spec.
//
// The invariant, per resource:
//
//	every Optional or Required top-level attribute is read inside Create, and
//	(unless it forces replacement) inside Update.
//
// "Read" means the model field appears as a value -- an argument, an operand,
// the right-hand side of an assignment -- directly in the method or in a
// package-local helper the method hands the model to. Assigning TO a field is
// not a read: that is what a flatten does, and a resource that only ever writes
// an attribute back from the server is exactly the resource that dropped the
// user's input on the way out.
package bindaudit

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Finding is one settable attribute an operation never reads.
type Finding struct {
	Package string // service package, e.g. "custom_variable"
	Attr    string // schema attribute name, e.g. "default_value_string"
	Op      string // "create" | "update"
	Body    string // the SDK request type(s) the op builds, e.g. "GlobalCustomVariableCreate"
}

// Key identifies a finding for baseline lookup.
func (f Finding) Key() string { return f.Package + "." + f.Attr + ":" + f.Op }

func (f Finding) String() string { return f.Key() }

// attrInfo is what the audit needs to know about one schema attribute.
type attrInfo struct {
	settable        bool
	requiresReplace bool
}

// Run audits every resource under repoRoot/internal/service. Findings come back
// sorted so a caller can diff them against a baseline.
func Run(repoRoot string) ([]Finding, error) {
	svcDir := filepath.Join(repoRoot, "internal", "service")
	entries, err := os.ReadDir(svcDir)
	if err != nil {
		return nil, err
	}

	var findings []Finding
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pkg := e.Name()
		got, err := auditPackage(filepath.Join(svcDir, pkg), pkg)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", pkg, err)
		}
		findings = append(findings, got...)
	}

	sort.Slice(findings, func(i, j int) bool { return findings[i].Key() < findings[j].Key() })
	return findings, nil
}

// pkgFiles parses every non-test Go file in a service package.
func pkgFiles(dir string) (*token.FileSet, []*ast.File, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, e := range ents {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing %s: %w", name, err)
		}
		files = append(files, f)
	}
	return fset, files, nil
}

// auditPackage audits one service package.
func auditPackage(dir, pkg string) ([]Finding, error) {
	_, files, err := pkgFiles(dir)
	if err != nil {
		return nil, err
	}

	attrs := resourceSchemaAttrs(files)
	if len(attrs) == 0 {
		return nil, nil // data-source-only package, or no resource schema
	}

	funcs := indexFuncs(files)
	create := findMethod(files, "Create", "CreateRequest")
	update := findMethod(files, "Update", "UpdateRequest")

	// The model is whichever struct Create declares, not whichever struct best
	// matches the schema: a package holds the data source model too, and it
	// carries the same attribute names.
	models := modelStructs(files)
	model, ok := pickModel(models, declaredModelName(create), declaredModelName(update))
	if !ok {
		return nil, nil // no resource model: data-source-only or hand-rolled beyond this audit
	}

	var out []Finding
	check := func(fn *ast.FuncDecl, op string, skipReplace bool) {
		if fn == nil {
			return
		}
		read := readsIn(fn, model, funcs, map[string]bool{})
		body := requestTypes(fn)
		var attrNames []string
		for name := range attrs {
			attrNames = append(attrNames, name)
		}
		sort.Strings(attrNames)
		for _, name := range attrNames {
			info := attrs[name]
			if !info.settable || (skipReplace && info.requiresReplace) {
				continue
			}
			goName, ok := model.byTF[name]
			if !ok {
				continue // attribute with no model field: a schema/model mismatch, not this audit's business
			}
			if !read[goName] {
				out = append(out, Finding{Package: pkg, Attr: name, Op: op, Body: body})
			}
		}
	}
	check(create, "create", false)
	if callsAPI(update, funcs, map[string]bool{}) {
		// An Update that issues no request does not drop anything silently: it
		// either refuses in-place change with a diagnostic, or the resource has
		// no update endpoint at all. Both are loud.
		check(update, "update", true)
	}
	return out, nil
}

// requestTypes names the SDK request structs a method builds, so a baseline
// entry says which body has no room for the attribute. Opt-wrappers are
// unwrapped to the payload type they carry.
func requestTypes(fn *ast.FuncDecl) string {
	if fn == nil || fn.Body == nil {
		return ""
	}
	seen := map[string]bool{}
	var names []string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		sel, ok := lit.Type.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || id.Name != "generated" {
			return true
		}
		name := sel.Sel.Name
		if strings.HasSuffix(name, "Params") || strings.HasPrefix(name, "Opt") {
			return true // the params struct, or a wrapper whose payload is visited too
		}
		if !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
		return true
	})
	return strings.Join(names, ", ")
}

// callsAPI reports whether a method reaches the API, directly or through a
// package-local helper: it mentions the SDK client, the raw HTTP escape hatch,
// or a generated request/params type.
func callsAPI(fn *ast.FuncDecl, funcs funcIndex, seen map[string]bool) bool {
	if fn == nil || fn.Body == nil || seen[fn.Name.Name] {
		return false
	}
	seen[fn.Name.Name] = true
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && (id.Name == "generated" || id.Name == "conn") {
			found = true
			return false
		}
		switch sel.Sel.Name {
		case "Client", "RawPut", "RawPost", "RawPatch", "RawDelete", "RawGet":
			found = true
			return false
		}
		return true
	})
	if found {
		return true
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if target, ok := funcs[calleeName(call)]; ok && callsAPI(target, funcs, seen) {
			found = true
			return false
		}
		return true
	})
	return found
}

// resourceSchemaAttrs collects the top-level attributes of every
// `func …ResourceSchema(ctx) schema.Schema` in the package, with whether each
// is settable and whether it forces replacement.
//
// Only the top level is walked: a nested attribute's sub-attributes travel with
// their parent, so binding the parent binds them.
func resourceSchemaAttrs(files []*ast.File) map[string]attrInfo {
	out := map[string]attrInfo{}
	for _, f := range files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !strings.HasSuffix(fn.Name.Name, "ResourceSchema") {
				continue
			}
			lit := returnedSchemaLit(fn)
			if lit == nil {
				continue
			}
			for _, elt := range lit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok || key.Name != "Attributes" {
					continue
				}
				attrMap, ok := kv.Value.(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, ae := range attrMap.Elts {
					akv, ok := ae.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					name, ok := stringLit(akv.Key)
					if !ok {
						continue
					}
					val, ok := akv.Value.(*ast.CompositeLit)
					if !ok {
						continue
					}
					out[name] = attrInfo{
						settable:        boolField(val, "Optional") || boolField(val, "Required"),
						requiresReplace: mentions(val, "RequiresReplace"),
					}
				}
			}
		}
	}
	return out
}

// returnedSchemaLit finds the `schema.Schema{…}` composite literal a schema
// function returns.
func returnedSchemaLit(fn *ast.FuncDecl) *ast.CompositeLit {
	var found *ast.CompositeLit
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if found != nil {
			return false
		}
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		sel, ok := lit.Type.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Schema" {
			return true
		}
		found = lit
		return false
	})
	return found
}

// boolField reports whether a composite literal sets `name: true`.
func boolField(lit *ast.CompositeLit, name string) bool {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != name {
			continue
		}
		id, ok := kv.Value.(*ast.Ident)
		return ok && id.Name == "true"
	}
	return false
}

// mentions reports whether an identifier by that name appears anywhere in the
// subtree -- used to spot a RequiresReplace plan modifier without caring which
// typed variant of it was used.
func mentions(n ast.Node, name string) bool {
	found := false
	ast.Inspect(n, func(node ast.Node) bool {
		if found {
			return false
		}
		if id, ok := node.(*ast.Ident); ok && id.Name == name {
			found = true
		}
		return true
	})
	return found
}

// modelInfo is a struct whose fields carry tfsdk tags: a Terraform model.
type modelInfo struct {
	name string
	byTF map[string]string // tfsdk name -> Go field name
}

// modelStructs indexes every tfsdk-tagged struct in the package.
func modelStructs(files []*ast.File) []modelInfo {
	var out []modelInfo
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok || st.Fields == nil {
				return true
			}
			byTF := map[string]string{}
			for _, fld := range st.Fields.List {
				if fld.Tag == nil || len(fld.Names) == 0 {
					continue
				}
				tf := tfsdkName(fld.Tag.Value)
				if tf == "" {
					continue
				}
				byTF[tf] = fld.Names[0].Name
			}
			if len(byTF) > 0 {
				out = append(out, modelInfo{name: ts.Name.Name, byTF: byTF})
			}
			return true
		})
	}
	return out
}

// pickModel returns the named struct, trying each name in order.
func pickModel(models []modelInfo, names ...string) (modelInfo, bool) {
	for _, want := range names {
		if want == "" {
			continue
		}
		for _, m := range models {
			if m.name == want {
				return m, true
			}
		}
	}
	return modelInfo{}, false
}

// declaredModelName returns the type of the first tfsdk-model-shaped local a
// CRUD method declares (`var plan FooModel`). It is how the audit tells the
// resource model from the data source's.
func declaredModelName(fn *ast.FuncDecl) string {
	if fn == nil || fn.Body == nil {
		return ""
	}
	name := ""
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if name != "" {
			return false
		}
		vs, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		if id, ok := vs.Type.(*ast.Ident); ok {
			name = id.Name
			return false
		}
		return true
	})
	return name
}

func tfsdkName(rawTag string) string {
	tag := strings.Trim(rawTag, "`")
	_, rest, found := strings.Cut(tag, `tfsdk:"`)
	if !found {
		return ""
	}
	name, _, _ := strings.Cut(rest, `"`)
	return name
}

// funcIndex maps a package-local function/method name to its declaration.
type funcIndex map[string]*ast.FuncDecl

func indexFuncs(files []*ast.File) funcIndex {
	out := funcIndex{}
	for _, f := range files {
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				out[fn.Name.Name] = fn
			}
		}
	}
	return out
}

// findMethod returns the resource method named `name` whose second parameter is
// the framework request type (so the data source's Read is not mistaken for the
// resource's).
func findMethod(files []*ast.File, name, reqType string) *ast.FuncDecl {
	for _, f := range files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || fn.Name.Name != name || fn.Type.Params == nil {
				continue
			}
			for _, p := range fn.Type.Params.List {
				if sel, ok := p.Type.(*ast.SelectorExpr); ok && sel.Sel.Name == reqType {
					return fn
				}
			}
		}
	}
	return nil
}

// readsIn returns the model fields a function reads, following package-local
// calls it hands the model to. depth guards against recursion.
func readsIn(fn *ast.FuncDecl, model modelInfo, funcs funcIndex, seen map[string]bool) map[string]bool {
	out := map[string]bool{}
	if fn == nil || fn.Body == nil || seen[fn.Name.Name] {
		return out
	}
	seen[fn.Name.Name] = true

	// Local variables of the model type, plus any parameter of it.
	vars := modelVars(fn, model.name)

	// Fields on the left of an assignment are written, not read. Collect them
	// first so the read walk can exclude exactly those selector nodes.
	written := map[ast.Node]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range as.Lhs {
			written[lhs] = true
		}
		return true
	})

	goNames := map[string]bool{}
	for _, g := range model.byTF {
		goNames[g] = true
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok || written[n] || !goNames[sel.Sel.Name] {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && vars[id.Name] {
			out[sel.Sel.Name] = true
		}
		return true
	})

	// Follow package-local calls that receive the model.
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !passesModel(call, vars) {
			return true
		}
		callee := calleeName(call)
		if callee == "" {
			return true
		}
		target, ok := funcs[callee]
		if !ok {
			return true
		}
		for k := range readsIn(target, model, funcs, seen) {
			out[k] = true
		}
		return true
	})

	return out
}

// modelVars names the identifiers in a function that hold the model type,
// whether declared (`var plan FooModel`), assigned, or received as a parameter.
func modelVars(fn *ast.FuncDecl, modelName string) map[string]bool {
	out := map[string]bool{}
	isModel := func(e ast.Expr) bool {
		switch t := e.(type) {
		case *ast.Ident:
			return t.Name == modelName
		case *ast.StarExpr:
			id, ok := t.X.(*ast.Ident)
			return ok && id.Name == modelName
		}
		return false
	}
	if fn.Type.Params != nil {
		for _, p := range fn.Type.Params.List {
			if !isModel(p.Type) {
				continue
			}
			for _, n := range p.Names {
				out[n.Name] = true
			}
		}
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		vs, ok := n.(*ast.ValueSpec)
		if !ok || !isModel(vs.Type) {
			return true
		}
		for _, n := range vs.Names {
			out[n.Name] = true
		}
		return true
	})
	return out
}

// passesModel reports whether a call hands one of the model variables over.
func passesModel(call *ast.CallExpr, vars map[string]bool) bool {
	for _, arg := range call.Args {
		e := arg
		if u, ok := e.(*ast.UnaryExpr); ok {
			e = u.X
		}
		if id, ok := e.(*ast.Ident); ok && vars[id.Name] {
			return true
		}
	}
	return false
}

// calleeName returns the called function's bare name for a package-local call
// (`helper(...)` or `r.helper(...)`), and "" for anything qualified by another
// package.
func calleeName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		if id, ok := fn.X.(*ast.Ident); ok && (id.Name == "r" || id.Name == "d") {
			return fn.Sel.Name
		}
	}
	return ""
}

func stringLit(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	return strings.Trim(lit.Value, `"`), true
}
