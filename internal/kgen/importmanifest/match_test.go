package importmanifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/kfs"
)

func idmsManifest() *Manifest {
	return Build(
		map[string]string{},
		map[string]string{
			"saml_group_association":   "/v3/idms/{id}/group-association",
			"idms_open_id_access_rule": "/v4/idms/open-id/{id}/access-rule",
		},
		map[string]string{},
		map[string]string{},
		map[string]archetypeInfo{},
		[]string{"kion_idms", "kion_saml_group_association", "kion_idms_open_id_access_rule"},
		map[string]bool{},
	)
}

var idmsAttrs = map[string]map[string]bool{
	"kion_idms": {"id": true, "idms_type_id": true, "name": true},
}

func TestApplyParentMatchesSetsTheParentFilter(t *testing.T) {
	t.Parallel()
	m := idmsManifest()
	require.NoError(t, applyParentMatches(m, map[string]map[string]string{
		"saml_group_association": {"idms_type_id": "3"},
	}, idmsAttrs))

	r := byType(m, "kion_saml_group_association")
	require.NotNil(t, r.Parent)
	assert.Equal(t, map[string]string{"idms_type_id": "3"}, r.Parent.Match)
	other := byType(m, "kion_idms_open_id_access_rule")
	require.NotNil(t, other.Parent)
	assert.Nil(t, other.Parent.Match)
}

func TestApplyParentMatchesFailsLoudly(t *testing.T) {
	t.Parallel()
	cases := map[string]map[string]map[string]string{
		"field the parent does not have": {"saml_group_association": {"idms_kind": "3"}},
		"resource the manifest lacks":    {"no_such_thing": {"idms_type_id": "3"}},
		"resource with no parent":        {"idms": {"idms_type_id": "3"}},
	}
	for name, matches := range cases {
		t.Run(name, func(t *testing.T) {
			err := applyParentMatches(idmsManifest(), matches, idmsAttrs)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "match")
		})
	}
}

// The rule is authored in config_overrides.yaml and reaches the manifest
// through the rendered generator_config.yaml.
func TestLoadDataSourceMatches(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := dir + "/generator_config.yaml"
	require.NoError(t, kfs.OS{}.WriteFile(path, []byte(`
data_sources:
  saml_group_association:
    read:
      path: /v3/idms/{id}/group-association
      method: GET
      match:
        idms_type_id: "3"
  idms:
    read:
      path: /v3/idms
      method: GET
`), 0o600))

	got, err := loadDataSourceMatches(kfs.OS{}, path)
	require.NoError(t, err)
	assert.Equal(t, map[string]map[string]string{"saml_group_association": {"idms_type_id": "3"}}, got)
}
