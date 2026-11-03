package importmodules_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/importmodules"
)

// `.id` is the shape kion-import rewrite-refs emits, and Rewrite retargets it
// to module.<label>.id, so it is NOT dangling. This is the case that matters:
// references, not literals, are what a configuration should end up holding.
func TestFindDanglingRefs_IDIsRetargetedNotDangling(t *testing.T) {
	t.Parallel()
	src := `resource "kion_ou" "parent" {
  name = "eng"
}

resource "kion_ou" "child" {
  name         = "sub"
  parent_ou_id = kion_ou.parent.id
}
`
	got, err := importmodules.FindDanglingRefs([]byte(src), fixtureManifest())
	require.NoError(t, err)
	assert.Empty(t, got, ".id references are retargeted, not reported")
}

// Reaching into any other attribute of a converted resource cannot be
// retargeted: a module exposes outputs, not attributes, and there may be no
// output of that name. Reported rather than guessed at.
func TestFindDanglingRefs_ReportsNonIDTraversal(t *testing.T) {
	t.Parallel()
	src := `resource "kion_ou" "parent" {
  name = "eng"
}

resource "kion_ou" "child" {
  name         = kion_ou.parent.name
  parent_ou_id = 1
}
`
	got, err := importmodules.FindDanglingRefs([]byte(src), fixtureManifest())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "name", got[0].Attr)
	assert.Equal(t, "kion_ou.parent", got[0].Target)
	assert.Contains(t, got[0].From, "child")
}

// A reference to a type with NO manifest entry is fine: that block is left
// exactly where it was, so the reference still resolves.
func TestFindDanglingRefs_IgnoresUnrewrittenTarget(t *testing.T) {
	t.Parallel()
	src := `resource "not_kion_thing" "other" {
  a = 1
}

resource "kion_ou" "child" {
  name         = "sub"
  parent_ou_id = not_kion_thing.other.id
}
`
	got, err := importmodules.FindDanglingRefs([]byte(src), fixtureManifest())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestFindDanglingRefs_NoneWhenNothingRewritten(t *testing.T) {
	t.Parallel()
	src := `resource "not_kion_thing" "a" {
  x = not_kion_thing.b.id
}
`
	got, err := importmodules.FindDanglingRefs([]byte(src), fixtureManifest())
	require.NoError(t, err)
	assert.Empty(t, got)
}

// TestRewrite_RetargetsReferences is the behavior that makes this compose with
// `kion-import rewrite-refs`: a reference to a converted resource has to point
// at the module, from converted and unconverted blocks alike.
func TestRewrite_RetargetsReferences(t *testing.T) {
	t.Parallel()
	src := `resource "kion_ou" "kion_ou_11" {
  name = "eng"
}

resource "kion_ou" "kion_ou_12" {
  name         = "sub"
  parent_ou_id = kion_ou.kion_ou_11.id
}

resource "not_kion_thing" "keep" {
  a = kion_ou.kion_ou_11.id
}
`
	out, warnings, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	assert.Empty(t, warnings)

	got := norm(string(out))
	assert.Contains(t, got, "parent_ou_id = module.kion_ou_11.id")
	// The unconverted block references the same resource; leaving it pointing
	// at kion_ou.kion_ou_11 would emit a file Terraform refuses to load.
	assert.Contains(t, got, "a = module.kion_ou_11.id")
	assert.NotContains(t, got, "kion_ou.kion_ou_11.id")
	// ...but the block itself is otherwise untouched.
	assert.Contains(t, got, `resource "not_kion_thing" "keep"`)
}

// A resource label is unique per type, a module label is unique across the
// configuration. kion_cloud_rule.soc_2 and kion_compliance_program.soc_2 are
// both legal and both occur on a real install -- a cloud rule and the
// compliance program it enforces share a name. Mapping both to module "soc_2"
// produced "Duplicate module call" and Terraform refused the file; 29 labels
// collided in one 588-block import.
func TestRewrite_DisambiguatesCollidingLabels(t *testing.T) {
	t.Parallel()
	src := `resource "kion_cloud_rule" "soc_2" {
  name = "rule"
}

resource "kion_compliance_program" "soc_2" {
  name = "program"
}

resource "kion_ou" "unique" {
  name = "eng"
}
`
	out, _, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	got := norm(string(out))

	assert.Contains(t, got, `module "cloud_rule_soc_2"`)
	assert.Contains(t, got, `module "compliance_program_soc_2"`)
	// A label that does not collide keeps its name; qualifying everything
	// would rename 559 blocks to fix 29.
	assert.Contains(t, got, `module "unique"`)
	assert.NotContains(t, got, `module "soc_2"`)
}

// A reference into a resource whose label was qualified must follow it, or the
// retarget points at a module that does not exist.
func TestRewrite_ReferenceFollowsQualifiedLabel(t *testing.T) {
	t.Parallel()
	src := `resource "kion_cloud_rule" "soc_2" {
  name = "rule"
}

resource "kion_compliance_program" "soc_2" {
  name = "program"
}

resource "kion_ou" "child" {
  name         = "eng"
  parent_ou_id = kion_cloud_rule.soc_2.id
}
`
	out, _, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	got := norm(string(out))

	assert.Contains(t, got, "parent_ou_id = module.cloud_rule_soc_2.id")
	assert.NotContains(t, got, "module.soc_2.id")
}

// A reference to a resource that is NOT converted must survive verbatim.
func TestRewrite_LeavesUnconvertedReferencesAlone(t *testing.T) {
	t.Parallel()
	src := `resource "not_kion_thing" "a" {
  x = 1
}

resource "kion_ou" "child" {
  name = not_kion_thing.a.id
}
`
	out, _, err := importmodules.Rewrite([]byte(src), fixtureManifest(), "./modules")
	require.NoError(t, err)
	assert.Contains(t, norm(string(out)), "name = not_kion_thing.a.id")
}

// var./local./data. traversals are not resource references and must not be
// reported, or every generated file with a variable in it would fail.
func TestFindDanglingRefs_VariablesAreNotReferences(t *testing.T) {
	t.Parallel()
	src := `variable "n" {}

resource "kion_ou" "child" {
  name = var.n
}
`
	got, err := importmodules.FindDanglingRefs([]byte(src), fixtureManifest())
	require.NoError(t, err)
	assert.Empty(t, got)
}
