package crud

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noteModel() []ModelField {
	return []ModelField{
		{TFSDK: "id", GoName: "Id", Type: "types.String"},
		{TFSDK: "name", GoName: "Name", Type: "types.String"},
		{TFSDK: "last_update_user_id", GoName: "LastUpdateUserId", Type: "types.Int64"},
	}
}

func fieldByTF(t *testing.T, fields []rawField, json string) rawField {
	t.Helper()
	for _, f := range fields {
		if f.JSON == json {
			return f
		}
	}
	t.Fatalf("no raw field %q", json)
	return rawField{}
}

// TestRawModelFieldsDefaultKinds pins the untouched mapping: without a
// read_kinds entry a field keeps its plain scalar wire type.
func TestRawModelFieldsDefaultKinds(t *testing.T) {
	fields, idGo, err := rawModelFields(noteModel(), nil)
	require.NoError(t, err)
	assert.Equal(t, "Id", idGo)

	last := fieldByTF(t, fields, "last_update_user_id")
	assert.Equal(t, "int64", last.WireType)
	assert.Equal(t, "types.Int64Value(w.LastUpdateUserId)", last.ToExpr)
}

// TestRawModelFieldsNullIntKind covers the project_note failure: the private
// read renders last_update_user_id as {"Int":1,"Valid":true}, so the wire field
// must be the tolerant wrapper rather than int64.
func TestRawModelFieldsNullIntKind(t *testing.T) {
	fields, _, err := rawModelFields(noteModel(), map[string]string{"last_update_user_id": "null_int"})
	require.NoError(t, err)

	last := fieldByTF(t, fields, "last_update_user_id")
	assert.Equal(t, "*flex.NullInt", last.WireType, "pointer so omitempty still drops an unset value")
	assert.Equal(t, "flex.NullIntPtrToFramework(w.LastUpdateUserId)", last.ToExpr)
	assert.Equal(t, "flex.NullIntFromFramework(plan.LastUpdateUserId)", last.FromExpr)

	// Siblings are untouched.
	assert.Equal(t, "string", fieldByTF(t, fields, "name").WireType)
}

// TestRawModelFieldsNullStringKind covers the dashboard failure.
func TestRawModelFieldsNullStringKind(t *testing.T) {
	fields, _, err := rawModelFields(noteModel(), map[string]string{"name": "null_string"})
	require.NoError(t, err)

	name := fieldByTF(t, fields, "name")
	assert.Equal(t, "*flex.NullString", name.WireType)
	assert.Equal(t, "flex.NullStringPtrToFramework(w.Name)", name.ToExpr)
	assert.Equal(t, "flex.NullStringFromFramework(plan.Name)", name.FromExpr)
}

