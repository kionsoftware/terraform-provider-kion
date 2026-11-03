package kalign

import (
	"bytes"
	"strings"
	"testing"
)

func sampleResolved() Resolved {
	return Resolved{
		Model:   ServiceModel{Service: "ou_note", Name: "OuNoteModel", Fields: make([]ModelField, 2)},
		SDKType: "OUNote", Overlap: 2,
		Pairs: []Pair{
			{Model: ModelField{GoName: "Id", TFSDK: "id"}, SDK: SDKField{GoName: "ID", GoType: "OptUint64"}, FlexFn: "OptUint64ToFramework", HaveFlex: true},
			{Model: ModelField{GoName: "Name", TFSDK: "name"}, SDK: SDKField{GoName: "Name", GoType: "OptString"}, FlexFn: "OptStringToFramework", HaveFlex: true},
		},
	}
}

func TestCheckOne_Clean(t *testing.T) {
	var b bytes.Buffer
	n := checkOne(&errWriter{w: &b}, sampleResolved())
	if n != 0 {
		t.Errorf("findings = %d, want 0", n)
	}
	out := b.String()
	for _, want := range []string{"ou_note", "generated.OUNote (overlap 2/2)", "ok    id -> OptUint64"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

// checkOne prints every finding but counts only the actionable ones. DRIFT is
// advisory: it compares one schema to one SDK struct, which is not how this
// provider binds, so an attribute can be perfectly bound and still appear
// there. See Resolved.Actionable.
func TestCheckOne_Findings(t *testing.T) {
	r := sampleResolved()
	r.MissingInSDK = []string{"ghost"}
	r.TypeMismatch = []string{"bad: schema types.String vs SDK OptUint64"}
	r.MissingFlex = []string{"OptWeirdToFramework (for field \"weird\")"}
	var b bytes.Buffer
	n := checkOne(&errWriter{w: &b}, r)
	if n != 2 {
		t.Errorf("actionable findings = %d, want 2 (the DRIFT one is advisory)", n)
	}
	out := b.String()
	for _, want := range []string{"DRIFT no SDK field", "TYPE  bad:", "FLEX  missing converter"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestCheckOne_NoSDKType(t *testing.T) {
	var b bytes.Buffer
	checkOne(&errWriter{w: &b}, Resolved{Model: ServiceModel{Service: "x", Name: "XModel"}})
	if !strings.Contains(b.String(), "no SDK type overlaps") {
		t.Errorf("expected no-overlap message:\n%s", b.String())
	}
}

func TestGenOne_Flatten(t *testing.T) {
	var b bytes.Buffer
	todos := genOne(&errWriter{w: &b}, sampleResolved())
	if todos != 0 {
		t.Errorf("todos = %d, want 0", todos)
	}
	out := b.String()
	for _, want := range []string{
		"func flattenOuNote(in *generated.OUNote, m *OuNoteModel) {",
		"m.Id = flex.OptUint64ToFramework(in.ID)",
		"m.Name = flex.OptStringToFramework(in.Name)",
		"DO NOT EDIT",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("gen output missing %q:\n%s", want, out)
		}
	}
}

func TestGenOne_NestedAndMissingFlexBecomeTODOs(t *testing.T) {
	r := sampleResolved()
	r.Pairs = append(r.Pairs,
		Pair{Model: ModelField{GoName: "Criteria", TFSDK: "criteria"}, SDK: SDKField{GoType: "*Nested"}, Nested: true},
		Pair{Model: ModelField{GoName: "Weird", TFSDK: "weird"}, SDK: SDKField{GoName: "Weird", GoType: "OptWeird"}, FlexFn: "OptWeirdToFramework", HaveFlex: false},
	)
	var b bytes.Buffer
	todos := genOne(&errWriter{w: &b}, r)
	if todos != 2 {
		t.Errorf("todos = %d, want 2", todos)
	}
	out := b.String()
	if !strings.Contains(out, "TODO(nested): m.Criteria") {
		t.Errorf("missing nested TODO:\n%s", out)
	}
	if !strings.Contains(out, "TODO(flex): m.Weird") {
		t.Errorf("missing flex TODO:\n%s", out)
	}
}

func TestGenOne_SkipLowConfidence(t *testing.T) {
	var b bytes.Buffer
	genOne(&errWriter{w: &b}, Resolved{Model: ServiceModel{Service: "x"}, LowConfidence: true})
	if !strings.Contains(b.String(), "skipped (no confident SDK type match)") {
		t.Errorf("expected skip comment:\n%s", b.String())
	}
}

func TestCheckOne_LowConfidence(t *testing.T) {
	var b bytes.Buffer
	r := sampleResolved()
	r.LowConfidence = true
	checkOne(&errWriter{w: &b}, r)
	if !strings.Contains(b.String(), "LOW CONFIDENCE") {
		t.Errorf("expected low-confidence marker:\n%s", b.String())
	}
}
