package importmodules_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/importmodules"
)

// fixtureManifest is a small, self-contained manifest covering every shape
// the tests below need: a plain module (kion_ou), the two known
// module-block-meta-argument collisions (kion_cloud_rule's `source`,
// kion_compliance_program's `version`), and a block-typed variable
// (kion_fake_block's `settings`, standing in for any resource whose schema
// renders an attribute as a nested block). It is built in Go rather than
// loaded from modules/module_manifest.json so these tests do not drift with
// the real provider's schema -- they pin the rewriter's behavior against a
// manifest shape, not against whatever attributes kion_ou happens to have
// this week.
func fixtureManifest() *importmodules.Manifest {
	return &importmodules.Manifest{
		Version: 1,
		Modules: []importmodules.Module{
			{
				TFType: "kion_ou",
				Path:   "terraform-kion-ou",
				Variables: []importmodules.Variable{
					{Attr: "name", Var: "name"},
					{Attr: "description", Var: "description"},
					{Attr: "parent_ou_id", Var: "parent_ou_id"},
				},
			},
			{
				TFType: "kion_cloud_rule",
				Path:   "terraform-kion-cloud-rule",
				Variables: []importmodules.Variable{
					{Attr: "name", Var: "name"},
					{Attr: "source", Var: "cloud_rule_source"},
				},
			},
			{
				TFType: "kion_compliance_program",
				Path:   "terraform-kion-compliance-program",
				Variables: []importmodules.Variable{
					{Attr: "name", Var: "name"},
					{Attr: "version", Var: "compliance_program_version"},
				},
			},
			{
				TFType: "kion_fake_block",
				Path:   "terraform-kion-fake-block",
				Variables: []importmodules.Variable{
					{Attr: "name", Var: "name"},
					{Attr: "settings", Var: "settings", Block: true},
				},
			},
		},
	}
}

func norm(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// countAttr counts assignments to exactly the attribute named name. A plain
// strings.Count would over-count: the renamed variables these tests exist to
// check -- cloud_rule_source, compliance_program_version -- end with the very
// name whose absence is being asserted, so `strings.Count(got, "source =")`
// sees both the module's own source and the renamed one. Anchoring on a
// non-identifier character to the left is what makes "did the original
// attribute survive under its original name?" an answerable question.
func countAttr(s, name string) int {
	re := regexp.MustCompile(`(^|[^0-9A-Za-z_])` + regexp.QuoteMeta(name) + ` =`)
	return len(re.FindAllString(s, -1))
}

// TestRewrite_simple: a plain resource block with no name collisions becomes
// a module call with the module's source and every attribute mapped through
// to its variable, unchanged in value.
func TestRewrite_simple(t *testing.T) {
	t.Parallel()
	src := `resource "kion_ou" "example" {
  name        = "engineering"
  description = "eng org unit"
}
`
	out, warnings, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	assert.Empty(t, warnings)

	got := norm(string(out))
	assert.Contains(t, got, `module "example" {`)
	assert.Contains(t, got, `source = "./modules/terraform-kion-ou"`)
	assert.Contains(t, got, `name = "engineering"`)
	assert.Contains(t, got, `description = "eng org unit"`)
	assert.NotContains(t, got, `resource "kion_ou"`)
}

// TestRewrite_cloudRuleSourceCollision locks in the one case a naive rewrite
// gets wrong: kion_cloud_rule has its own `source` attribute, which must NOT
// land in the module block as `source = "aws"` -- Terraform would read that
// as the module's own source address, not a resource attribute. It must be
// renamed to cloud_rule_source, and the module's `source` must still be the
// module path.
func TestRewrite_cloudRuleSourceCollision(t *testing.T) {
	t.Parallel()
	src := `resource "kion_cloud_rule" "example" {
  name   = "my-rule"
  source = "aws"
}
`
	out, warnings, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	assert.Empty(t, warnings)

	got := norm(string(out))
	assert.Contains(t, got, `source = "./modules/terraform-kion-cloud-rule"`)
	assert.Contains(t, got, `cloud_rule_source = "aws"`)
	// The original resource's `source = "aws"` must not survive as a second,
	// conflicting `source` attribute: the only bare `source` is the module's.
	assert.Equal(t, 1, countAttr(got, "source"))
	assert.Equal(t, 1, countAttr(got, "cloud_rule_source"))
}

// TestRewrite_complianceProgramVersionCollision is the same shape as the
// cloud_rule case for kion_compliance_program's `version` attribute, which
// collides with the module block's own `version` meta-argument.
func TestRewrite_complianceProgramVersionCollision(t *testing.T) {
	t.Parallel()
	src := `resource "kion_compliance_program" "example" {
  name    = "SOC 2"
  version = "2024"
}
`
	out, warnings, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	assert.Empty(t, warnings)

	got := norm(string(out))
	assert.Contains(t, got, `source = "./modules/terraform-kion-compliance-program"`)
	assert.Contains(t, got, `compliance_program_version = "2024"`)
	assert.Equal(t, 0, countAttr(got, "version"))
	assert.Equal(t, 1, countAttr(got, "compliance_program_version"))
}

// TestRewrite_dropsComputedOnlyAttribute: an attribute the manifest has no
// entry for at all (a computed-only field like last_updated -- excluded from
// the manifest because Generate never turns it into a module input) is
// dropped from the module call and reported as a Warning rather than
// silently carried over or left to fail a later `terraform validate`.
func TestRewrite_dropsComputedOnlyAttribute(t *testing.T) {
	t.Parallel()
	src := `resource "kion_ou" "example" {
  name         = "engineering"
  last_updated = "2024-01-01T00:00:00Z"
}
`
	out, warnings, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)

	require.Len(t, warnings, 1)
	assert.Equal(t, importmodules.Warning{Resource: "kion_ou", Label: "example", Attr: "last_updated"}, warnings[0])

	got := norm(string(out))
	assert.Contains(t, got, `name = "engineering"`)
	assert.NotContains(t, got, "last_updated")
}

