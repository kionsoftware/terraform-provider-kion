package crud

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"slices"
	"strings"
)

// NewFileSource returns the production Source that reads SDK and generated
// model files via go/ast.
func NewFileSource() Source { return fileSource{} }

type fileSource struct{}

// ClientMethods parses oas_client_gen.go and returns one ClientMethod per
// (c *Client) method, sorted by name. The HTTP method + path come from the
// ogen doc-comment route line (e.g. "// POST /v3/label"); the body/params/
// result types come from the signature.
func (fileSource) ClientMethods(clientFile string) ([]ClientMethod, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, clientFile, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", clientFile, err)
	}
	var out []ClientMethod
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || recvType(fn.Recv) != "Client" {
			continue
		}
		method, path := routeFromDoc(fn.Doc.Text())
		if method == "" {
			continue // not an operation method
		}
		cm := ClientMethod{
			Name:       fn.Name.Name,
			HTTPMethod: method,
			Path:       path,
			ResultType: firstResultType(fn.Type.Results),
		}
		for _, p := range fn.Type.Params.List {
			for _, name := range p.Names {
				switch name.Name {
				case "request":
					cm.BodyType = exprIdent(p.Type)
					if _, ok := p.Type.(*ast.StarExpr); ok {
						cm.BodyPtr = true
					}
				case "params":
					cm.ParamsType = exprIdent(p.Type)
				}
			}
		}
		out = append(out, cm)
	}
	slices.SortFunc(out, func(a, b ClientMethod) int { return cmp.Compare(a.Name, b.Name) })
	return out, nil
}

// Structs parses a schemas or parameters file into a name->Struct map. Per
// field it records the Go name, the json wire name (from the json tag), the
// type identifier, and whether the type is an Opt*-wrapper (optional).
func (fileSource) Structs(file string) (map[string]Struct, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	out := map[string]Struct{}
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			s := Struct{Name: ts.Name.Name}
			for _, fld := range st.Fields.List {
				if len(fld.Names) == 0 {
					continue // embedded field; entity bodies have none
				}
				typ := exprIdent(fld.Type)
				if at, ok := fld.Type.(*ast.ArrayType); ok {
					typ = "[]" + exprIdent(at.Elt)
				}
				_, ptr := fld.Type.(*ast.StarExpr)
				for _, nm := range fld.Names {
					s.Fields = append(s.Fields, Field{
						GoName:   nm.Name,
						JSONName: jsonName(fld.Tag),
						Type:     typ,
						Optional: strings.HasPrefix(typ, "Opt"),
						Ptr:      ptr,
					})
				}
			}
			out[s.Name] = s
		}
	}
	return out, nil
}

// MarkerImpls maps each unexported empty marker method (e.g. "getLabelRes")
// to the concrete receiver types that implement it. ogen uses these markers
// to model result unions; the read success envelope is one of them.
func (fileSource) MarkerImpls(schemaFile string) (map[string][]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, schemaFile, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", schemaFile, err)
	}
	out := map[string][]string{}
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 {
			continue
		}
		if fn.Name.IsExported() || fn.Type.Params.NumFields() != 0 {
			continue
		}
		out[fn.Name.Name] = append(out[fn.Name.Name], recvType(fn.Recv))
	}
	for k := range out {
		slices.Sort(out[k])
	}
	return out, nil
}

// ModelFields parses a generated <name>_schema_gen.go for the named model
// struct and returns its fields with tfsdk tags and rendered field types.
func (fileSource) ModelFields(schemaGenFile, modelType string) ([]ModelField, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, schemaGenFile, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", schemaGenFile, err)
	}
	var out []ModelField
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok || ts.Name.Name != modelType {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}
		found = true
		for _, fld := range st.Fields.List {
			if len(fld.Names) == 0 {
				continue
			}
			out = append(out, ModelField{
				GoName: fld.Names[0].Name,
				TFSDK:  structTag(fld.Tag, "tfsdk"),
				Type:   renderType(fld.Type),
			})
		}
		return false
	})
	if !found {
		return nil, fmt.Errorf("model %s not found in %s", modelType, schemaGenFile)
	}
	return out, nil
}

// renderType renders a field type expression: types.String, string, *T, ...
func renderType(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprIdent(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + renderType(t.X)
	}
	return ""
}

// structTag returns the value of the named key in a struct-field tag literal.
func structTag(tag *ast.BasicLit, key string) string {
	if tag == nil {
		return ""
	}
	return reflect.StructTag(strings.Trim(tag.Value, "`")).Get(key)
}

// jsonName returns the wire name from a json tag (first comma-segment), or ""
// when absent or "-".
func jsonName(tag *ast.BasicLit) string {
	v := structTag(tag, "json")
	if v == "" {
		return ""
	}
	if i := strings.IndexByte(v, ','); i >= 0 {
		v = v[:i]
	}
	if v == "-" {
		return ""
	}
	return v
}

// routeFromDoc returns ("POST","/v3/label") from the last "METHOD /path" line
// of an ogen client-method doc comment.
func routeFromDoc(doc string) (string, string) {
	lines := strings.Split(strings.TrimSpace(doc), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		fields := strings.Fields(strings.TrimSpace(lines[i]))
		if len(fields) == 2 && isHTTPVerb(fields[0]) && strings.HasPrefix(fields[1], "/") {
			return fields[0], fields[1]
		}
	}
	return "", ""
}

func isHTTPVerb(s string) bool {
	switch s {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	}
	return false
}

func recvType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	return strings.TrimPrefix(exprIdent(recv.List[0].Type), "*")
}

// exprIdent returns the identifier for T, *T, or pkg.T -> "T".
func exprIdent(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return exprIdent(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	}
	return ""
}

func firstResultType(res *ast.FieldList) string {
	if res == nil || len(res.List) == 0 {
		return ""
	}
	return exprIdent(res.List[0].Type)
}
