package kalign

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func parseSrc(t *testing.T, src string) *ast.File {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return f
}

func TestTagValue(t *testing.T) {
	tests := []struct {
		name string
		raw  string // struct tag literal, without backticks
		key  string
		want string
	}{
		{"simple json", `json:"id"`, "json", "id"},
		{"with omitempty", `json:"id,omitempty"`, "json", "id"},
		{"tfsdk", `tfsdk:"create_user_id"`, "tfsdk", "create_user_id"},
		{"multi-tag picks key", `json:"name" tfsdk:"the_name"`, "tfsdk", "the_name"},
		{"missing key", `json:"id"`, "tfsdk", ""},
		{"empty value", `json:""`, "json", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lit := &ast.BasicLit{Kind: token.STRING, Value: "`" + tc.raw + "`"}
			if got := tagValue(lit, tc.key); got != tc.want {
				t.Errorf("tagValue(%q, %q) = %q, want %q", tc.raw, tc.key, got, tc.want)
			}
		})
	}
	if got := tagValue(nil, "json"); got != "" {
		t.Errorf("tagValue(nil) = %q, want empty", got)
	}
}

func TestTypeString(t *testing.T) {
	f := parseSrc(t, `package p
type T struct {
	A types.String
	B *Foo
	C []int
	D map[string]int
	E OptNilBool
}`)
	got := map[string]string{}
	forEachStruct(f, func(_ string, st *ast.StructType) {
		for _, fld := range st.Fields.List {
			got[fld.Names[0].Name] = typeString(fld.Type)
		}
	})
	want := map[string]string{
		"A": "types.String",
		"B": "*Foo",
		"C": "[]int",
		"D": "map[string]int",
		"E": "OptNilBool",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("typeString(%s) = %q, want %q", k, got[k], v)
		}
	}
}

func TestGoExported(t *testing.T) {
	tests := map[string]string{
		"ou_note":         "OuNote",
		"account_linkage": "AccountLinkage",
		"label":           "Label",
		"scope_criteria":  "ScopeCriteria",
		"":                "",
		"a__b":            "AB",
	}
	for in, want := range tests {
		if got := goExported(in); got != want {
			t.Errorf("goExported(%q) = %q, want %q", in, got, want)
		}
	}
}
