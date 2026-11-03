package importmanifest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixture() *Manifest {
	return Build(
		map[string]string{ // kind -> by-id read path, from generator_config.yaml resources:
			"ou":                        "/v3/ou/{id}",
			"aws_resource_tag":          "",
			"ou_enforcement":            "/v3/ou/{id}/enforcement",
			"ou_permission_mapping":     "/v3/ou/{id}/permission-mapping",
			"app_config":                "/v3/app-config",
			"global_permission_mapping": "/v3/global/permission-mapping",
			"account":                   "/v4/account-legacy/{id}", // must lose to the data source below
		},
		map[string]string{ // kind -> list read path, from generator_config.yaml data_sources:
			"account": "/v3/account",
		},
		map[string]string{}, // kind -> flat list path, from private_endpoints.yaml list_read_only: (unused by this fixture)
		map[string]string{}, // kind -> by-id read path, from private_endpoints.yaml resources: (unused by this fixture)
		map[string]archetypeInfo{ // kind -> archetype fields, from crud_archetypes.yaml
			"aws_resource_tag": {Kind: "no_read"},
			"ou_enforcement":   {Kind: "parent_list"},
			"ou_permission_mapping": {
				Kind: "association", KeyField: "app_role_id", ParentField: "ou_id",
			},
			"app_config": {Kind: "singleton"},
			"global_permission_mapping": {
				Kind: "association", KeyField: "app_role_id",
			},
		},
		[]string{
			"kion_ou", "kion_aws_resource_tag", "kion_ou_enforcement",
			"kion_ou_permission_mapping", "kion_app_config", "kion_account",
			"kion_global_permission_mapping",
		},
		map[string]bool{}, // tf_type -> has a top-level "name" attribute (unused by this fixture)
	)
}

func byType(m *Manifest, tfType string) Resource {
	for _, r := range m.Resources {
		if r.TFType == tfType {
			return r
		}
	}
	return Resource{}
}

func TestBuildEmitsARowPerTFType(t *testing.T) {
	t.Parallel()
	m := fixture()
	assert.Len(t, m.Resources, 7)
	assert.Equal(t, ManifestVersion, m.Version)
}

func TestBuildSortsByTFType(t *testing.T) {
	t.Parallel()
	m := fixture()
	for i := 1; i < len(m.Resources); i++ {
		assert.Less(t, m.Resources[i-1].TFType, m.Resources[i].TFType)
	}
}

func TestBuildGenericResource(t *testing.T) {
	t.Parallel()
	r := byType(fixture(), "kion_ou")
	assert.Equal(t, ShapeGeneric, r.ReadShape)
	assert.Equal(t, "/v3/ou", r.ListPath)
	assert.Equal(t, FormatID, r.ImportID.Format)
	assert.True(t, r.Readable)
}

func TestBuildNoReadResourceCarriesReasonAndNoPath(t *testing.T) {
	t.Parallel()
	r := byType(fixture(), "kion_aws_resource_tag")
	assert.False(t, r.Readable)
	assert.Empty(t, r.ListPath)
	assert.Contains(t, r.Reason, "no_read")
}

func TestBuildParentListResourceGetsItsParentBlock(t *testing.T) {
	t.Parallel()
	r := byType(fixture(), "kion_ou_enforcement")
	require.NotNil(t, r.Parent)
	assert.Equal(t, "/v3/ou", r.Parent.ListPath)
	assert.Equal(t, "/v3/ou/{parent_id}/enforcement", r.Parent.ChildPath)
	assert.Equal(t, "ou_id", r.Parent.ParentIDField)
	assert.Equal(t, ShapeParentList, r.ReadShape)
	// The read finds the child under its parent, so the import id carries both.
	assert.Equal(t, FormatParentSlashKey, r.ImportID.Format)
	assert.Equal(t, "id", r.ImportID.KeyField)
	assert.True(t, r.Readable)
}

func TestBuildAssociationUsesParentSlashKey(t *testing.T) {
	t.Parallel()
	r := byType(fixture(), "kion_ou_permission_mapping")
	assert.Equal(t, FormatParentSlashKey, r.ImportID.Format)
	assert.Equal(t, "app_role_id", r.ImportID.KeyField)
	require.NotNil(t, r.Parent)
	assert.Equal(t, "ou_id", r.Parent.ParentIDField)
	assert.Equal(t, ShapeAssociation, r.ReadShape)
}

func TestBuildSpecialResourceUsesItsBespokePath(t *testing.T) {
	t.Parallel()
	r := byType(fixture(), "kion_app_config")
	assert.Equal(t, ShapeSpecial, r.ReadShape)
	assert.Equal(t, "/v3/app-config", r.ListPath)
}

