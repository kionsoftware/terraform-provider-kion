package kalign

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStructsFromFile(t *testing.T) {
	f := parseSrc(t, `package generated
type OUNote struct {
	CreateUserID OptNilUint64 `+"`json:\"create_user_id\"`"+`
	Name         OptString    `+"`json:\"name\"`"+`
	Untagged     string
}
type Empty struct{}`)
	got := structsFromFile(f)

	ou, ok := got["OUNote"]
	if !ok {
		t.Fatalf("OUNote not parsed; got keys %v", keys(got))
	}
	if len(ou) != 2 { // Untagged (no json tag) is skipped
		t.Fatalf("OUNote fields = %d, want 2: %+v", len(ou), ou)
	}
	if ou[0].GoName != "CreateUserID" || ou[0].JSON != "create_user_id" || ou[0].GoType != "OptNilUint64" {
		t.Errorf("field 0 = %+v", ou[0])
	}
	if _, ok := got["Empty"]; ok {
		t.Errorf("Empty struct (no json fields) should be omitted")
	}
}

func TestStructsFromFile_namedBasicUnderlying(t *testing.T) {
	f := parseSrc(t, `package generated
type Stage uint64
type OptStage struct {
	Value Stage `+"`json:\"value\"`"+`
}
type Role struct {
	Stage OptStage  `+"`json:\"stage\"`"+`
	Name  OptString `+"`json:\"name\"`"+`
}`)
	got := structsFromFile(f)["Role"]
	if len(got) != 2 || got[0].Underlying != "uint64" || got[1].Underlying != "" {
		t.Errorf("Role fields = %+v, want Stage underlying uint64 only", got)
	}
}

func TestModelFromFile(t *testing.T) {
	f := parseSrc(t, `package ou_note
import "github.com/hashicorp/terraform-plugin-framework/types"
type OuNoteModel struct {
	CreateUserId types.Int64  `+"`tfsdk:\"create_user_id\"`"+`
	Name         types.String `+"`tfsdk:\"name\"`"+`
	NoTag        types.String
}`)
	m := modelFromFile(f, "ou_note", "x.go")
	if m == nil {
		t.Fatal("modelFromFile returned nil")
	}
	if m.Service != "ou_note" || m.Name != "OuNoteModel" || m.File != "x.go" {
		t.Errorf("meta = %+v", m)
	}
	if len(m.Fields) != 2 { // NoTag skipped
		t.Fatalf("fields = %d, want 2", len(m.Fields))
	}
	if m.Fields[0].GoName != "CreateUserId" || m.Fields[0].TFSDK != "create_user_id" || m.Fields[0].TFType != "types.Int64" {
		t.Errorf("field 0 = %+v", m.Fields[0])
	}
}

func TestModelFromFile_NoModel(t *testing.T) {
	f := parseSrc(t, `package p
type Foo struct{ A int `+"`tfsdk:\"a\"`"+` }`)
	if m := modelFromFile(f, "p", "x.go"); m != nil {
		t.Errorf("expected nil when no *Model struct present, got %+v", m)
	}
}

func TestAddFlexFuncs(t *testing.T) {
	f := parseSrc(t, `package flex
func OptStringToFramework() {}
func StringToFramework() {}
func (r recv) Method() {}`)
	got := map[string]bool{}
	addFlexFuncs(f, got)
	if !got["OptStringToFramework"] || !got["StringToFramework"] {
		t.Errorf("expected top-level funcs, got %v", got)
	}
	if got["Method"] {
		t.Errorf("method with receiver should not be recorded")
	}
}

func TestFileSource_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	// flex dir
	flexDir := filepath.Join(dir, "flex")
	mustWrite(t, filepath.Join(flexDir, "conv.go"), "package flex\nfunc OptStringToFramework() {}\n")
	mustWrite(t, filepath.Join(flexDir, "conv_test.go"), "package flex\nfunc ShouldBeIgnored() {}\n")
	// service dir
	svcDir := filepath.Join(dir, "service", "label")
	mustWrite(t, filepath.Join(svcDir, "label_schema_gen.go"),
		"package label\ntype LabelModel struct{ Key string `tfsdk:\"key\"` }\n")
	// sdk file
	sdkFile := filepath.Join(dir, "sdk.go")
	mustWrite(t, sdkFile, "package generated\ntype Label struct{ Key string `json:\"key\"` }\n")

	src := NewFileSource()

	sdk, err := src.SDKStructs(sdkFile)
	if err != nil || len(sdk["Label"]) != 1 {
		t.Fatalf("SDKStructs: %v, %+v", err, sdk)
	}
	flex, err := src.FlexFuncs(flexDir)
	if err != nil || !flex["OptStringToFramework"] || flex["ShouldBeIgnored"] {
		t.Fatalf("FlexFuncs: %v, %+v", err, flex)
	}
	models, err := src.ServiceModels(filepath.Join(dir, "service"), "")
	if err != nil || len(models) != 1 || models[0].Name != "LabelModel" {
		t.Fatalf("ServiceModels: %v, %+v", err, models)
	}
	// filtering by service name
	none, err := src.ServiceModels(filepath.Join(dir, "service"), "nonexistent")
	if err != nil || len(none) != 0 {
		t.Fatalf("ServiceModels filter: %v, %+v", err, none)
	}
}

func TestFileSource_Errors(t *testing.T) {
	src := NewFileSource()
	if _, err := src.SDKStructs("does-not-exist.go"); err == nil {
		t.Error("SDKStructs on missing file should error")
	}
	if _, err := src.FlexFuncs("does-not-exist-dir"); err == nil {
		t.Error("FlexFuncs on missing dir should error")
	}
}

func TestFileSource_ServiceModels_ParseError(t *testing.T) {
	dir := t.TempDir()
	svc := filepath.Join(dir, "svc", "bad")
	mustWrite(t, filepath.Join(svc, "bad_schema_gen.go"), "package bad\nthis is not valid go {{{")
	if _, err := NewFileSource().ServiceModels(filepath.Join(dir, "svc"), ""); err == nil {
		t.Error("expected a parse error for a malformed schema file")
	}
}

func keys(m map[string][]SDKField) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
