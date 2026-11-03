package kalign

import "testing"

func TestFindings(t *testing.T) {
	r := Resolved{
		MissingInSDK: []string{"a"},
		TypeMismatch: []string{"b"},
		MissingFlex:  []string{"c"},
		NestedAttrs:  []string{"n"}, // nested is not a finding
	}
	if got := r.Findings(); got != 3 {
		t.Errorf("Findings = %d, want 3", got)
	}
	if got := (Resolved{}).Findings(); got != 0 {
		t.Errorf("empty Findings = %d, want 0", got)
	}
}