// TestBuildDataSourcePathIsPreferredOverResourceReadPath guards the
// resolution order: generator_config.yaml's data_sources: entry is the real
// collection endpoint, and must win over the resources: by-id read even when
// both are present and disagree. kion_account's fixture inputs are set up so
// the two would produce different list paths if this rule were dropped
// (falling back to the by-id read would yield "/v4/account-legacy", not the
// data source's "/v3/account").
func TestBuildDataSourcePathIsPreferredOverResourceReadPath(t *testing.T) {
	t.Parallel()
	r := byType(fixture(), "kion_account")
	assert.Equal(t, "/v3/account", r.ListPath)
	assert.True(t, r.Readable)
}

// TestBuildFallsBackToResourceReadPathWhenNoDataSource guards the other half
// of the resolution order: when generator_config.yaml has no data_sources:
// entry for a kind, the resources: by-id read is used instead of leaving the
// resource unreadable.
func TestBuildFallsBackToResourceReadPathWhenNoDataSource(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{"widget": "/v3/widget/{id}"},
		map[string]string{}, // no data source for "widget"
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{},
		[]string{"kion_widget"},
		map[string]bool{},
	)
	r := byType(m, "kion_widget")
	assert.Equal(t, "/v3/widget", r.ListPath)
	assert.True(t, r.Readable)
}

// TestBuildAliasIsNotEnumerated guards against double-importing. The provider
// serves kion_aws_cloudformation_template and kion_cft from one implementation
// over one endpoint, likewise kion_aws_iam_policy and kion_iam_policy. Giving the
// alias its own enumerable row read the same objects twice: on a real install
// that was 1,348 identical iam-policy ids and 296 cft ids, 1,644 duplicate import
// blocks that would have put two Terraform resources in charge of each record.
//
// The row stays, so every tf_type the provider serves is still accounted for in
// the report, but it is not readable and carries no list path.
func TestBuildAliasIsNotEnumerated(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{
			"cft":        "/v3/cft/{id}",
			"iam_policy": "/v3/iam-policy/{id}",
		},
		map[string]string{
			"cft":        "/v3/cft",
			"iam_policy": "/v3/iam-policy",
		},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{},
		[]string{"kion_aws_cloudformation_template", "kion_aws_iam_policy", "kion_cft"},
		map[string]bool{},
	)

	cft := byType(m, "kion_aws_cloudformation_template")
	assert.Equal(t, "kion_cft", cft.AliasOf)
	assert.False(t, cft.Readable)
	assert.Empty(t, cft.ListPath, "an alias must not carry a list path, or it gets enumerated")
	assert.Equal(t, "aws_cloudformation_template", cft.Kind, "Kind stays the tf_type-derived kind")
	assert.Contains(t, cft.Reason, "kion_cft")

	iamPolicy := byType(m, "kion_aws_iam_policy")
	assert.Equal(t, "kion_iam_policy", iamPolicy.AliasOf)
	assert.False(t, iamPolicy.Readable)

	// The canonical type is unaffected and still enumerable.
	canonical := byType(m, "kion_cft")
	assert.Empty(t, canonical.AliasOf)
	assert.True(t, canonical.Readable)
	assert.Equal(t, "/v3/cft", canonical.ListPath)
}

// TestBuildTwoPlaceholderPathIsUnreadable guards the compound-path veto: a
// chosen path with two or more "{...}" placeholders can't be enumerated from
// either id alone, so it must be marked unreadable even when its archetype
// would otherwise fix it readable and ShapeSpecial.
//
// kion_custom_variable_override used to be the live example, until a plural
// sibling collection turned out to make it enumerable after all (see
// TestBuildCustomVariableOverrideIsRescuedByItsParentCollections). No shipped
// resource hits this branch now, so the case is exercised synthetically -- the
// veto is still the right default for a compound path with no such sibling.
func TestBuildTwoPlaceholderPathIsUnreadable(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{
			"widget_setting": "/v3/widget/{widget_id}/setting/{setting_id}",
		},
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{
			"widget_setting": {Kind: "cv_override"},
		},
		[]string{"kion_widget_setting"},
		map[string]bool{},
	)
	r := byType(m, "kion_widget_setting")
	assert.False(t, r.Readable)
	assert.Equal(t, ShapeNone, r.ReadShape)
	assert.Empty(t, r.ListPath)
	assert.Nil(t, r.Parent)
	assert.Contains(t, r.Reason, "/v3/widget/{widget_id}/setting/{setting_id}")
	// The reason must say the identity is compound, and must NOT claim the
	// API has no read -- it does (a real, public GET), the identity is just
	// two ids instead of one. See generate.go's placeholders >= 2 case.
	assert.Contains(t, r.Reason, "widget_id")
	assert.Contains(t, r.Reason, "setting_id")
	assert.Contains(t, r.Reason, "compound identity")
	assert.NotContains(t, r.Reason, "no read")
}

