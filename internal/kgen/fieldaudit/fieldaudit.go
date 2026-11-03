// Package fieldaudit reports scalar fields an API response carries that the
// provider exposes nowhere: not as a resource attribute, and not on the data
// source's filter surface.
//
// The gap is structural rather than a per-resource oversight. buildListDSData
// (internal/kgen/crud/datasource.go) walks the RESOURCE's attributes -- which
// are generated from the create request body -- and looks each one up in the
// response payload. A field the server owns and the client cannot set is
// absent from that request body, so the loop never visits it, and it reaches
// neither the schema nor <pkg>ToRow. Provenance flags are the systematic
// casualty: ct_managed, system_managed_policy and their siblings are exactly
// the fields that say "Kion shipped this, a customer did not", and exactly the
// fields a create body has no reason to carry.
//
// This package reads the generated output rather than the OpenAPI spec, on
// purpose: spec/openapi3.json is gitignored, so a spec-driven check could not
// run in CI. The SDK module is an ordinary dependency and is always present.
package fieldaudit

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"terraform-provider-kion/internal/kgen/versions"
)

// SDKModulePath is the import path the provider resolves the SDK under. The
// go.mod `replace` may point it elsewhere; `go list -m` reports the directory
// that actually backs it, which is what this package reads.
const SDKModulePath = "github.com/kionsoftware/kion-sdk-go"

// SchemaFile is the SDK's generated model file, relative to the module root.
func SchemaFile() string { return "generated/" + versions.CurrentSDK() + "/oas_schemas_gen.go" }

// Finding is one scalar response field that the provider never exposes.
type Finding struct {
	Package string // provider service package, e.g. "azure_policy"
	Elem    string // SDK element struct the data source lists, e.g. "AzurePolicyAugmented"
	Field   string // the response field's JSON name, e.g. "ct_managed"
	GoType  string // its SDK Go type, e.g. "OptNilBool"
}

// Key identifies a finding independently of the SDK type, so a baseline entry
// survives a type change that does not change whether the field is exposed.
func (f Finding) Key() string { return f.Package + "." + f.Field }

type sdkField struct {
	JSON   string
	GoName string
	GoType string
}

// SDKDir returns the directory backing the SDK module.
func SDKDir() (string, error) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", SDKModulePath).Output()
	if err != nil {
		return "", fmt.Errorf("locating %s (is the module downloaded?): %w", SDKModulePath, err)
	}
	dir := strings.TrimSpace(string(out))
	if dir == "" {
		return "", fmt.Errorf("go list reported no directory for %s", SDKModulePath)
	}
	return dir, nil
}

// Run audits every generated data source under repoRoot/internal/service.
// Findings come back sorted, so a caller can diff them against a baseline.
func Run(repoRoot, sdkDir string) ([]Finding, error) {
	structs, err := parseSDKStructs(filepath.Join(sdkDir, SchemaFile()))
	if err != nil {
		return nil, err
	}

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
		dsPath := filepath.Join(svcDir, pkg, pkg+"_data_source.go")
		if _, err := os.Stat(dsPath); err != nil {
			continue // id-only or bespoke data source: no filter surface to audit
		}

		elem, rowKeys, rowRead, reached, err := parseToRow(dsPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", pkg, err)
		}
		if elem == "" {
			continue // no ToRow: nothing filterable to compare against
		}
		fields, ok := structs[elem]
		if !ok {
			continue // element type is not an SDK struct (bespoke shape)
		}

		exposed := map[string]bool{}
		for k := range rowKeys {
			exposed[k] = true
		}
		attrs, err := parseSchemaAttrs(filepath.Join(svcDir, pkg, pkg+"_schema_gen.go"))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", pkg, err)
		}
		for k := range attrs {
			exposed[k] = true
		}

		// Candidates are the element struct's own fields plus those of any
		// single-nested struct ToRow actually reaches through -- the augmented
		// wrappers (AzurePolicyAugmented.AzurePolicy) hold the real payload.
		candidates := append([]sdkField(nil), fields...)
		for _, f := range fields {
			inner := baseTypeName(f.GoType)
			if _, isStruct := structs[inner]; isStruct && reached[f.GoName] {
				candidates = append(candidates, structs[inner]...)
			}
		}

		seen := map[string]bool{}
		for _, f := range candidates {
			// rowRead catches a field published under a different name than its
			// JSON one, which comparing names alone reports as a false gap.
			if seen[f.JSON] || exposed[f.JSON] || rowRead[f.GoName] {
				continue
			}
			seen[f.JSON] = true
			if !isScalar(f.GoType, structs) {
				continue // structs and slices were never filter-surface candidates
			}
			findings = append(findings, Finding{
				Package: pkg, Elem: elem, Field: f.JSON, GoType: f.GoType,
			})
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Package != findings[j].Package {
			return findings[i].Package < findings[j].Package
		}
		return findings[i].Field < findings[j].Field
	})
	return findings, nil
}

