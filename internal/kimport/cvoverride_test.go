package kimport

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/importmanifest"
)

// cvOverrideResource mirrors the manifest row for
// kion_custom_variable_override: three parent entity types, a discriminator
// that keeps only the variables actually overridden at the entity, and a
// three-part import id naming the parent's type as well as its id.
func cvOverrideResource() importmanifest.Resource {
	return importmanifest.Resource{
		TFType:            "kion_custom_variable_override",
		Kind:              "custom_variable_override",
		Archetype:         "cv_override",
		ReadShape:         importmanifest.ShapeParentList,
		Readable:          true,
		RequireValidField: "override",
		Parents: []importmanifest.Parent{
			{Kind: "account", ListPath: "/v3/account", ChildPath: "/v3/account/{parent_id}/custom-variable", ParentIDField: "entity_id"},
			{Kind: "ou", ListPath: "/v3/ou", ChildPath: "/v3/ou/{parent_id}/custom-variable", ParentIDField: "entity_id"},
			{Kind: "project", ListPath: "/v3/project", ChildPath: "/v3/project/{parent_id}/custom-variable", ParentIDField: "entity_id"},
		},
		ImportID: importmanifest.ImportID{
			Format:   importmanifest.FormatKindParentSlashKey,
			KeyField: "custom_variable_id",
		},
	}
}

// overridden and inherited are the two shapes /v3/{entity}/{id}/custom-variable
// returns: "override" is non-null only on a variable actually set at the entity.
func overridden(cvID int, value string) map[string]any {
	return map[string]any{
		"custom_variable_id":   float64(cvID),
		"custom_variable_type": "string",
		"inherited":            map[string]any{"entity_id": float64(cvID), "entity_type": "custom_variable", "value": "base"},
		"override":             map[string]any{"value": value},
		"value":                value,
	}
}

func inherited(cvID int) map[string]any {
	return map[string]any{
		"custom_variable_id":   float64(cvID),
		"custom_variable_type": "string",
		"inherited":            map[string]any{"entity_id": float64(cvID), "entity_type": "custom_variable", "value": "base"},
		"override":             nil,
		"value":                "base",
	}
}

// TestCVOverrideEnumeratesAllThreeParentKinds is the whole point of the fix:
// the resource is polymorphic, so an account-only read would find 5 of the 21
// overrides a live demo install actually has.
func TestCVOverrideEnumeratesAllThreeParentKinds(t *testing.T) {
	t.Parallel()
	l := &routeLister{routes: map[string]any{
		"/v3/account":                    []map[string]any{rec("id", float64(15))},
		"/v3/account/15/custom-variable": []map[string]any{overridden(12, "i2"), inherited(1)},
		"/v3/ou":                         []map[string]any{rec("id", float64(2))},
		"/v3/ou/2/custom-variable":       []map[string]any{overridden(1, "Corporate IT")},
		"/v3/project":                    []map[string]any{rec("id", float64(9))},
		"/v3/project/9/custom-variable":  []map[string]any{overridden(3, "HIPAA")},
	}}

	res := Enumerate(context.Background(), l, cvOverrideResource())
	require.Equal(t, "ok", res.Status, res.Reason)

	ids := make([]string, 0, len(res.Records))
	for _, r := range res.Records {
		ids = append(ids, r.ID)
	}
	assert.ElementsMatch(t, []string{"account/15/12", "ou/2/1", "project/9/3"}, ids)
}

// TestCVOverrideSkipsInheritedVariables: the collection lists every variable
// visible at the entity, and only the overridden ones are this resource. On a
// live demo install that was 21 of 1812.
func TestCVOverrideSkipsInheritedVariables(t *testing.T) {
	t.Parallel()
	l := &routeLister{routes: map[string]any{
		"/v3/account":                   []map[string]any{rec("id", float64(2))},
		"/v3/account/2/custom-variable": []map[string]any{inherited(1), inherited(4), overridden(7, "70")},
		"/v3/ou":                        []map[string]any{},
		"/v3/project":                   []map[string]any{},
	}}

	res := Enumerate(context.Background(), l, cvOverrideResource())
	require.Equal(t, "ok", res.Status, res.Reason)
	require.Len(t, res.Records, 1)
	assert.Equal(t, "account/2/7", res.Records[0].ID)
	assert.Contains(t, res.Reason, `2 record(s) in the collection with no "override" set`,
		"the drop is counted and names the discriminator, never silent")
}