// TestBuildCustomVariableOverrideIsRescuedByItsParentCollections: derivation
// vetoes the two-placeholder by-id read, and the override tables then have to
// undo every part of that veto -- shape, readability, reason, the three parent
// sets, the discriminator, and the three-part import id. The last is the one
// with no other guard: multiParentOverrides establishes the read but leaves
// ImportID at the plain id Classify returned, which would emit "12" instead of
// "account/15/12" and resolve to nothing.
func TestBuildCustomVariableOverrideIsRescuedByItsParentCollections(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{
			"custom_variable_override": "/v3/account/{account_id}/custom-variable/{custom_variable_id}",
		},
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{
			"custom_variable_override": {Kind: "cv_override"},
		},
		[]string{"kion_custom_variable_override"},
		map[string]bool{},
	)
	r := byType(m, "kion_custom_variable_override")

	assert.True(t, r.Readable)
	assert.Equal(t, ShapeParentList, r.ReadShape)
	assert.Empty(t, r.Reason, "the compound-identity veto must not survive the rescue")
	assert.Equal(t, "override", r.RequireValidField)

	assert.Equal(t, FormatKindParentSlashKey, r.ImportID.Format)
	assert.Equal(t, "custom_variable_id", r.ImportID.KeyField)

	require.Len(t, r.Parents, 3, "the resource is polymorphic across all three entity types")
	kinds := make([]string, 0, len(r.Parents))
	for _, p := range r.Parents {
		kinds = append(kinds, p.Kind)
		assert.Contains(t, p.ChildPath, "{parent_id}")
		assert.Contains(t, p.ChildPath, "/custom-variable")
	}
	// Each Kind becomes the first segment of the import id, so it has to match
	// the entity_type the resource's own CRUD dispatches on.
	assert.Equal(t, []string{"account", "ou", "project"}, kinds)
	require.NotNil(t, r.Parent)
	assert.Equal(t, r.Parents[0], *r.Parent, "Parent must mirror Parents[0]")
}

// TestBuildParentlessAssociationUsesPlainID guards the else-branch of
// association.gtpl's ImportState: a parentless association (no parent_field
// declared, e.g. kion_global_permission_mapping) parses req.ID as a plain
// integer rather than splitting on "/". If Build claimed FormatParentSlashKey
// here, the downstream enumerator would emit zero import blocks, silently.
//
// That same else-branch assigns the parsed integer directly to the KEY
// field, not a synthesized id -- so for a parentless association the import
// id IS the key field's value, and KeyField must still be populated (Format
// stays FormatID) so the enumerator knows which raw field to read.
func TestBuildParentlessAssociationUsesPlainID(t *testing.T) {
	t.Parallel()
	r := byType(fixture(), "kion_global_permission_mapping")
	assert.Equal(t, FormatID, r.ImportID.Format)
	assert.Equal(t, "app_role_id", r.ImportID.KeyField)
	assert.True(t, r.Readable)
	assert.Equal(t, "/v3/global/permission-mapping", r.ListPath)
	assert.Nil(t, r.Parent)
}

// TestBuildKeyFieldComesFromArchetypeData guards against re-hardcoding the
// association key: it must come from crud_archetypes.yaml's key_field, not a
// package constant. A key_field other than "app_role_id" must still surface
// verbatim.
func TestBuildKeyFieldComesFromArchetypeData(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{"widget_permission_mapping": "/v3/widget-permission-mapping/{id}"},
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{
			"widget_permission_mapping": {
				Kind: "association", KeyField: "widget_role_id", ParentField: "widget_id",
			},
		},
		[]string{"kion_widget_permission_mapping"},
		map[string]bool{},
	)
	r := byType(m, "kion_widget_permission_mapping")
	assert.Equal(t, FormatParentSlashKey, r.ImportID.Format)
	assert.Equal(t, "widget_role_id", r.ImportID.KeyField)
}

