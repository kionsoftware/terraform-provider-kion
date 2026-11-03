package crud

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"
)

func TestRenderSweep_labelReal(t *testing.T) {
	got, _, err := renderSweep(labelResourceModel(t))
	if err != nil {
		t.Fatalf("renderSweep: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "sweep.go", got, parser.ParseComments); err != nil {
		t.Fatalf("generated sweep.go does not parse: %v\n%s", err, got)
	}
	wants := []string{
		`resource.AddTestSweepers("kion_label"`,
		"func sweepLabel(_ string) error",
		"conn.Client.GetLabelIndex(ctx, generated.GetLabelIndexParams{",
		"conn.Client.DeleteLabel(ctx, generated.DeleteLabelParams{ID: id})",
		"func sweepLabelMatch(item generated.Label) bool",
		`strings.HasPrefix(s, "test-acc")`,
	}
	for _, w := range wants {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("real sweeper missing %q", w)
		}
	}
}

// A resource that cannot be enumerated must register NOTHING. Registering a
// sweeper whose body returns nil reports success to `make sweep` while orphaned
// test-acc records pile up, which is worse than having no sweeper at all.
func TestRenderSweep_unsweepableRegistersNothing(t *testing.T) {
	for _, tc := range []struct {
		name string
		mut  func(*ResourceModel)
		want string
	}{
		{"no list", func(rm *ResourceModel) { rm.List = nil; rm.ListDowngrade = "envelope has no items field" }, "envelope has no items field"},
		{"no delete", func(rm *ResourceModel) { rm.Delete = nil }, "no delete endpoint"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rm := labelResourceModel(t)
			tc.mut(&rm)
			got, reason, err := renderSweep(rm)
			if err != nil {
				t.Fatalf("renderSweep: %v", err)
			}
			if _, err := parser.ParseFile(token.NewFileSet(), "sweep.go", got, parser.ParseComments); err != nil {
				t.Fatalf("generated sweep.go does not parse: %v\n%s", err, got)
			}
			for _, unwanted := range []string{"AddTestSweepers", "func sweepLabel", "GetLabelIndex"} {
				if bytes.Contains(got, []byte(unwanted)) {
					t.Errorf("unsweepable resource must not emit %q:\n%s", unwanted, got)
				}
			}
			if reason == "" {
				t.Error("renderSweep returned no reason")
			}
			if !bytes.Contains(got, []byte(tc.want)) {
				t.Errorf("generated file does not explain %q:\n%s", tc.want, got)
			}
		})
	}
}
