package rewriteref_test

import (
	"strings"
	"testing"

	"terraform-provider-kion/codegen"
	"terraform-provider-kion/internal/kgen/references"
	"terraform-provider-kion/internal/kimport/rewriteref"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func refs(t *testing.T) *references.Map {
	t.Helper()
	m, err := references.Parse(codegen.ReferencesYAML)
	require.NoError(t, err)
	return m
}

// realImports is the shape kion-import writes, including a compound
// "<parent>/<id>" id for a parent-scoped resource.
const realImports = `
import {
  to = kion_ou.kion_ou_11
  id = "11"
}
import {
  to = kion_ou.kion_ou_1
  id = "1"
}
import {
  to = kion_project.kion_project_9
  id = "9"
}
import {
  to = kion_user.kion_user_4
  id = "4"
}
import {
  to = kion_user.kion_user_94
  id = "94"
}
import {
  to = kion_ou_cloud_access_role_exemption.ex_1_4
  id = "1/4"
}
import {
  to = kion_user.kion_user_119
  id = "119"
}
import {
  to = kion_user_group.old_kion_developers
  id = "7"
}
import {
  to = kion_ou.kion_ou_2
  id = "2"
}
import {
  to = kion_ou.kion_ou_3
  id = "3"
}
import {
  to = kion_project.kion_project_10
  id = "10"
}
`

func index(t *testing.T) rewriteref.Index {
	t.Helper()
	idx, err := rewriteref.BuildIndex([]byte(realImports), "imports.tf")
	require.NoError(t, err)
	return idx
}

// TestBuildIndexFromImportBlocks: generated.tf has no ids in it -- they are
// Computed, so -generate-config-out omits them -- so the id -> label index can
// only come from the import blocks.
func TestBuildIndexFromImportBlocks(t *testing.T) {
	idx := index(t)

	assert.Equal(t, "kion_ou_11", idx["kion_ou"]["11"])
	assert.Equal(t, "kion_project_9", idx["kion_project"]["9"])
	assert.Equal(t, "kion_user_94", idx["kion_user"]["94"])

	// A compound import id is indexed by its record segment too, because a
	// foreign key holds the plain record id, not "<parent>/<id>".
	ex := idx["kion_ou_cloud_access_role_exemption"]
	assert.Equal(t, "ex_1_4", ex["4"], "record id resolves")
	assert.Equal(t, "ex_1_4", ex["1/4"], "full compound id resolves too")
}

// TestRewriteScalarAndList is the core case: a single FK and a list of them.
func TestRewriteScalarAndList(t *testing.T) {
	in := `
resource "kion_project" "kion_project_9" {
  name           = "Web"
  ou_id          = 11
  owner_user_ids = [4, 94]
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)
	got := string(out)

	assert.Contains(t, got, "ou_id          = kion_ou.kion_ou_11.id")
	assert.Contains(t, got, "owner_user_ids = [kion_user.kion_user_4.id, kion_user.kion_user_94.id]")
	assert.Contains(t, got, `name           = "Web"`, "non-reference attributes untouched")
	assert.Equal(t, 3, res.Rewritten)
	assert.Empty(t, res.Unresolved)
}

// TestNotAReferenceIsLeftAlone covers the trap the authored table exists for:
// portfolio_id is an AWS id sitting beside internal_portfolio_ids, which IS a
// kion_service_catalog. Rewriting the former would point at a foreign id space.
func TestNotAReferenceIsLeftAlone(t *testing.T) {
	in := `
resource "kion_service_catalog" "sc" {
  portfolio_id    = 11
  account_type_id = 1
  car_external_id = 9
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	got := string(out)
	assert.Contains(t, got, "portfolio_id    = 11", "an AWS id is not a Kion reference")
	assert.Contains(t, got, "account_type_id = 1", "an enum is not a reference")
	assert.Contains(t, got, "car_external_id = 9")
	assert.Zero(t, res.Rewritten)
	assert.Empty(t, res.Unresolved, "a non-reference is not a boundary gap either")
}

// TestUnresolvedIsReportedNotRewritten: a referent that is not managed in this
// configuration is the boundary. The literal stays, and it is reported so the
// operator knows what needs a data source, a variable, or another import.
func TestUnresolvedIsReportedNotRewritten(t *testing.T) {
	in := `
resource "kion_project" "kion_project_9" {
  ou_id          = 11
  owner_user_ids = [4, 777]
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	got := string(out)
	assert.Contains(t, got, "kion_user.kion_user_4.id", "the resolvable one is rewritten")
	assert.Contains(t, got, "777", "the unresolvable one is left as a literal")

	require.Len(t, res.Unresolved, 1)
	assert.Equal(t, "kion_project.kion_project_9", res.Unresolved[0].Resource)
	assert.Equal(t, "owner_user_ids", res.Unresolved[0].Attr)
	assert.Equal(t, "kion_user", res.Unresolved[0].Target)
	assert.Equal(t, "777", res.Unresolved[0].Value)
	assert.Equal(t, map[string]int{"kion_user": 1}, res.TargetCounts())
}

// TestParentOUZeroKeepsSentinel: parent_ou_id = 0 means "top-level OU". A
// reference there would invent a parent that does not exist.
func TestParentOUZeroKeepsSentinel(t *testing.T) {
	in := `
resource "kion_ou" "kion_ou_1" {
  name         = "Root"
  parent_ou_id = 0
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	assert.Contains(t, string(out), "parent_ou_id = 0")
	assert.Zero(t, res.Rewritten)
	assert.Empty(t, res.Unresolved, "0 is a sentinel, not a missing referent")
}

// TestSelfReferenceIsRefused: a block whose FK resolves to itself would be a
// cycle Terraform refuses to plan. It is the degenerate one-node case of the
// general check, and is reported like any other.
func TestSelfReferenceIsRefused(t *testing.T) {
	in := `
resource "kion_ou" "kion_ou_11" {
  parent_ou_id = 11
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	assert.Contains(t, string(out), "parent_ou_id = 11")
	assert.Zero(t, res.Rewritten)
	require.Len(t, res.Cycles, 1)
	assert.Equal(t, "kion_ou.kion_ou_11", res.Cycles[0].Resource)
	assert.Equal(t, []string{"kion_ou.kion_ou_11"}, res.Cycles[0].Path)
}

// mutualMembership is the shape that made #37: user<->group membership is
// settable from BOTH sides, so rewriting both directions closes a loop and
// `terraform validate` fails with "Cycle: kion_user.kion_user_119,
// kion_user_group.old_kion_developers".
const mutualMembership = `
resource "kion_user" "kion_user_119" {
  user_group_ids = [7]
}

resource "kion_user_group" "old_kion_developers" {
  user_ids       = [119]
  owner_user_ids = [4]
}
`

// TestMutualMembershipIsNotACycle: one direction becomes a reference, the other
// keeps its literal. Which one wins is decided by file order, not by which
// attribute the table happens to hand back first.
func TestMutualMembershipIsNotACycle(t *testing.T) {
	out, res, err := rewriteref.Rewrite([]byte(mutualMembership), "generated.tf", index(t), refs(t))
	require.NoError(t, err)
	got := string(out)

	// kion_user comes first in the file, so its edge is offered first and kept.
	assert.Contains(t, got, "user_group_ids = [kion_user_group.old_kion_developers.id]")
	assert.Contains(t, got, "user_ids       = [119]", "the closing edge keeps its literal")
	// An attribute of the losing block that does NOT close a cycle still rewrites.
	assert.Contains(t, got, "owner_user_ids = [kion_user.kion_user_4.id]")
	assert.Equal(t, 2, res.Rewritten)
	assert.Empty(t, res.Unresolved, "both referents are managed here; nothing is missing")

	require.Len(t, res.Cycles, 1)
	c := res.Cycles[0]
	assert.Equal(t, "kion_user_group.old_kion_developers", c.Resource)
	assert.Equal(t, "user_ids", c.Attr)
	assert.Equal(t, "kion_user", c.Target)
	assert.Equal(t, "119", c.Value)
	assert.Equal(t, "kion_user.kion_user_119", c.Ref)
	assert.Equal(t, []string{"kion_user.kion_user_119", "kion_user_group.old_kion_developers"}, c.Path)
	assert.Equal(t, map[string]int{"kion_user_group.user_ids": 1}, res.CycleCounts())
}

// TestCycleRefusalIsDeterministic: which edge loses must not depend on Go map
// iteration order, or the output changes between runs of the same input.
func TestCycleRefusalIsDeterministic(t *testing.T) {
	first, firstRes, err := rewriteref.Rewrite([]byte(mutualMembership), "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	for i := 0; i < 25; i++ {
		out, res, err := rewriteref.Rewrite([]byte(mutualMembership), "generated.tf", index(t), refs(t))
		require.NoError(t, err)
		require.Equal(t, string(first), string(out), "run %d produced different output", i)
		require.Equal(t, firstRes.Cycles, res.Cycles, "run %d refused a different edge", i)
	}
}

// TestRerunDoesNotCloseACycle: a second pass sees the first pass's references
// as traversals, not literals. Unless those seed the graph, the edge that lost
// the first time would win the second and produce the cycle after all.
func TestRerunDoesNotCloseACycle(t *testing.T) {
	once, _, err := rewriteref.Rewrite([]byte(mutualMembership), "generated.tf", index(t), refs(t))
	require.NoError(t, err)
	twice, res, err := rewriteref.Rewrite(once, "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	assert.Equal(t, string(once), string(twice), "rewriting is idempotent")
	assert.Zero(t, res.Rewritten)
	assert.Len(t, res.Cycles, 1, "still reported on the second pass")
}

// TestThreeNodeCycleIsBroken: the self-reference guard this replaces caught
// only one-node loops. A parent chain that closes on itself is a cycle at any
// length.
func TestThreeNodeCycleIsBroken(t *testing.T) {
	in := `
resource "kion_ou" "kion_ou_1" {
  parent_ou_id = 2
}

resource "kion_ou" "kion_ou_2" {
  parent_ou_id = 3
}

resource "kion_ou" "kion_ou_3" {
  parent_ou_id = 1
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)
	got := string(out)

	assert.Contains(t, got, "kion_ou.kion_ou_2.id")
	assert.Contains(t, got, "kion_ou.kion_ou_3.id")
	assert.Equal(t, 2, res.Rewritten, "two of three edges make a chain, not a loop")

	require.Len(t, res.Cycles, 1)
	c := res.Cycles[0]
	assert.Equal(t, "kion_ou.kion_ou_3", c.Resource, "the last edge offered is the one refused")
	assert.Equal(t, "1", c.Value)
	assert.Equal(t, []string{"kion_ou.kion_ou_1", "kion_ou.kion_ou_2", "kion_ou.kion_ou_3"}, c.Path)
	assert.Contains(t, got, "parent_ou_id = 1", "the closing edge keeps its literal")
}

// TestDiamondIsNotACycle guards the obvious over-correction: two blocks
// referencing the same third block share a referent but form no loop, and both
// must still rewrite. Shared parents are the common case, not the exception.
func TestDiamondIsNotACycle(t *testing.T) {
	in := `
resource "kion_project" "kion_project_9" {
  ou_id          = 11
  owner_user_ids = [4]
}

resource "kion_project" "kion_project_10" {
  ou_id          = 11
  owner_user_ids = [4]
}

resource "kion_ou" "kion_ou_11" {
  parent_ou_id = 1
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)
	got := string(out)

	assert.Equal(t, 5, res.Rewritten)
	assert.Empty(t, res.Cycles, "a diamond is legal; nothing may be refused")
	assert.Equal(t, 2, strings.Count(got, "ou_id          = kion_ou.kion_ou_11.id"))
	assert.Equal(t, 2, strings.Count(got, "owner_user_ids = [kion_user.kion_user_4.id]"))
	assert.Contains(t, got, "parent_ou_id = kion_ou.kion_ou_1.id")
}

// TestCycleRefusalIsPerIdNotPerAttribute: a list holding one closing id and one
// safe id keeps only the closing literal. Refusing the whole attribute would
// throw away edges that were never the problem.
func TestCycleRefusalIsPerIdNotPerAttribute(t *testing.T) {
	in := `
resource "kion_user" "kion_user_119" {
  user_group_ids = [7]
}

resource "kion_user_group" "old_kion_developers" {
  user_ids = [119, 4, 94]
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	assert.Contains(t, string(out),
		"user_ids = [119, kion_user.kion_user_4.id, kion_user.kion_user_94.id]")
	assert.Equal(t, 3, res.Rewritten)
	assert.Len(t, res.Cycles, 1)
}

// TestNestedOUHierarchyBecomesAGraph is the payoff: once parents are
// references, Terraform orders creation itself. This is what makes the same
// configuration apply to a different install with no id map.
func TestNestedOUHierarchyBecomesAGraph(t *testing.T) {
	in := `
resource "kion_ou" "kion_ou_11" {
  name         = "Engineering"
  parent_ou_id = 1
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	assert.Contains(t, string(out), "parent_ou_id = kion_ou.kion_ou_1.id")
	assert.Equal(t, 1, res.Rewritten)
}

// TestAlreadyReferencedIsIdempotent: running twice must not corrupt a
// configuration that already holds references.
func TestAlreadyReferencedIsIdempotent(t *testing.T) {
	in := `
resource "kion_project" "kion_project_9" {
  ou_id = 11
}
`
	once, _, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)
	twice, res, err := rewriteref.Rewrite(once, "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	assert.Equal(t, string(once), string(twice))
	assert.Zero(t, res.Rewritten, "nothing left to rewrite")
	assert.Empty(t, res.Unresolved)
}

// TestCommentsAndFormattingSurvive: -generate-config-out writes a provenance
// comment above every block, and losing it would make the output harder to
// review than what it replaced.
func TestCommentsAndFormattingSurvive(t *testing.T) {
	in := `# __generated__ by Terraform
# __generated__ by Terraform from "9"
resource "kion_project" "kion_project_9" {
  ou_id = 11
}
`
	out, _, err := rewriteref.Rewrite([]byte(in), "generated.tf", index(t), refs(t))
	require.NoError(t, err)

	got := string(out)
	assert.Contains(t, got, "# __generated__ by Terraform")
	assert.Contains(t, got, `from "9"`)
	assert.True(t, strings.HasPrefix(got, "# __generated__"))
}

func TestBuildIndexRejectsGarbage(t *testing.T) {
	_, err := rewriteref.BuildIndex([]byte("import {"), "imports.tf")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing imports.tf")
}

func TestRewriteRejectsGarbage(t *testing.T) {
	_, _, err := rewriteref.Rewrite([]byte("resource {"), "generated.tf", index(t), refs(t))
	require.Error(t, err)
}

// payer_id names a billing source of any kind: the typed resources share the
// payer id space, so an AWS payer resolves to kion_billing_source_aws.
func TestPayerIDResolvesAcrossBillingSourceTypes(t *testing.T) {
	imports := `
import {
  to = kion_billing_source_aws.kion_billing_source_aws_5
  id = "5"
}
import {
  to = kion_billing_source_gcp.kion_billing_source_gcp_7
  id = "7"
}
`
	idx, err := rewriteref.BuildIndex([]byte(imports), "imports.tf")
	require.NoError(t, err)
	in := `
resource "kion_account" "a" {
  payer_id = 5
}
resource "kion_account" "b" {
  payer_id = 7
}
resource "kion_account" "c" {
  payer_id = 3
}
`
	out, res, err := rewriteref.Rewrite([]byte(in), "generated.tf", idx, refs(t))
	require.NoError(t, err)
	got := string(out)
	assert.Contains(t, got, "payer_id = kion_billing_source_aws.kion_billing_source_aws_5.id")
	assert.Contains(t, got, "payer_id = kion_billing_source_gcp.kion_billing_source_gcp_7.id")
	assert.Contains(t, got, "payer_id = 3", "a payer no imported type holds stays literal")
	assert.Equal(t, 2, res.Rewritten)
	require.Len(t, res.Unresolved, 1)
	assert.Equal(t, "kion_billing_source", res.Unresolved[0].Target)
}