// TestBuildHasParentComesFromDeclaredParentField guards the split between the
// two concerns Build must not conflate: hasParent (the import-id FORMAT)
// comes from crud_archetypes.yaml's declared parent_field, mirroring how
// internal/kgen/crud/assoc.go derives association.gtpl's HasParent from
// arch.ParentField -- never from whether the chosen read path happens to
// carry a parent-scoping placeholder. KeyField, by contrast, is populated
// for an association either way -- with a parent it is half of
// "<parent>/<key>"; without one, per association.gtpl's {{else}} branch, it
// IS the whole import id.
func TestBuildHasParentComesFromDeclaredParentField(t *testing.T) {
	t.Parallel()
	readPaths := map[string]string{"ou_permission_mapping": "/v3/ou-permission-mapping/{id}"}
	tfTypes := []string{"kion_ou_permission_mapping"}

	withParentField := Build(readPaths, map[string]string{}, map[string]string{}, map[string]string{}, map[string]archetypeInfo{
		"ou_permission_mapping": {Kind: "association", KeyField: "app_role_id", ParentField: "ou_id"},
	}, tfTypes, map[string]bool{})
	r := byType(withParentField, "kion_ou_permission_mapping")
	assert.Equal(t, FormatParentSlashKey, r.ImportID.Format)
	assert.Equal(t, "app_role_id", r.ImportID.KeyField)

	withoutParentField := Build(readPaths, map[string]string{}, map[string]string{}, map[string]string{}, map[string]archetypeInfo{
		"ou_permission_mapping": {Kind: "association", KeyField: "app_role_id"},
	}, tfTypes, map[string]bool{})
	r2 := byType(withoutParentField, "kion_ou_permission_mapping")
	assert.Equal(t, FormatID, r2.ImportID.Format)
	assert.Equal(t, "app_role_id", r2.ImportID.KeyField)
}

// TestBuildParentScopedPathDerivesParentIDField guards the placeholder-based
// Parent derivation for a chosen path with exactly one "{...}" that is NOT
// the path's final segment -- a real nested collection, not a plain by-id
// GET. Each case is a real generator_config.yaml data_sources: entry with a
// different parent_id_field shape.
func TestBuildParentScopedPathDerivesParentIDField(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name               string
		kind               string
		archetype          archetypeInfo
		path               string
		wantParentIDField  string
		wantParentListPath string
		wantChildPath      string
		wantShape          ReadShape
	}{
		{
			name: "ou_note", kind: "ou_note",
			path:               "/v3/ou/{id}/ou-note",
			wantParentIDField:  "ou_id",
			wantParentListPath: "/v3/ou",
			wantChildPath:      "/v3/ou/{parent_id}/ou-note",
			wantShape:          ShapeParentList,
		},
		{
			name: "project_enforcement", kind: "project_enforcement",
			archetype:          archetypeInfo{Kind: "parent_list"},
			path:               "/v3/project/{id}/enforcement",
			wantParentIDField:  "project_id",
			wantParentListPath: "/v3/project",
			wantChildPath:      "/v3/project/{parent_id}/enforcement",
			wantShape:          ShapeParentList,
		},
		{
			name: "idms_group_association", kind: "idms_group_association",
			path:               "/v3/idms/{id}/group-association",
			wantParentIDField:  "idms_id",
			wantParentListPath: "/v3/idms",
			wantChildPath:      "/v3/idms/{parent_id}/group-association",
			wantShape:          ShapeParentList,
		},
		{
			name: "compliance_control", kind: "compliance_control",
			archetype:          archetypeInfo{Kind: "entity"},
			path:               "/v4/compliance/program/{id}/control",
			wantParentIDField:  "compliance_program_id",
			wantParentListPath: "/v4/compliance/program",
			wantChildPath:      "/v4/compliance/program/{parent_id}/control",
			wantShape:          ShapeParentList,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			tfType := "kion_" + c.kind
			m := Build(
				map[string]string{},
				map[string]string{c.kind: c.path},
				map[string]string{},
				map[string]string{},
				map[string]archetypeInfo{c.kind: c.archetype},
				[]string{tfType},
				map[string]bool{},
			)
			r := byType(m, tfType)
			require.NotNil(t, r.Parent, c.name)
			assert.Equal(t, c.wantParentIDField, r.Parent.ParentIDField, c.name)
			assert.Equal(t, c.wantParentListPath, r.Parent.ListPath, c.name)
			assert.Equal(t, c.wantChildPath, r.Parent.ChildPath, c.name)
			assert.Equal(t, c.wantShape, r.ReadShape, c.name)
			assert.True(t, r.Readable, c.name)
		})
	}
}

