package crud

import "testing"

func TestClientMethods_parsesRouteAndSignature(t *testing.T) {
	got, err := NewFileSource().ClientMethods("testdata/oas_client_gen.go")
	if err != nil {
		t.Fatalf("ClientMethods: %v", err)
	}
	byName := map[string]ClientMethod{}
	for _, m := range got {
		byName[m.Name] = m
	}

	post := byName["PostLabel"]
	if post.HTTPMethod != "POST" || post.Path != "/v3/label" {
		t.Errorf("PostLabel route = %s %s, want POST /v3/label", post.HTTPMethod, post.Path)
	}
	if post.BodyType != "OptCreateLabel" || post.ParamsType != "" || post.ResultType != "PostLabelRes" {
		t.Errorf("PostLabel sig = body=%q params=%q res=%q", post.BodyType, post.ParamsType, post.ResultType)
	}

	get := byName["GetLabel"]
	if get.HTTPMethod != "GET" || get.Path != "/v3/label/{id}" {
		t.Errorf("GetLabel route = %s %s, want GET /v3/label/{id}", get.HTTPMethod, get.Path)
	}
	if get.BodyType != "" || get.ParamsType != "GetLabelParams" || get.ResultType != "GetLabelRes" {
		t.Errorf("GetLabel sig = body=%q params=%q res=%q", get.BodyType, get.ParamsType, get.ResultType)
	}

	patch := byName["PatchLabel"]
	if patch.BodyType != "OptUpdateLabel" || patch.ParamsType != "PatchLabelParams" {
		t.Errorf("PatchLabel sig = body=%q params=%q", patch.BodyType, patch.ParamsType)
	}

	del := byName["DeleteLabel"]
	if del.HTTPMethod != "DELETE" || del.ParamsType != "DeleteLabelParams" {
		t.Errorf("DeleteLabel sig = %s params=%q", del.HTTPMethod, del.ParamsType)
	}
}

func TestStructs_fieldsTagsOptionality(t *testing.T) {
	m, err := NewFileSource().Structs("testdata/oas_schemas_gen.go")
	if err != nil {
		t.Fatalf("Structs: %v", err)
	}
	cl := m["CreateLabel"]
	want := map[string]Field{
		"Color": {GoName: "Color", JSONName: "color", Type: "string", Optional: false},
		"ID":    {GoName: "ID", JSONName: "id", Type: "OptUint64", Optional: true},
		"Key":   {GoName: "Key", JSONName: "key", Type: "string", Optional: false},
		"Value": {GoName: "Value", JSONName: "value", Type: "string", Optional: false},
	}
	got := map[string]Field{}
	for _, f := range cl.Fields {
		got[f.GoName] = f
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("CreateLabel.%s = %+v, want %+v", k, got[k], w)
		}
	}

	p, err := NewFileSource().Structs("testdata/oas_parameters_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	if f := p["GetLabelParams"].Fields; len(f) != 1 || f[0].GoName != "ID" || f[0].Type != "int64" {
		t.Errorf("GetLabelParams fields = %+v", f)
	}
}

func TestStructs_sliceFieldType(t *testing.T) {
	m, err := NewFileSource().Structs("testdata/oas_schemas_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	var value Field
	for _, f := range m["OptNilLabelArray"].Fields {
		if f.GoName == "Value" {
			value = f
		}
	}
	if value.Type != "[]Label" {
		t.Errorf("OptNilLabelArray.Value type = %q, want []Label", value.Type)
	}
}

func TestMarkerImpls_findsResponseEnvelope(t *testing.T) {
	m, err := NewFileSource().MarkerImpls("testdata/oas_schemas_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	if got := m["getLabelRes"]; len(got) != 1 || got[0] != "LabelResponse" {
		t.Errorf("getLabelRes impls = %v, want [LabelResponse]", got)
	}
}

func TestModelFields_readsTFSDKTags(t *testing.T) {
	fs, err := NewFileSource().ModelFields("testdata/label_schema_gen.go", "LabelModel")
	if err != nil {
		t.Fatal(err)
	}
	byTF := map[string]ModelField{}
	for _, f := range fs {
		byTF[f.TFSDK] = f
	}
	if f := byTF["id"]; f.GoName != "Id" || f.Type != "types.String" {
		t.Errorf("id field = %+v", f)
	}
	if f := byTF["color"]; f.GoName != "Color" || f.Type != "types.String" {
		t.Errorf("color field = %+v", f)
	}
}
