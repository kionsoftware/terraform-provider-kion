package crud

import (
	"os"
	"path/filepath"
	"testing"

	"terraform-provider-kion/internal/kgen/kfs"
)

// fixtureSDKVersion is the SDK sub-package fixture goldens are rendered against.
const fixtureSDKVersion = "v3_16"

// fixtureSpec is one resource to generate from fixture SDK sources.
type fixtureSpec struct {
	Name string     // resource name, e.g. "thing"
	Dir  string     // holds oas_{client,schemas,parameters}_gen.go and <Name>_schema_gen.go
	Ops  resOps     // the resource op-set
	DS   dsOps      // the data-source op-set; zero for none
	Arch *archetype // nil renders the plain entity archetype
	// TestValues is a test_values.yaml path; empty emits no acceptance tests.
	TestValues string
	// Model overrides the <Dir>/<Name>_schema_gen.go the model is read from.
	Model string
}

// generateFixture runs one resource through generateResource and returns each emitted file's source by base name.
func generateFixture(t *testing.T, spec fixtureSpec) (map[string]string, error) {
	t.Helper()
	prev := sdkVersion
	sdkVersion = fixtureSDKVersion
	t.Cleanup(func() { sdkVersion = prev })

	root := t.TempDir()
	dir := filepath.Join(root, "internal", "service", spec.Name)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	modelPath := spec.Model
	if modelPath == "" {
		modelPath = filepath.Join(spec.Dir, spec.Name+"_schema_gen.go")
	}
	model, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, spec.Name+"_schema_gen.go"), model, 0o600); err != nil {
		t.Fatal(err)
	}
	tv := spec.TestValues
	if tv == "" {
		tv = filepath.Join(root, "no_test_values.yaml")
	}

	g := &generator{fs: kfs.OS{}, src: NewFileSource(), root: root}
	if _, err := g.generateResource(root, spec.Name, spec.Ops, spec.DS, fixtureIndexFrom(t, spec.Dir), tv, false, spec.Arch, true); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := make(map[string]string, len(entries))
	for _, e := range entries {
		if e.Name() == spec.Name+"_schema_gen.go" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		out[e.Name()] = string(b)
	}
	return out, nil
}

// assertGolden compares got with testdata/golden/<name>; UPDATE_GOLDEN=1 rewrites it.
func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	golden := filepath.Join("testdata", "golden", name)
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(golden, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with UPDATE_GOLDEN=1 to create): %v", err)
	}
	if got != string(want) {
		t.Errorf("render drifted from %s; re-run with UPDATE_GOLDEN=1 if intentional", golden)
	}
}