// TestBuildOverridesOpenIDFamilyParent guards Fix 3: the OpenID family's
// derived parent list (the text before each resource's trailing "/{id}") is
// not itself listable -- /v4/idms/open-id 405s live. parentOverrides must
// replace each of the three resources' derivation with a Parent scoped by
// the IDMS id instead: ListPath "/v3/idms" (the real enumerable list),
// ParentIDField "idms_id", and the live-verified child path.
func TestBuildOverridesOpenIDFamilyParent(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{ // real generator_config.yaml resources: read paths
			"idms_open_id":                   "/v4/idms/open-id/{id}",
			"idms_open_id_access_rule":       "/v4/idms/open-id/access-rule/{id}",
			"idms_open_id_group_association": "/v4/idms/open-id/group-association/{id}",
		},
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{},
		[]string{
			"kion_idms_open_id",
			"kion_idms_open_id_access_rule",
			"kion_idms_open_id_group_association",
		},
		map[string]bool{},
	)

	cases := []struct {
		tfType        string
		wantChildPath string
	}{
		{"kion_idms_open_id", "/v4/idms/open-id/{parent_id}"},
		{"kion_idms_open_id_access_rule", "/v4/idms/open-id/{parent_id}/access-rule"},
		{"kion_idms_open_id_group_association", "/v4/idms/open-id/{parent_id}/group-association"},
	}
	for _, c := range cases {
		r := byType(m, c.tfType)
		require.NotNil(t, r.Parent, c.tfType)
		assert.Equal(t, "/v3/idms", r.Parent.ListPath, c.tfType)
		assert.Equal(t, "idms_id", r.Parent.ParentIDField, c.tfType)
		assert.Equal(t, "idms", r.Parent.Kind, c.tfType)
		assert.Equal(t, c.wantChildPath, r.Parent.ChildPath, c.tfType)
		assert.Equal(t, ShapeParentList, r.ReadShape, c.tfType)
		assert.True(t, r.Readable, c.tfType)
		assert.Empty(t, r.ListPath, c.tfType)
		assert.Nil(t, r.Parents, c.tfType)
	}
}

// TestBuildOverridesCloudAccessRoleExemptionParent guards the two
// *_cloud_access_role_exemption entries added to parentOverrides: both
// resources are crud_archetypes.yaml no_read, and their private
// per-resource read (/v1/ou/{id}/... and /v1/project/{id}/...) would
// otherwise derive an unlistable /v1/ou or /v1/project as the parent list.
// The override must replace that with the real, listable /v3/ou / /v3/project.
func TestBuildOverridesCloudAccessRoleExemptionParent(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]string{ // private_endpoints.yaml resources: read paths
			"ou_cloud_access_role_exemption":      "/v1/ou/{id}/ou-cloud-access-role",
			"project_cloud_access_role_exemption": "/v1/project/{id}/ou-cloud-access-role",
		},
		map[string]archetypeInfo{
			"ou_cloud_access_role_exemption":      {Kind: "no_read"},
			"project_cloud_access_role_exemption": {Kind: "no_read"},
		},
		[]string{
			"kion_ou_cloud_access_role_exemption",
			"kion_project_cloud_access_role_exemption",
		},
		map[string]bool{},
	)

	ou := byType(m, "kion_ou_cloud_access_role_exemption")
	require.NotNil(t, ou.Parent)
	assert.Equal(t, "/v3/ou", ou.Parent.ListPath)
	assert.Equal(t, "/v1/ou/{parent_id}/ou-cloud-access-role", ou.Parent.ChildPath)
	assert.Equal(t, "ou_exemptions", ou.Parent.ChildRecordsKey)
	assert.Empty(t, ou.RequireValidField, "the exemption list holds only this kind")
	assert.Equal(t, "ou_id", ou.Parent.ParentIDField)
	assert.Equal(t, ShapeParentList, ou.ReadShape)
	assert.True(t, ou.Readable)

	project := byType(m, "kion_project_cloud_access_role_exemption")
	require.NotNil(t, project.Parent)
	assert.Equal(t, "/v3/project", project.Parent.ListPath)
	assert.Equal(t, "/v1/project/{parent_id}/ou-cloud-access-role", project.Parent.ChildPath)
	assert.Equal(t, "project_exemptions", project.Parent.ChildRecordsKey)
	assert.Empty(t, project.RequireValidField)
	assert.Equal(t, "project_id", project.Parent.ParentIDField)
	assert.Equal(t, ShapeParentList, project.ReadShape)
	assert.True(t, project.Readable)
}

// TestBuildNoReadArchetypeIsReadableWithAListPath guards Fix 1's core claim:
// crud_archetypes.yaml's no_read means "no by-id GET," not "unreadable."
// kion_aws_resource_tag has no by-id read but IS listable at a flat
// collection endpoint named only in private_endpoints.yaml's list_read_only:
// section -- which generator_config.yaml-only derivation could never see.
func TestBuildNoReadArchetypeIsReadableWithAListPath(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{},
		map[string]string{},
		map[string]string{"aws_resource_tag": "/v3/aws-resource-tag"}, // private_endpoints.yaml list_read_only:
		map[string]string{},
		map[string]archetypeInfo{"aws_resource_tag": {Kind: "no_read"}},
		[]string{"kion_aws_resource_tag"},
		map[string]bool{},
	)
	r := byType(m, "kion_aws_resource_tag")
	assert.True(t, r.Readable)
	assert.Equal(t, ShapeGeneric, r.ReadShape)
	assert.Equal(t, "/v3/aws-resource-tag", r.ListPath)
	assert.Empty(t, r.Reason)
}