// TestCVOverrideIDsAreDistinctAcrossParentKinds: entity ids are per-kind, so
// account 2 and OU 2 both overriding variable 4 are two different records. The
// kind prefix is what keeps their import ids apart -- without it both render
// "2/4" and RenderImports would drop one as a duplicate.
func TestCVOverrideIDsAreDistinctAcrossParentKinds(t *testing.T) {
	t.Parallel()
	l := &routeLister{routes: map[string]any{
		"/v3/account":                   []map[string]any{rec("id", float64(2))},
		"/v3/account/2/custom-variable": []map[string]any{overridden(4, "a")},
		"/v3/ou":                        []map[string]any{rec("id", float64(2))},
		"/v3/ou/2/custom-variable":      []map[string]any{overridden(4, "b")},
		"/v3/project":                   []map[string]any{},
	}}

	res := Enumerate(context.Background(), l, cvOverrideResource())
	require.Equal(t, "ok", res.Status, res.Reason)
	ids := []string{res.Records[0].ID, res.Records[1].ID}
	assert.ElementsMatch(t, []string{"account/2/4", "ou/2/4"}, ids)
}

// TestCVOverrideOneParentKindFailingStillReturnsOthers: the three collections
// are independent reads, so a 500 on accounts must not cost the OU and project
// overrides.
func TestCVOverrideOneParentKindFailingStillReturnsOthers(t *testing.T) {
	t.Parallel()
	l := &routeLister{routes: map[string]any{
		"/v3/account":                   []map[string]any{rec("id", float64(2))},
		"/v3/account/2/custom-variable": errors.New("500 boom"),
		"/v3/ou":                        []map[string]any{rec("id", float64(2))},
		"/v3/ou/2/custom-variable":      []map[string]any{overridden(1, "x")},
		"/v3/project":                   []map[string]any{},
	}}

	res := Enumerate(context.Background(), l, cvOverrideResource())
	require.Equal(t, "ok", res.Status)
	require.Len(t, res.Records, 1)
	assert.Equal(t, "ou/2/1", res.Records[0].ID)
	assert.Contains(t, res.Reason, "500")
	assert.Contains(t, res.Reason, "account")
}

// TestCVOverrideRecordMissingKeyFieldIsCountedNotEmitted: the key field is
// custom_variable_id, not "id"; a record without it cannot produce a resolvable
// import id, and skipping it silently would misreport coverage.
func TestCVOverrideRecordMissingKeyFieldIsCountedNotEmitted(t *testing.T) {
	t.Parallel()
	l := &routeLister{routes: map[string]any{
		"/v3/account": []map[string]any{rec("id", float64(2))},
		"/v3/account/2/custom-variable": []map[string]any{
			{"override": map[string]any{"value": "x"}},
		},
		"/v3/ou":      []map[string]any{},
		"/v3/project": []map[string]any{},
	}}

	res := Enumerate(context.Background(), l, cvOverrideResource())
	assert.Equal(t, "empty", res.Status)
	assert.Empty(t, res.Records)
	assert.Contains(t, res.Reason, "missing key field")
}

// TestParentSegmentOnlyPrefixesTheThreePartFormat guards the blast radius: the
// two-part formats every other parent-scoped resource uses must keep a bare
// parent id.
func TestParentSegmentOnlyPrefixesTheThreePartFormat(t *testing.T) {
	t.Parallel()
	p := importmanifest.Parent{Kind: "ou"}

	kindFormat := importmanifest.Resource{ImportID: importmanifest.ImportID{Format: importmanifest.FormatKindParentSlashKey}}
	assert.Equal(t, "ou/7", parentSegment(kindFormat, p, "7"))

	twoPart := importmanifest.Resource{ImportID: importmanifest.ImportID{Format: importmanifest.FormatParentSlashKey}}
	assert.Equal(t, "7", parentSegment(twoPart, p, "7"))

	plain := importmanifest.Resource{ImportID: importmanifest.ImportID{Format: importmanifest.FormatID}}
	assert.Equal(t, "7", parentSegment(plain, p, "7"))

	// A parent block with no Kind cannot render the prefix; fall back rather
	// than emit a leading slash.
	assert.Equal(t, "7", parentSegment(kindFormat, importmanifest.Parent{}, "7"))
}
