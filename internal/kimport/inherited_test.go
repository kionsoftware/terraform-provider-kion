package kimport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/importmanifest"
)

// exemptionResource is a synthetic row over an inherited, kind-mixing
// collection, modeled on the older cloud-rule exemption listing. It exercises
// ParentIDJSON and RequireValidField; the real exemption resources no longer
// read that listing (see records_key_test.go).
func exemptionResource() importmanifest.Resource {
	return importmanifest.Resource{
		TFType:            "kion_ou_cloud_access_role_exemption",
		Kind:              "ou_cloud_access_role_exemption",
		Archetype:         "no_read",
		ReadShape:         importmanifest.ShapeParentList,
		Readable:          true,
		RequireValidField: "ou_cloud_access_role_id",
		Parent: &importmanifest.Parent{
			Kind:          "ou",
			ListPath:      "/v3/ou",
			ChildPath:     "/v1/ou/{parent_id}/cloud-access-role-exemption",
			ParentIDField: "ou_id",
			ParentIDJSON:  "OUID",
		},
		ImportID: importmanifest.ImportID{Format: importmanifest.FormatParentSlashKey, KeyField: "id"},
	}
}

func validWrap(n int) map[string]any { return map[string]any{"Int": float64(n), "Valid": true} }
func absentWrap() map[string]any     { return map[string]any{"Int": float64(0), "Valid": false} }

// TestInheritedCollectionUsesRecordOwner covers the two traps in the exemption
// collections at once. The same record is returned under every OU in its
// subtree, so the id in the path is not its owner -- taking the owner from the
// record's own OUID is what makes one import block per record, addressed to the
// OU that actually owns it. A live install returned 329 rows for 22 records.
func TestInheritedCollectionUsesRecordOwner(t *testing.T) {
	// Record 5 is owned by OU 1 but shows up under OUs 1, 2 and 3.
	shared := []map[string]any{
		{"id": float64(5), "OUID": float64(1), "ou_cloud_access_role_id": validWrap(3)},
	}
	l := &routeLister{routes: map[string]any{
		"/v3/ou": []map[string]any{
			{"id": float64(1)}, {"id": float64(2)}, {"id": float64(3)},
		},
		"/v1/ou/1/cloud-access-role-exemption": shared,
		"/v1/ou/2/cloud-access-role-exemption": shared,
		"/v1/ou/3/cloud-access-role-exemption": shared,
	}}

	res := Enumerate(context.Background(), l, exemptionResource())
	require.Equal(t, "ok", res.Status, res.Reason)

	ids := make([]string, 0, len(res.Records))
	for _, r := range res.Records {
		ids = append(ids, r.ID)
	}
	// Every copy addresses the owning OU, never the OU it was fetched through.
	for _, id := range ids {
		assert.Equal(t, "1/5", id, "owner comes from OUID, not the path")
	}
}

// TestWrongKindRecordsAreFiltered covers the other trap: the collection returns
// cloud RULE exemptions alongside cloud ACCESS ROLE exemptions. Only 6 of 22
// records on a live install were the latter; without the filter the other 16
// imported as this resource type while being something else.
func TestWrongKindRecordsAreFiltered(t *testing.T) {
	l := &routeLister{routes: map[string]any{
		"/v3/ou": []map[string]any{{"id": float64(1)}},
		"/v1/ou/1/cloud-access-role-exemption": []map[string]any{
			{"id": float64(5), "OUID": float64(1), "ou_cloud_access_role_id": validWrap(3)},
			// A cloud RULE exemption: same collection, different resource.
			{"id": float64(8), "OUID": float64(1), "ou_cloud_access_role_id": absentWrap(), "ou_cloud_rule_id": validWrap(46)},
			// No discriminator key at all.
			{"id": float64(9), "OUID": float64(1)},
		},
	}}

	res := Enumerate(context.Background(), l, exemptionResource())
	require.Equal(t, "ok", res.Status, res.Reason)
	require.Len(t, res.Records, 1, "only the cloud access role exemption survives")
	assert.Equal(t, "1/5", res.Records[0].ID)
	assert.Contains(t, res.Reason, `no "ou_cloud_access_role_id" set`,
		"the drop is reported, never silent, and names the discriminator")
}

