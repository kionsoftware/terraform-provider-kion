package config

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Source abstracts reading the OpenAPI spec and the service packages so the
// derive/check/gen logic is unit-testable against a mock.
type Source interface {
	// Operations returns operationId -> {Method, Path} from the OpenAPI spec.
	Operations(specPath string) (map[string]Op, error)
	// ServiceOps returns, per service package under root, the SDK op names
	// invoked in each CRUD method of the resource and the Read of the data
	// source. Empty slices mean a stub (no SDK calls yet).
	ServiceOps(root string) ([]ServiceOps, error)
}

// Op is a resolved API operation.
type Op struct {
	Method string // GET, POST, PATCH, PUT, DELETE
	Path   string // e.g. /v3/label/{id}
}

// ServiceOps is the raw list of SDK op names a service's methods call.
type ServiceOps struct {
	Name                         string
	Create, Read, Update, Delete []string
	DataSourceRead               []string
	// Raw* are set when an op goes over raw HTTP rather than the SDK, which the
	// op-name lists cannot express: a raw call is
	// r.Meta().RawPost(ctx, "/v3/...", body), and "RawPost" is not an
	// operationId, so the op simply vanished and the derived config reported the
	// resource INCOMPLETE or dropped a verb. Both the raw_create archetype and
	// the fully raw_http one need this.
	RawCreate, RawRead, RawUpdate, RawDelete *Op
}

type fileSource struct{}

// NewFileSource returns the production Source backed by the filesystem.
func NewFileSource() Source { return fileSource{} }

func (fileSource) Operations(specPath string) (map[string]Op, error) {
	raw, err := os.ReadFile(specPath)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	ops := make(map[string]Op)
	for path, methods := range doc.Paths {
		for method, op := range methods {
			switch strings.ToLower(method) {
			case "get", "post", "put", "patch", "delete":
				if op.OperationID != "" {
					ops[op.OperationID] = Op{Method: strings.ToUpper(method), Path: path}
				}
			}
		}
	}
	return ops, nil
}

func (fileSource) ServiceOps(root string) ([]ServiceOps, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []ServiceOps
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		so := ServiceOps{Name: name}
		if f := parseGo(filepath.Join(root, name, name+".go")); f != nil {
			so.Create = callsInMethod(f, "Create")
			so.Read = callsInMethod(f, "Read")
			so.Update = callsInMethod(f, "Update")
			so.Delete = callsInMethod(f, "Delete")
			so.RawCreate = rawCallInMethod(f, "Create")
			so.RawRead = rawCallInMethod(f, "Read")
			so.RawUpdate = rawCallInMethod(f, "Update")
			so.RawDelete = rawCallInMethod(f, "Delete")
		}
		if f := parseGo(filepath.Join(root, name, name+"_data_source.go")); f != nil {
			so.DataSourceRead = callsInMethod(f, "Read")
		}
		out = append(out, so)
	}
	return out, nil
}

func parseGo(path string) *ast.File {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil
	}
	return f
}

// callsInMethod returns the selector names of every method call inside the named
// method (e.g. "conn.PostLabel(...)" -> "PostLabel"). Callers filter these
// against the operationId index to keep only real SDK ops.
func callsInMethod(f *ast.File, method string) []string {
	var names []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name.Name != method || fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				names = append(names, sel.Sel.Name)
			}
			return true
		})
	}
	return names
}

// rawVerbMethod maps a conns raw helper to its HTTP method.
var rawVerbMethod = map[string]string{
	"RawPost": "POST", "RawPatch": "PATCH", "RawPut": "PUT",
	"RawDelete": "DELETE", "RawGet": "GET",
}

// rawCallInMethod finds a raw HTTP call in the named method and reconstructs
// the op it performs, so a resource whose create bypasses the SDK is still
// described by the derived config rather than reported as missing one.
func rawCallInMethod(f *ast.File, method string) *Op {
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name.Name != method || fn.Body == nil {
			continue
		}
		// A write wins over a GET: a read-modify-write update reads first.
		var found, get *Op
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if found != nil {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			httpMethod, ok := rawVerbMethod[sel.Sel.Name]
			if !ok {
				return true
			}
			// The path is the first path-shaped string literal in the call.
			// It is not always a direct argument: a by-id op wraps it, as
			// RawPatch(ctx, strings.Replace("/beta/dashboard/{id}", …), body),
			// so this searches the whole expression rather than just Args.
			for _, arg := range call.Args {
				if p, ok := pathLiteral(arg); ok {
					op := &Op{Method: httpMethod, Path: specRoute(p)}
					if httpMethod != "GET" {
						found = op
					} else if get == nil {
						get = op
					}
					return false
				}
			}
			return true
		})
		if found != nil {
			return found
		}
		if get != nil {
			return get
		}
	}
	return nil
}

// specRoute rewrites a single query-string discriminator into the route fixspec
// gives the spec (/x?k=v becomes /x/__qs__/k/v), the name the config and the
// version gates use. Any other query is left as written.
func specRoute(p string) string {
	base, query, ok := strings.Cut(p, "?")
	if !ok {
		return p
	}
	k, v, ok := strings.Cut(query, "=")
	if !ok || k == "" || v == "" || strings.ContainsAny(v, "&=") {
		return p
	}
	return base + "/__qs__/" + k + "/" + v
}

// pathLiteral finds the first string literal in an expression that looks like an
// API path.
func pathLiteral(e ast.Expr) (string, bool) {
	var out string
	ast.Inspect(e, func(n ast.Node) bool {
		if out != "" {
			return false
		}
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}
		p, err := strconv.Unquote(lit.Value)
		if err == nil && strings.HasPrefix(p, "/") {
			out = p
			return false
		}
		return true
	})
	return out, out != ""
}
