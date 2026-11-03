package kalign

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindingKey(t *testing.T) {
	f := Finding{Service: "ou_enforcement", Attr: "cloud_rule_id", Kind: KindType}
	if got := f.Key(); got != "ou_enforcement.cloud_rule_id:type" {
		t.Errorf("Key = %q", got)
	}
}

func TestResolve_recordsActionableFindings(t *testing.T) {
	model := ServiceModel{
		Service: "sc", Name: "ScModel",
		Fields: []ModelField{
			{TFSDK: "id", TFType: "types.Int64"},
			{TFSDK: "bad", TFType: "types.String"},
		},
	}
	sdk := map[string][]SDKField{"Sc": {
		{JSON: "id", GoType: "OptUint64"},
		{JSON: "bad", GoType: "OptUint64"},
	}}
	r := Resolve(model, sdk, map[string]bool{"OptUint64ToFramework": true})
	got := findingKeys(r.ActionableFindings())
	want := []string{"sc.bad:type"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("findings = %v, want %v", got, want)
	}
	if len(r.ActionableFindings()) != r.Actionable() {
		t.Errorf("ActionableFindings disagrees with Actionable")
	}
}

func TestCompareBaseline(t *testing.T) {
	findings := []Finding{
		{Service: "a", Attr: "x", Kind: KindType},
		{Service: "b", Attr: "y", Kind: KindFlex},
	}
	base := Baseline{
		"a.x:type": "accepted",
		"c.z:flex": "fixed since",
	}
	unexpected, stale := CompareBaseline(findings, base)
	if len(unexpected) != 1 || unexpected[0].Key() != "b.y:flex" {
		t.Errorf("unexpected = %v", unexpected)
	}
	if len(stale) != 1 || stale[0] != "c.z:flex" {
		t.Errorf("stale = %v", stale)
	}

	unexpected, stale = CompareBaseline(findings[:1], Baseline{"a.x:type": "ok"})
	if len(unexpected)+len(stale) != 0 {
		t.Errorf("clean compare = %v, %v", unexpected, stale)
	}
}

func TestReportRatchet(t *testing.T) {
	var b bytes.Buffer
	ok := ReportRatchet(&b, []Finding{{Service: "a", Attr: "x", Kind: KindType}}, Baseline{"a.x:type": "why"}, "base.yaml")
	if !ok {
		t.Errorf("baselined finding must pass: %s", b.String())
	}

	b.Reset()
	ok = ReportRatchet(&b, []Finding{{Service: "n", Attr: "w", Kind: KindFlex}}, Baseline{"s.t:type": "why"}, "base.yaml")
	if ok {
		t.Fatal("new finding and stale entry must fail")
	}
	for _, want := range []string{"n.w:flex", "s.t:type", "base.yaml"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("report missing %q:\n%s", want, b.String())
		}
	}

	b.Reset()
	if ReportRatchet(&b, []Finding{{Service: "a", Attr: "x", Kind: KindType}}, Baseline{"a.x:type": "  "}, "base.yaml") {
		t.Error("an entry with no reason must fail")
	}
	if ReportRatchet(&b, []Finding{{Service: "a", Attr: "x", Kind: KindType}}, Baseline{"a.x:type": "TODO: untriaged"}, "base.yaml") {
		t.Error("an untriaged TODO entry must fail")
	}
}

func TestBaselineRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "align_baseline.yaml")
	if _, err := LoadBaseline(path); err == nil {
		t.Fatal("a missing baseline must be an error, not an empty ratchet")
	}
	findings := []Finding{
		{Service: "b", Attr: "y", Kind: KindFlex},
		{Service: "a", Attr: "x", Kind: KindType},
	}
	if err := WriteBaseline(path, findings, Baseline{"a.x:type": "kept reason", "gone.q:type": "dropped"}); err != nil {
		t.Fatal(err)
	}
	got, err := LoadBaseline(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got["a.x:type"] != "kept reason" || !strings.HasPrefix(got["b.y:flex"], "TODO") {
		t.Errorf("round trip = %v", got)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(string(raw), "a.x:type") > strings.Index(string(raw), "b.y:flex") {
		t.Errorf("entries must be sorted:\n%s", raw)
	}
}

func findingKeys(fs []Finding) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, f.Key())
	}
	return out
}