// TestBuildPrivateListReadOnlyPathBeatsGeneratorConfigPaths guards the
// resolution order's top slot: list_read_only names a real, already-flat
// collection outright, so it must win even when generator_config.yaml also
// has a (by-id) path for the same kind.
func TestBuildPrivateListReadOnlyPathBeatsGeneratorConfigPaths(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{"widget": "/v3/widget/{id}"},
		map[string]string{"widget": "/v3/widget-list"},
		map[string]string{"widget": "/v3/widget-flat-list"},
		map[string]string{},
		map[string]archetypeInfo{},
		[]string{"kion_widget"},
		map[string]bool{},
	)
	r := byType(m, "kion_widget")
	assert.Equal(t, "/v3/widget-flat-list", r.ListPath)
	assert.True(t, r.Readable)
}

// TestBuildPrivateResourceReadPathIsLastResort guards the resolution order's
// bottom slot: private_endpoints.yaml's resources: read path is used only
// when generator_config.yaml has nothing at all for that kind. Uses a
// fictitious kind (not kion_dashboard, which explicitFlatListOverrides would
// otherwise mask) so the fallback chain itself is what's under test.
func TestBuildPrivateResourceReadPathIsLastResort(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{},                            // nothing in generator_config.yaml resources:
		map[string]string{},                            // nothing in generator_config.yaml data_sources:
		map[string]string{},                            // not list-read-only
		map[string]string{"widget": "/v1/widget/{id}"}, // private_endpoints.yaml resources:
		map[string]archetypeInfo{"widget": {Kind: "raw_http"}},
		[]string{"kion_widget"},
		map[string]bool{},
	)
	r := byType(m, "kion_widget")
	assert.True(t, r.Readable)
	assert.Equal(t, "/v1/widget", r.ListPath)
}

// TestBuildExplicitParentOverrideForFundingSourceNote guards Fix 2's
// parent-scoped half: kion_funding_source_note's real read
// (/v2/funding-source/{id}/funding-source-note) is recorded in neither
// generator_config.yaml nor private_endpoints.yaml, so explicitParentOverrides
// is the only thing that can produce it.
func TestBuildExplicitParentOverrideForFundingSourceNote(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{"funding_source_note": {Kind: "raw_http"}},
		[]string{"kion_funding_source_note"},
		map[string]bool{},
	)
	r := byType(m, "kion_funding_source_note")
	require.NotNil(t, r.Parent)
	assert.Equal(t, "/v3/funding-source", r.Parent.ListPath)
	assert.Equal(t, "/v2/funding-source/{parent_id}/funding-source-note", r.Parent.ChildPath)
	assert.Equal(t, "funding_source_id", r.Parent.ParentIDField)
	assert.Equal(t, ShapeParentList, r.ReadShape)
	assert.True(t, r.Readable)
	assert.Empty(t, r.ListPath)
}

// TestBuildExplicitFlatListOverrideForDashboard guards Fix 2's flat-list
// half: kion_dashboard's real plural collection, /v1/dashboards, is recorded
// in neither generator_config.yaml nor private_endpoints.yaml (which only
// records the private by-id read), so explicitFlatListOverrides is the only
// thing that can produce it.
func TestBuildExplicitFlatListOverrideForDashboard(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]string{"dashboard": "/v1/dashboard/{id}"}, // present, but must lose to the override
		map[string]archetypeInfo{"dashboard": {Kind: "raw_http"}},
		[]string{"kion_dashboard"},
		map[string]bool{},
	)
	r := byType(m, "kion_dashboard")
	assert.Equal(t, "/v1/dashboards", r.ListPath)
	assert.Equal(t, ShapeGeneric, r.ReadShape)
	assert.True(t, r.Readable)
	assert.Nil(t, r.Parent)
}