// TestEveryRecordFilteredOutIsEmptyNotOk is the project case: on a live install
// all 12 records its collection returns are cloud rule exemptions, so the
// resource has none rather than twelve.
func TestEveryRecordFilteredOutIsEmptyNotOk(t *testing.T) {
	l := &routeLister{routes: map[string]any{
		"/v3/ou": []map[string]any{{"id": float64(1)}},
		"/v1/ou/1/cloud-access-role-exemption": []map[string]any{
			{"id": float64(8), "OUID": float64(1), "ou_cloud_access_role_id": absentWrap()},
		},
	}}

	res := Enumerate(context.Background(), l, exemptionResource())
	assert.Equal(t, "empty", res.Status)
	assert.Contains(t, res.Reason, `no "ou_cloud_access_role_id" set`)
}

// TestOwnerKeyAcceptsBothRenderings: OUID is a bare number on the wire while
// project_id is a SQL null wrapper, and the enumerator has to read either.
func TestOwnerKeyAcceptsBothRenderings(t *testing.T) {
	assert.Equal(t, "11", stringifyWrapper(float64(11)), "bare number")
	assert.Equal(t, "9", stringifyWrapper(validWrap(9)), "sql null wrapper")
	assert.Equal(t, "", stringifyWrapper(absentWrap()), "not valid has no owner")
	assert.Equal(t, "", stringifyWrapper(nil))
}

func TestIsValidWrapper(t *testing.T) {
	assert.True(t, isValidWrapper(validWrap(3)))
	assert.False(t, isValidWrapper(absentWrap()))
	assert.False(t, isValidWrapper(nil))
	assert.True(t, isValidWrapper(float64(3)), "a bare id counts as present")
	assert.False(t, isValidWrapper(float64(0)), "a literal 0 is an absent id")

	// custom_variable_override's "override" is a plain object with no Valid
	// flag; a wrapper that has one keeps its original meaning above.
	assert.True(t, isValidWrapper(map[string]any{"value": "x"}), "plain non-empty object is present")
	assert.False(t, isValidWrapper(map[string]any{}), "empty object is not")
}

// TestParentMatchSkipsIneligibleParents covers /v3/idms/{id}/group-association,
// which fails for every IDMS that is not SAML. Those parents must never be
// read, so they cannot turn the resource into an error.
func TestParentMatchSkipsIneligibleParents(t *testing.T) {
	r := importmanifest.Resource{
		TFType:    "kion_saml_group_association",
		ReadShape: importmanifest.ShapeParentList,
		Readable:  true,
		Parent: &importmanifest.Parent{
			Kind: "idms", ListPath: "/v3/idms",
			ChildPath:     "/v3/idms/{parent_id}/group-association",
			ParentIDField: "idms_id",
			Match:         map[string]string{"idms_type_id": "3"},
		},
		ImportID: importmanifest.ImportID{Format: importmanifest.FormatID},
	}
	l := &routeLister{routes: map[string]any{
		"/v3/idms": []map[string]any{
			{"id": float64(1), "idms_type_id": float64(1)},
			{"id": float64(2), "idms_type_id": float64(3)},
			{"id": float64(3), "idms_type_id": nil},
		},
		"/v3/idms/2/group-association": []map[string]any{{"id": float64(7)}},
	}}

	res := Enumerate(context.Background(), l, r)
	require.Equal(t, "ok", res.Status, res.Reason)
	assert.Empty(t, res.Reason)
	require.Len(t, res.Records, 1)
	assert.Equal(t, "7", res.Records[0].ID)
	assert.NotContains(t, l.calls, "/v3/idms/1/group-association")
	assert.NotContains(t, l.calls, "/v3/idms/3/group-association")

	// No eligible parent at all is empty, not an error.
	l2 := &routeLister{routes: map[string]any{
		"/v3/idms": []map[string]any{{"id": float64(1), "idms_type_id": float64(1)}},
	}}
	res = Enumerate(context.Background(), l2, r)
	assert.Equal(t, "empty", res.Status, res.Reason)
	assert.Contains(t, res.Reason, "1 parent(s) not matching")
}
