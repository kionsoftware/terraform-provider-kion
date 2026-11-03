package crud

import "testing"

// codegen/renames.yaml applies to the read as well as to request bodies. The
// OpenID group association's read calls the attribute should_update_on_login
// while its body and model say update_on_login; without the rename the value
// was never read, so it was lost on import.
func TestRespBindsHonorsRenames(t *testing.T) {
	byTF := map[string]ModelField{
		"id":              {GoName: "Id", TFSDK: "id", Type: "types.String"},
		"update_on_login": {GoName: "UpdateOnLogin", TFSDK: "update_on_login", Type: "types.Bool"},
	}
	fields := []Field{
		{GoName: "ID", JSONName: "id", Type: "OptUint64"},
		{GoName: "ShouldUpdateOnLogin", JSONName: "should_update_on_login", Type: "OptNilBool"},
	}
	binds, _, err := respBinds(fields, byTF, "id", "v.", nil, nil, false, map[string]string{"should_update_on_login": "update_on_login"})
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range binds {
		if b.ModelGo == "UpdateOnLogin" {
			return
		}
	}
	t.Errorf("renamed response field not bound: %+v", binds)
}