// TestBuildOverridesBudgetWithTwoParents guards Fix 4: /v3/budget 405s live
// (there is no flat list), and budgets hang off two parents, both verified
// live -- /v3/ou/{id}/budget and /v3/project/{id}/budget. multiParentOverrides
// must populate Parents with both, and Parent must still mirror Parents[0]
// so a reader that only knows about the single-Parent field still works.
func TestBuildOverridesBudgetWithTwoParents(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{"budget": "/v3/budget/{id}"},
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{},
		[]string{"kion_budget"},
		map[string]bool{},
	)
	r := byType(m, "kion_budget")

	require.Len(t, r.Parents, 2)
	require.NotNil(t, r.Parent)
	assert.Equal(t, *r.Parent, r.Parents[0], "Parent must mirror Parents[0]")

	assert.Equal(t, "ou", r.Parents[0].Kind)
	assert.Equal(t, "/v3/ou", r.Parents[0].ListPath)
	assert.Equal(t, "/v3/ou/{parent_id}/budget", r.Parents[0].ChildPath)
	assert.Equal(t, "ou_id", r.Parents[0].ParentIDField)

	assert.Equal(t, "project", r.Parents[1].Kind)
	assert.Equal(t, "/v3/project", r.Parents[1].ListPath)
	assert.Equal(t, "/v3/project/{parent_id}/budget", r.Parents[1].ChildPath)
	assert.Equal(t, "project_id", r.Parents[1].ParentIDField)

	assert.Equal(t, ShapeParentList, r.ReadShape)
	assert.True(t, r.Readable)
	assert.Empty(t, r.ListPath)
}

// TestBuildNameFieldOnlyWhenSchemaHasNameAttribute guards Fix 5: NameField
// must come from the schema snapshot's declared attributes, not a package
// constant -- a resource without a top-level "name" attribute (like
// kion_aws_resource_tag) must get an empty NameField, not "name".
func TestBuildNameFieldOnlyWhenSchemaHasNameAttribute(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{
			"ou":               "/v3/ou/{id}",
			"aws_resource_tag": "/v3/aws-resource-tag",
		},
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{},
		[]string{"kion_ou", "kion_aws_resource_tag"},
		map[string]bool{"kion_ou": true}, // kion_aws_resource_tag deliberately absent
	)
	ou := byType(m, "kion_ou")
	assert.Equal(t, "name", ou.NameField)

	tag := byType(m, "kion_aws_resource_tag")
	assert.Empty(t, tag.NameField)
}

// TestBuildCompoundKeyParentReadProducesNestedCollectionRow guards
// kion_scope_criteria's real shape: generator_config.yaml's data_sources:
// section (checked before resources:) records scope_criteria's read as
// "/beta/scope" with NO placeholder -- the same flat list scope itself uses --
// so this exercises Build's 0-placeholder default branch, not the
// parent-scoped placeholder branch ShapeParentList goes through. Collection
// must come out converted to its JSON key ("criteria_records"), not the
// Go-style name crud_archetypes.yaml declares ("CriteriaRecords").
func TestBuildCompoundKeyParentReadProducesNestedCollectionRow(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{"scope_criteria": "/beta/scope/{id}"}, // resources: read path (loses to data_sources: below)
		map[string]string{"scope_criteria": "/beta/scope"},      // data_sources: read path
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{
			"scope_criteria": {
				Kind:          "compound_key_parent_read",
				Collection:    "CriteriaRecords",
				ParentIDField: "scope_id",
				ChildIDField:  "criteria_id",
			},
		},
		[]string{"kion_scope_criteria"},
		map[string]bool{},
	)
	r := byType(m, "kion_scope_criteria")
	assert.Equal(t, ShapeNestedCollection, r.ReadShape)
	assert.Equal(t, "/beta/scope", r.ListPath)
	assert.Equal(t, "criteria_records", r.Collection)
	assert.Equal(t, "scope_id", r.ParentIDField)
	assert.Equal(t, "criteria_id", r.ChildIDField)
	assert.Equal(t, FormatParentSlashKey, r.ImportID.Format)
	assert.True(t, r.Readable)
	assert.Nil(t, r.Parent, "ShapeNestedCollection has no second-endpoint Parent block")
}

// TestJSONKeyFromGoName guards the Go-style-name -> JSON-key fold Build
// applies to compound_key_parent_read's Collection field. Its doc comment
// already names the acronym limitation ("ID" -> "i_d") as intentional; this
// test locks in the ordinary cases plus that documented edge case, so a
// change to the fold's behavior (acronym-aware or not) is a deliberate
// decision, not an accident.
func TestJSONKeyFromGoName(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		{"CriteriaRecords", "criteria_records"},
		{"Name", "name"},
		{"", ""},
		{"a", "a"},
		{"ID", "i_d"}, // documented limitation: not acronym-aware
	}
	for _, c := range cases {
		assert.Equal(t, c.want, jsonKeyFromGoName(c.in), c.in)
	}
}

func TestManifestMarshalsDeterministically(t *testing.T) {
	t.Parallel()
	a, err := json.MarshalIndent(fixture(), "", "  ")
	require.NoError(t, err)
	b, err := json.MarshalIndent(fixture(), "", "  ")
	require.NoError(t, err)
	assert.Equal(t, string(a), string(b))
}