// parseSDKStructs indexes every struct in the SDK's generated model file by
// name, keeping declaration order so output is stable.
func parseSDKStructs(path string) (map[string][]sdkField, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parsing SDK models: %w", err)
	}

	out := map[string][]sdkField{}
	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			return true
		}
		var fields []sdkField
		for _, fld := range st.Fields.List {
			if fld.Tag == nil || len(fld.Names) == 0 {
				continue
			}
			tag, err := strconv.Unquote(fld.Tag.Value)
			if err != nil {
				continue
			}
			name := jsonName(tag)
			if name == "" || name == "-" {
				continue
			}
			fields = append(fields, sdkField{
				JSON: name, GoName: fld.Names[0].Name, GoType: typeString(fld.Type),
			})
		}
		if len(fields) > 0 {
			out[ts.Name.Name] = fields
		}
		return true
	})
	return out, nil
}

// parseToRow pulls the element type, the map keys <pkg>ToRow emits, the SDK
// fields those keys read, and the element fields its body reaches through, out
// of a generated data source.
//
// read is what makes a renamed exposure count as exposed. A data source is free
// to publish a field under a different name -- the bespoke cloud_account one
// emits AccountEmail as `email` -- and comparing JSON names alone reports that
// as unexposed when it is merely spelled differently. Recording the SDK field
// each emitted key actually reads answers the real question: does anything in
// this row come from that field?
func parseToRow(path string) (elem string, keys, read, reached map[string]bool, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return "", nil, nil, nil, err
	}

	keys, read, reached = map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !strings.HasSuffix(fn.Name.Name, "ToRow") || fn.Recv != nil {
			continue
		}
		if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
			continue
		}
		sel, ok := fn.Type.Params.List[0].Type.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		elem = sel.Sel.Name
		param := ""
		if names := fn.Type.Params.List[0].Names; len(names) > 0 {
			param = names[0].Name
		}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.KeyValueExpr:
				if lit, ok := v.Key.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if s, err := strconv.Unquote(lit.Value); err == nil {
						keys[s] = true
					}
				}
				collectFields(v.Value, param, read)
			case *ast.AssignStmt:
				// row["x"] = <expr>
				for _, lhs := range v.Lhs {
					idx, ok := lhs.(*ast.IndexExpr)
					if !ok {
						continue
					}
					if lit, ok := idx.Index.(*ast.BasicLit); ok && lit.Kind == token.STRING {
						if s, err := strconv.Unquote(lit.Value); err == nil {
							keys[s] = true
						}
					}
					for _, rhs := range v.Rhs {
						collectFields(rhs, param, read)
					}
				}
			case *ast.SelectorExpr:
				// <param>.<Field>... -- records which nested field is reached into
				if inner, ok := v.X.(*ast.SelectorExpr); ok {
					if id, ok := inner.X.(*ast.Ident); ok && param != "" && id.Name == param {
						reached[inner.Sel.Name] = true
					}
				}
			}
			return true
		})
		break
	}
	return elem, keys, read, reached, nil
}

