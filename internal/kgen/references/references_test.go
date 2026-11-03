package references_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"terraform-provider-kion/codegen"
	"terraform-provider-kion/internal/kgen/references"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func load(t *testing.T) *references.Map {
	t.Helper()
	m, err := references.Parse(codegen.ReferencesYAML)
	require.NoError(t, err)
	return m
}

// providerSchema is the shape of codegen/schema_snapshots/new.json that this
// test needs: every resource, and every attribute's optional/required flags.
type providerSchema struct {
	ProviderSchemas map[string]struct {
		ResourceSchemas map[string]struct {
			Block struct {
				Attributes map[string]struct {
					Optional bool `json:"optional"`
					Required bool `json:"required"`
					Computed bool `json:"computed"`
				} `json:"attributes"`
			} `json:"block"`
		} `json:"resource_schemas"`
	} `json:"provider_schemas"`
}

func snapshot(t *testing.T) providerSchema {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "codegen", "schema_snapshots", "new.json"))
	require.NoError(t, err, "schema snapshot must be committed")
	var doc providerSchema
	require.NoError(t, json.Unmarshal(raw, &doc))
	return doc
}

// TestReferencesAreComplete is what keeps references.yaml honest as the
// provider grows. Every settable attribute whose name ends in _id or _ids must
// be classified as either a reference or explicitly not one. A new resource
// with a new foreign key fails here rather than silently going un-rewritten,
// which would leave a literal id in a generated configuration with nothing to
// say so.
//
// Only SETTABLE attributes are required: a computed-only attribute never
// appears in generated configuration, so there is nothing to rewrite.
func TestReferencesAreComplete(t *testing.T) {
	m := load(t)
	doc := snapshot(t)

	unclassified := map[string][]string{} // attr -> resources carrying it
	for _, provider := range doc.ProviderSchemas {
		for tfType, res := range provider.ResourceSchemas {
			for attr, meta := range res.Block.Attributes {
				if !references.IsForeignKeyShaped(attr) {
					continue
				}
				if !meta.Optional && !meta.Required {
					continue // computed-only; never written to config
				}
				if m.Classified(attr) {
					continue
				}
				unclassified[attr] = append(unclassified[attr], tfType)
			}
		}
	}

	if len(unclassified) == 0 {
		return
	}
	attrs := make([]string, 0, len(unclassified))
	for a := range unclassified {
		attrs = append(attrs, a)
	}
	sort.Strings(attrs)
	var b strings.Builder
	b.WriteString("unclassified foreign-key-shaped attributes; add each to codegen/references.yaml\n")
	b.WriteString("under `references` (with its target tf_type) or `not_references` (with the reason):\n")
	for _, a := range attrs {
		rs := unclassified[a]
		sort.Strings(rs)
		b.WriteString("  " + a + "  on: " + strings.Join(rs, ", ") + "\n")
	}
	t.Fatal(b.String())
}

// TestEveryTargetIsARealResource catches a typo in a target: a reference
// pointing at a tf_type the provider does not serve would produce a rewrite to
// a resource block that cannot exist.
func TestEveryTargetIsARealResource(t *testing.T) {
	m := load(t)
	doc := snapshot(t)

	known := map[string]bool{}
	for _, provider := range doc.ProviderSchemas {
		for tfType := range provider.ResourceSchemas {
			known[tfType] = true
		}
	}
	require.NotEmpty(t, known)

	for _, attr := range m.Attrs() {
		target, ok := m.Target("", attr)
		require.True(t, ok)
		assert.True(t, known[target], "%s targets %q, which is not a resource this provider serves", attr, target)
	}
}