// TestBuildParentFallbackKeepsPlainID guards the distinction between reaching a
// record through a parent and having an ImportState that splits on "/".
//
// compliance_family and compliance_level declare no archetype, so they get
// entity.gtpl's id-only ImportState, and only fall back to a parent-scoped read
// when their flat list 405s. Giving them a compound id made Read parse
// "12/284" as an integer: 652 failures on a real install.
func TestBuildParentFallbackKeepsPlainID(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{"compliance_family": "/v3/compliance/family/{id}"},
		map[string]string{"compliance_family": "/v3/compliance/family/{id}"},
		nil, nil,
		map[string]archetypeInfo{}, // no archetype: defaults to entity
		[]string{"kion_compliance_family"},
		nil,
	)
	r := byType(m, "kion_compliance_family")
	assert.Equal(t, FormatID, r.ImportID.Format,
		"an entity that merely falls back to a parent-scoped read keeps a plain id")

	// A declared parent_list still gets the compound id its ImportState parses.
	m2 := Build(
		map[string]string{"ou_enforcement": "/v3/ou/{id}/enforcement"},
		map[string]string{"ou_enforcement": "/v3/ou/{id}/enforcement"},
		nil, nil,
		map[string]archetypeInfo{"ou_enforcement": {Kind: "parent_list", ParentField: "ou_id"}},
		[]string{"kion_ou_enforcement"},
		nil,
	)
	r2 := byType(m2, "kion_ou_enforcement")
	assert.Equal(t, FormatParentSlashKey, r2.ImportID.Format)
	assert.Equal(t, "id", r2.ImportID.KeyField)
}

// TestBuildOverridesGovcloudParent: the derived parent list for GovCloud,
// /v3/billing-source, only accepts POST (405 on GET). /v4/billing-source lists.
func TestBuildOverridesGovcloudParent(t *testing.T) {
	t.Parallel()
	m := Build(
		map[string]string{"billing_source_govcloud": "/v3/billing-source/{id}/govcloud"},
		map[string]string{},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{},
		[]string{"kion_billing_source_govcloud"},
		map[string]bool{},
	)

	r := byType(m, "kion_billing_source_govcloud")
	require.NotNil(t, r.Parent)
	assert.Equal(t, "/v4/billing-source", r.Parent.ListPath)
	assert.Equal(t, "/v3/billing-source/{parent_id}/govcloud", r.Parent.ChildPath)
	assert.Equal(t, "billing_source_id", r.Parent.ParentIDField)
	assert.Equal(t, ShapeParentList, r.ReadShape)
}

// TestBuildDiscriminatesBillingSourceKinds: /v4/billing-source lists every
// payer type, so each typed billing source must keep only its own kind, or one
// AWS payer imports as an AWS, a GCP and an OCI billing source at once.
func TestBuildDiscriminatesBillingSourceKinds(t *testing.T) {
	t.Parallel()
	want := map[string]string{
		"kion_billing_source":           "custom_billing_source",
		"kion_billing_source_aws":       "aws_payer",
		"kion_billing_source_gcp":       "gcp_payer",
		"kion_billing_source_oci":       "oci_payer",
		"kion_billing_source_azure":     "azure_payer",
		"kion_billing_source_anthropic": "anthropic_billing_source",
	}
	read := map[string]string{}
	var types []string
	for tf := range want {
		read[strings.TrimPrefix(tf, "kion_")] = "/v4/billing-source/{id}"
		types = append(types, tf)
	}
	m := Build(read, map[string]string{}, map[string]string{}, map[string]string{},
		map[string]archetypeInfo{}, types, map[string]bool{})
	for tf, field := range want {
		assert.Equal(t, field, byType(m, tf).RequireValidField, tf)
	}
}

// kion_billing_source_azure is read through the polymorphic /v4/billing-source
// like the other typed billing sources; its generator_config read path is a
// schema-only key, not a route, so it must not derive a parent read from it.
func TestBuildAzureBillingSourceListsV4(t *testing.T) {
	t.Parallel()
	m := Build(map[string]string{"billing_source_azure": "/v1/payer/{id}/azure"}, map[string]string{}, map[string]string{}, map[string]string{},
		map[string]archetypeInfo{}, []string{"kion_billing_source_azure"}, map[string]bool{})
	r := byType(m, "kion_billing_source_azure")
	assert.Equal(t, "/v4/billing-source", r.ListPath)
	assert.Nil(t, r.Parent)
	assert.Equal(t, ShapeGeneric, r.ReadShape)
	assert.Equal(t, "azure_payer", r.RequireValidField)
}
