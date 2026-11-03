package kalign

import (
	"fmt"
	"go/ast"
	"strings"
)

// tagValue extracts the first (pre-comma) value of a struct tag key,
// e.g. tagValue(`json:"id,omitempty"`, "json") == "id".
func tagValue(tag *ast.BasicLit, key string) string {
	if tag == nil {
		return ""
	}
	raw := strings.Trim(tag.Value, "`")
	_, after, ok := strings.Cut(raw, key+`:"`)
	if !ok {
		return ""
	}
	val, _, ok := strings.Cut(after, `"`)
	if !ok {
		return ""
	}
	if v, _, found := strings.Cut(val, ","); found {
		val = v
	}
	return val
}

// typeString renders a type expression as source text, e.g. "types.String",
// "OptNilUint64", "*ScopeCriteriaRecordCriteria", "[]string".
func typeString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	default:
		return fmt.Sprintf("%T", e)
	}
}

// goExported turns a snake_case service name into CamelCase, e.g.
// "ou_note" -> "OuNote".
func goExported(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}