// TestKnownTraps pins the classifications that a name-based guess gets wrong.
// These are the reason the table is authored rather than derived.
func TestKnownTraps(t *testing.T) {
	m := load(t)

	for attr, want := range map[string]string{
		"payer_id":               "kion_billing_source",
		"ugroup_ids":             "kion_user_group",
		"user_groups_ids":        "kion_user_group",
		"internal_portfolio_ids": "kion_service_catalog",
		"internal_ami_ids":       "kion_ami",
		"program_id":             "kion_compliance_program",
		"open_id_id":             "kion_idms_open_id",
	} {
		got, ok := m.Target("", attr)
		assert.True(t, ok, "%s should be a reference", attr)
		assert.Equal(t, want, got, "%s target", attr)
	}

	// Cloud-provider and external ids must never be rewritten: they name
	// records in a foreign id space. portfolio_id is the sharpest case -- it
	// sits beside internal_portfolio_ids, which IS a kion_service_catalog.
	for _, attr := range []string{
		"portfolio_id", "aws_ami_id", "azure_object_id", "google_cloud_project_id",
		"car_external_id", "service_external_id", "account_type_id", "entity_id",
	} {
		_, ok := m.Target("", attr)
		assert.False(t, ok, "%s must not be treated as a reference", attr)
		assert.True(t, m.Classified(attr), "%s must still be classified", attr)
	}
}

func TestParseRejectsContradictionAndBadTarget(t *testing.T) {
	_, err := references.Parse([]byte("references:\n  ou_id: kion_ou\nnot_references:\n  ou_id: nope\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "both references and not_references")

	_, err = references.Parse([]byte("references:\n  ou_id: ou\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a tf_type")
}

func TestOverrideBeatsGlobal(t *testing.T) {
	m, err := references.Parse([]byte(
		"references:\n  ou_id: kion_ou\noverrides:\n  kion_weird:\n    ou_id: kion_project\n"))
	require.NoError(t, err)

	got, ok := m.Target("kion_weird", "ou_id")
	assert.True(t, ok)
	assert.Equal(t, "kion_project", got)

	got, ok = m.Target("kion_project", "ou_id")
	assert.True(t, ok)
	assert.Equal(t, "kion_ou", got, "other resources keep the global mapping")
}

func TestIsForeignKeyShaped(t *testing.T) {
	assert.True(t, references.IsForeignKeyShaped("ou_id"))
	assert.True(t, references.IsForeignKeyShaped("owner_user_ids"))
	assert.False(t, references.IsForeignKeyShaped("id"), "a resource's own identity is not a reference")
	assert.False(t, references.IsForeignKeyShaped("name"))
	assert.False(t, references.IsForeignKeyShaped("valid"))
}

// A family lists the tf_types sharing one id space, so a reference resolves
// against whichever of them holds the id. payer_id is the case: every billing
// source type is keyed on the payer id.
func TestFamilyMembers(t *testing.T) {
	m, err := references.Parse([]byte(
		"references:\n  payer_id: kion_billing_source\nfamilies:\n  kion_billing_source: [kion_billing_source, kion_billing_source_aws]\n"))
	require.NoError(t, err)
	assert.Equal(t, []string{"kion_billing_source", "kion_billing_source_aws"}, m.Members("kion_billing_source"))
	assert.Equal(t, []string{"kion_ou"}, m.Members("kion_ou"), "a type with no family is its own only member")

	_, err = references.Parse([]byte("families:\n  kion_billing_source: [billing_source_aws]\n"))
	assert.Error(t, err, "family members must be tf_types")
}

// Every family member must be a resource this provider serves, like targets.
func TestEveryFamilyMemberIsARealResource(t *testing.T) {
	m := load(t)
	doc := snapshot(t)
	known := map[string]bool{}
	for _, provider := range doc.ProviderSchemas {
		for tfType := range provider.ResourceSchemas {
			known[tfType] = true
		}
	}
	for _, attr := range m.Attrs() {
		target, _ := m.Target("", attr)
		for _, member := range m.Members(target) {
			assert.True(t, known[member], "%s family member %q is not a resource this provider serves", target, member)
		}
	}
}