// wrapperAccessors are the ogen Opt fields and methods that sit between an
// access path and the value; they name no SDK field of their own.
var wrapperAccessors = map[string]bool{"Value": true, "Set": true, "Null": true, "Or": true, "Get": true}

// collectFields records every SDK field name an expression reads off param,
// including through an Opt wrapper (lbl.Ami.Value.Name yields Ami and Name).
func collectFields(expr ast.Expr, param string, out map[string]bool) {
	if param == "" {
		return
	}
	ast.Inspect(expr, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok || wrapperAccessors[sel.Sel.Name] || !rootsAt(sel.X, param) {
			return true
		}
		out[sel.Sel.Name] = true
		return true
	})
}

// rootsAt reports whether a selector chain bottoms out at the named identifier.
func rootsAt(e ast.Expr, param string) bool {
	for {
		switch v := e.(type) {
		case *ast.Ident:
			return v.Name == param
		case *ast.SelectorExpr:
			e = v.X
		case *ast.CallExpr:
			e = v.Fun
		case *ast.IndexExpr:
			e = v.X
		case *ast.ParenExpr:
			e = v.X
		default:
			return false
		}
	}
}

// parseSchemaAttrs collects the attribute names a generated schema declares.
// A missing file is not an error: not every service package has one.
func parseSchemaAttrs(path string) (map[string]bool, error) {
	out := map[string]bool{}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	ast.Inspect(f, func(n ast.Node) bool {
		kv, ok := n.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		lit, ok := kv.Key.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}
		out[s] = true
		return true
	})
	return out, nil
}

func jsonName(tag string) string {
	_, rest, found := strings.Cut(tag, `json:"`)
	if !found {
		return ""
	}
	rest, _, _ = strings.Cut(rest, `"`)
	rest, _, _ = strings.Cut(rest, ",")
	return rest
}

func typeString(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return "*" + typeString(v.X)
	case *ast.ArrayType:
		return "[]" + typeString(v.Elt)
	case *ast.SelectorExpr:
		return typeString(v.X) + "." + v.Sel.Name
	case *ast.MapType:
		return "map"
	default:
		return "?"
	}
}

// optPrefixes are the ogen wrappers, longest first so OptNilPointer is
// stripped before OptNil and Opt.
var optPrefixes = []string{"OptNilPointer", "OptPointer", "OptNil", "Opt", "Nil"}

func baseTypeName(t string) string {
	t = strings.TrimPrefix(t, "*")
	if strings.HasPrefix(t, "[]") {
		return ""
	}
	for _, p := range optPrefixes {
		if strings.HasPrefix(t, p) && len(t) > len(p) {
			return t[len(p):]
		}
	}
	return t
}

// isScalar reports whether a field could ever have been a filter key. Slices
// and nested structs could not: filter.Match compares scalars.
func isScalar(goType string, structs map[string][]sdkField) bool {
	if strings.HasPrefix(strings.TrimPrefix(goType, "*"), "[]") {
		return false
	}
	base := baseTypeName(goType)
	if base == "" {
		return false
	}
	if _, isStruct := structs[base]; isStruct {
		return false
	}
	if base == "jx.Raw" {
		return false // an arbitrary JSON document, not something filter.Match compares
	}
	if strings.Contains(base, ".") { // time.Time and friends
		return true
	}
	// Lowercased on purpose: stripping the ogen wrapper off OptString leaves
	// "String", off OptNilUint64 leaves "Uint64". Comparing those against Go's
	// lowercase predeclared names would classify every optional scalar in the
	// SDK as a composite, which is to say it would find nothing at all.
	switch strings.ToLower(base) {
	case "string", "bool", "int", "int32", "int64",
		"uint", "uint32", "uint64", "float32", "float64":
		return true
	}
	return false
}
