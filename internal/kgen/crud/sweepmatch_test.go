package crud

import "testing"

// GET /v3/idms/{id}/group-association serves SAML only and crashes on any other
// IDMS type (#48), so the group-association sweepers walk SAML parents only.
func TestSweepMatchExpr(t *testing.T) {
	idx := sdkIndex{structs: map[string]Struct{
		"IDMS": {Fields: []Field{
			{GoName: "ID", JSONName: "id", Type: "OptUint64"},
			{GoName: "IdmsTypeID", JSONName: "idms_type_id", Type: "OptNilUint64"},
			{GoName: "Kind", JSONName: "kind", Type: "int64"},
		}},
	}}
	for _, tc := range []struct {
		match map[string]string
		want  string
	}{
		{map[string]string{"idms_type_id": "3"}, "item.IdmsTypeID.Set && !item.IdmsTypeID.Null && item.IdmsTypeID.Value == 3"},
		{map[string]string{"kind": "2"}, "item.Kind == 2"},
	} {
		got, err := sweepMatchExpr("IDMS", tc.match, idx)
		if err != nil || got != tc.want {
			t.Errorf("sweepMatchExpr(%v) = %q, %v; want %q", tc.match, got, err, tc.want)
		}
	}
	if _, err := sweepMatchExpr("IDMS", map[string]string{"nope": "1"}, idx); err == nil {
		t.Error("a field the element lacks must be an error")
	}
	if _, err := sweepMatchExpr("IDMS", map[string]string{"idms_type_id": "x"}, idx); err == nil {
		t.Error("a non-integer value must be an error")
	}
}