// TestRawModelFieldsRejectsMismatchedKind keeps a typo in private_endpoints.yaml
// from generating code that does not compile: the override must match the
// model's own type.
func TestRawModelFieldsRejectsMismatchedKind(t *testing.T) {
	_, _, err := rawModelFields(noteModel(), map[string]string{"name": "null_int"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a types.Int64 attribute")

	_, _, err = rawModelFields(noteModel(), map[string]string{"last_update_user_id": "null_string"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a types.String attribute")
}

func TestRawModelFieldsRejectsUnknownKind(t *testing.T) {
	_, _, err := rawModelFields(noteModel(), map[string]string{"name": "null_bool"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown read_kind")
}

func exemptionModel() []ModelField {
	return []ModelField{
		{TFSDK: "id", GoName: "Id", Type: "types.String"},
		{TFSDK: "ou_id", GoName: "OuId", Type: "types.Int64"},
		{TFSDK: "ou_cloud_access_role_id", GoName: "OuCloudAccessRoleId", Type: "types.Int64"},
	}
}

func byTF(model []ModelField) map[string]ModelField {
	m := map[string]ModelField{}
	for _, mf := range model {
		m[mf.TFSDK] = mf
	}
	return m
}

// TestBuildParentReadOwnerAndFilter covers what turns a no_read resource's
// empty-shell import into a real one: the owner comes from the record's own key
// (a collection may return records the path parent does not own), and the
// discriminator is declared so records of a neighboring kind are rejected.
func TestBuildParentReadOwnerAndFilter(t *testing.T) {
	pr := parentRead{
		Path:       "/v1/ou/{parent_id}/cloud-access-role-exemption",
		ParentTF:   "ou_id",
		ParentJSON: "OUID",
		Require:    "ou_cloud_access_role_id",
		Fields: []readShapeSub{
			{TF: "ou_cloud_access_role_id", From: "ou_cloud_access_role_id", Kind: "null_int"},
		},
	}
	d, err := buildParentRead("ou_cloud_access_role_exemption", pr, byTF(exemptionModel()), "Id")
	require.NoError(t, err)

	assert.True(t, d.HasParent, "a {parent_id} path needs the compound import id")
	assert.Equal(t, "OuId", d.ParentGo)
	assert.Equal(t, "OuCloudAccessRoleId", d.RequireGo)
	assert.True(t, d.UsesFlex)

	// The owner key is a flex.NullInt, not int64: it is a bare number on some
	// collections (OUID) and a SQL null wrapper on others (project_id).
	assert.Contains(t, d.WireStructGo, "OUID *flex.NullInt `json:\"OUID\"`")
	assert.Contains(t, d.WireStructGo, "OuCloudAccessRoleId *flex.NullInt `json:\"ou_cloud_access_role_id\"`")
	// Declared once, even though it is both the discriminator and a field.
	assert.Equal(t, 1, strings.Count(d.WireStructGo, "ou_cloud_access_role_id"))

	assert.Contains(t, d.FlattenGo, "m.OuId = flex.NullIntPtrToFramework(rec.OUID)",
		"parent comes from the record, not the path")
	assert.Contains(t, d.FlattenGo, "m.Id = types.StringValue(strconv.FormatInt(rec.ID, 10))")
}

// TestBuildParentReadFlatCollection: a collection with no {parent_id} needs no
// parent and keeps the plain import id (kion_aws_resource_tag).
func TestBuildParentReadFlatCollection(t *testing.T) {
	model := []ModelField{
		{TFSDK: "id", GoName: "Id", Type: "types.String"},
		{TFSDK: "resource_key", GoName: "ResourceKey", Type: "types.String"},
	}
	d, err := buildParentRead("aws_resource_tag", parentRead{
		Path:   "/v3/aws-resource-tag",
		Fields: []readShapeSub{{TF: "resource_key", From: "resource_key", Kind: "string"}},
	}, byTF(model), "Id")
	require.NoError(t, err)

	assert.False(t, d.HasParent)
	assert.Empty(t, d.RequireGo)
	assert.NotContains(t, d.WireStructGo, "flex.NullInt")
	assert.Contains(t, d.FlattenGo, "m.ResourceKey = types.StringValue(rec.ResourceKey)")
}

func noteCreateIDModel() []ModelField {
	return []ModelField{
		{TFSDK: "id", GoName: "Id", Type: "types.String"},
		{TFSDK: "funding_source_id", GoName: "FundingSourceId", Type: "types.Int64"},
		{TFSDK: "name", GoName: "Name", Type: "types.String"},
	}
}

// TestResolveRawCreateID covers #71: POST /v2/funding-source-note answers with
// no record_id, so the resource recovers the id from its parent collection
// rather than writing 0 to state.
func TestResolveRawCreateID(t *testing.T) {
	d, err := resolveRaw("funding_source_note", rawResourceOps{
		Create: rawOp{Method: "POST", Path: "/v2/funding-source-note"},
		Read:   rawOp{Method: "GET", Path: "/v2/funding-source-note/{id}"},
		CreateID: &createID{
			ListPath: "/v2/funding-source/{parent_id}/funding-source-note",
			ParentTF: "funding_source_id",
		},
	}, noteCreateIDModel())
	require.NoError(t, err)

	require.NotNil(t, d.CreateID)
	assert.Equal(t, "/v2/funding-source/{parent_id}/funding-source-note", d.CreateID.ListPath)
	assert.Equal(t, "FundingSourceId", d.CreateID.ParentGo, "the model field the parent id is read from")
}

// TestResolveRawWithoutCreateID pins the default: a create that does report a
// record_id renders no collection diff at all.
func TestResolveRawWithoutCreateID(t *testing.T) {
	d, err := resolveRaw("dashboard", rawResourceOps{
		Create: rawOp{Method: "POST", Path: "/beta/dashboard"},
		Read:   rawOp{Method: "GET", Path: "/v1/dashboard/{id}"},
	}, noteCreateIDModel())
	require.NoError(t, err)
	assert.Nil(t, d.CreateID)
}

func TestBuildCreateIDRejectsBadDeclaration(t *testing.T) {
	m := byTF(noteCreateIDModel())

	_, err := buildCreateID("x", createID{ListPath: "/v2/funding-source-note", ParentTF: "funding_source_id"}, m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must contain {parent_id}")

	_, err = buildCreateID("x", createID{ListPath: "/v2/fs/{parent_id}/note"}, m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent_tf is required")

	_, err = buildCreateID("x", createID{ListPath: "/v2/fs/{parent_id}/note", ParentTF: "nope"}, m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not a model attribute")

	// The parent id is read with ValueInt64(), so anything else would generate
	// a file that does not compile.
	_, err = buildCreateID("x", createID{ListPath: "/v2/fs/{parent_id}/note", ParentTF: "name"}, m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a types.Int64 attribute")
}

func TestBuildParentReadRejectsBadDeclaration(t *testing.T) {
	m := byTF(exemptionModel())

	_, err := buildParentRead("x", parentRead{Path: "/v1/ou/{parent_id}/e"}, m, "Id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent_tf is empty")

	_, err = buildParentRead("x", parentRead{Path: "/v1/ou/{parent_id}/e", ParentTF: "nope"}, m, "Id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not a model attribute")

	_, err = buildParentRead("x", parentRead{Path: "/v1/ou/{parent_id}/e", ParentTF: "ou_id"}, m, "Id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "needs parent_json")
}

// TestBuildParentReadNamedRecords: a collection whose records sit under a named
// key inside data ({"data":{"ou_exemptions":[...]}}) decodes that key, not data.
func TestBuildParentReadNamedRecords(t *testing.T) {
	d, err := buildParentRead("ou_cloud_access_role_exemption", parentRead{
		Path:       "/v1/ou/{parent_id}/ou-cloud-access-role",
		Records:    "ou_exemptions",
		ParentTF:   "ou_id",
		ParentJSON: "ou_id",
		Fields: []readShapeSub{
			{TF: "ou_cloud_access_role_id", From: "ou_cloud_access_role_id", Kind: "int"},
		},
	}, byTF(exemptionModel()), "Id")
	require.NoError(t, err)

	assert.Contains(t, d.EnvelopeGo, "Records []ou_cloud_access_role_exemptionRecord `json:\"ou_exemptions\"`")
	assert.Contains(t, d.EnvelopeGo, "`json:\"data\"`")
	assert.Equal(t, "env.Data.Records", d.RecordsExpr)
	assert.Contains(t, d.FlattenGo, "m.OuCloudAccessRoleId = types.Int64Value(rec.OuCloudAccessRoleId)")
}

// TestBuildParentReadFlatEnvelope pins the default: records are data itself.
func TestBuildParentReadFlatEnvelope(t *testing.T) {
	d, err := buildParentRead("aws_resource_tag", parentRead{Path: "/v3/aws-resource-tag"},
		byTF([]ModelField{{TFSDK: "id", GoName: "Id", Type: "types.String"}}), "Id")
	require.NoError(t, err)
	assert.Equal(t, "Data []aws_resource_tagRecord `json:\"data\"`", d.EnvelopeGo)
	assert.Equal(t, "env.Data", d.RecordsExpr)
}

// TestNoReadScanPrefersParentRead: the acceptance test's existence check reads
// the same collection Read does, not the spec's unrelated read path.
func TestNoReadScanPrefersParentRead(t *testing.T) {
	pr := &parentRead{Path: "/v1/ou/{parent_id}/ou-cloud-access-role", ParentTF: "ou_id", Records: "ou_exemptions"}
	s := noReadScanFor("ou_cloud_access_role_exemption", pr, "/v3/ou/{id}/cloud-rule/exemption", exemptionModel())
	assert.Equal(t, noReadScan{ListPath: "/v1/ou/{id}/ou-cloud-access-role", ParentIDTF: "ou_id", RecordsKey: "ou_exemptions"}, s)

	flat := noReadScanFor("aws_resource_tag", &parentRead{Path: "/v3/aws-resource-tag"}, "/v3/aws-resource-tag", nil)
	assert.Equal(t, noReadScan{ListPath: "/v3/aws-resource-tag"}, flat)

	none := noReadScanFor("ou_thing", nil, "/v3/ou/{id}/thing", []ModelField{{TFSDK: "ou_id"}})
	assert.Equal(t, noReadScan{ListPath: "/v3/ou/{id}/thing", ParentIDTF: "ou_id"}, none)
}