// TestRewrite_leavesUnknownResourceTypeUntouched: a resource block whose type
// has no manifest entry (no generated module for it) is left exactly as
// written -- still a `resource` block, never turned into a `module` call --
// so a rewrite of a mixed file cannot silently drop something it doesn't
// know how to handle.
func TestRewrite_leavesUnknownResourceTypeUntouched(t *testing.T) {
	t.Parallel()
	src := `resource "kion_not_in_manifest" "example" {
  name = "untouched"
}
`
	out, warnings, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	assert.Empty(t, warnings)

	got := norm(string(out))
	assert.Contains(t, got, `resource "kion_not_in_manifest" "example" {`)
	assert.Contains(t, got, `name = "untouched"`)
	assert.NotContains(t, got, "module ")
}

// TestRewrite_nestedBlockToObject: -generate-config-out renders a
// block-typed schema attribute as a literal nested block, not a `dynamic`
// block or an object-valued attribute. Rewrite must fold that block's own
// attributes into the object literal the module's variable expects --
// mirroring how the module's own main.tf unwraps that variable via `dynamic
// { for_each = var.x == null ? [] : [var.x]; content { ... } }`.
func TestRewrite_nestedBlockToObject(t *testing.T) {
	t.Parallel()
	src := `resource "kion_fake_block" "example" {
  name = "example"
  settings {
    enabled = true
    level   = "high"
  }
}
`
	out, warnings, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	assert.Empty(t, warnings)

	got := norm(string(out))
	assert.Contains(t, got, `settings = { enabled = true level = "high" }`)
	assert.NotContains(t, got, "settings {")
}

// TestRewrite_outputReparses: the rewritten output must itself be valid HCL
// -- a rewrite that produces something Terraform cannot parse is worse than
// no rewrite at all.
func TestRewrite_outputReparses(t *testing.T) {
	t.Parallel()
	src := `resource "kion_ou" "example" {
  name        = "engineering"
  description = "eng org unit"
}

resource "kion_cloud_rule" "example" {
  name   = "my-rule"
  source = "aws"
}

resource "kion_not_in_manifest" "example" {
  name = "untouched"
}
`
	out, warnings, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	_ = warnings

	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL(out, "rewritten.tf")
	assert.False(t, diags.HasErrors(), "rewritten output does not parse: %s", diags.Error())
	assert.Empty(t, []*hcl.Diagnostic(diags))
}

// TestRewrite_deterministic: rewriting the same input twice must produce
// byte-identical output -- required for `kgen import-modules` to be safely
// re-runnable and for CI to diff its output meaningfully.
func TestRewrite_deterministic(t *testing.T) {
	t.Parallel()
	src := `resource "kion_ou" "a" {
  name = "a"
}

resource "kion_cloud_rule" "b" {
  name   = "b"
  source = "aws"
}

resource "kion_not_in_manifest" "c" {
  name = "c"
}
`
	out1, _, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	out2, _, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)

	assert.Equal(t, string(out1), string(out2))
}

// TestCountKnownBlocks reports the same rewritten/untouched split Rewrite
// itself would produce, used by the CLI's summary line.
func TestCountKnownBlocks(t *testing.T) {
	t.Parallel()
	src := `resource "kion_ou" "a" {
  name = "a"
}

resource "kion_not_in_manifest" "b" {
  name = "b"
}
`
	rewritten, untouched, err := importmodules.CountKnownBlocks([]byte(src), fixtureManifest())
	require.NoError(t, err)
	assert.Equal(t, 1, rewritten)
	assert.Equal(t, 1, untouched)
}
