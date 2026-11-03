package migrate

import (
	"strings"
	"testing"
)

// TestRewriteFile_dropFoldBlockAttr covers the three non-projection transforms:
// dropping an obsolete attribute (ConfigDrops), folding scattered attributes
// into a nested object (AttrsToObject, azure_policy), and converting a
// nested-object block into a list attribute (BlockToListAttr, project_funding).
func TestRewriteFile_dropFoldBlockAttr(t *testing.T) {
	src := `resource "kion_azure_policy" "p" {
  name        = "p1"
  description = "d"
  policy      = "{}"
  parameters  = "{}"
  last_updated = "x"
}

resource "kion_project" "pr" {
  name = "proj"
  project_funding {
    amount            = 100
    funding_source_id = 7
  }
}
`
	out, changes, _, err := RewriteFile([]byte(src), map[string]Transform{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	norm := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	ng := norm(got)

	// azure_policy: body folded, last_updated dropped.
	if !strings.Contains(ng, "azure_policy = {") {
		t.Errorf("azure_policy object not created:\n%s", got)
	}
	if strings.Contains(got, "last_updated") {
		t.Errorf("last_updated not dropped:\n%s", got)
	}
	for _, f := range []string{"name", "description", "policy", "parameters"} {
		if !strings.Contains(ng, f+" =") {
			t.Errorf("folded field %q missing:\n%s", f, got)
		}
	}
	// project_funding block → list attribute.
	if !strings.Contains(ng, "project_funding = [{") {
		t.Errorf("project_funding not converted to list attr:\n%s", got)
	}
	if strings.Contains(got, "project_funding {") {
		t.Errorf("project_funding block not removed:\n%s", got)
	}
	if len(changes) == 0 {
		t.Error("no changes reported")
	}
}

// TestRewriteFile_readOnlyDrops covers the ConfigDrops pass on a block that also
// carries an attribute the new schema still takes. last_updated is gone from the
// new schema and has to leave the config; create_user_id is Required in both
// providers, so removing it would strip a required argument out of a working
// configuration. The surrounding settable attributes must survive untouched.
func TestRewriteFile_readOnlyDrops(t *testing.T) {
	src := `resource "kion_project_note" "n" {
  name           = "note"
  project_id     = 3
  text           = "hello"
  create_user_id = 1
  last_updated   = "x"
}
`
	out, changes, _, err := RewriteFile([]byte(src), map[string]Transform{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if strings.Contains(got, "last_updated") {
		t.Errorf("last_updated not dropped:\n%s", got)
	}
	for _, keep := range []string{"name", "project_id", "text", "create_user_id"} {
		if !strings.Contains(got, keep) {
			t.Errorf("settable attribute %q was removed:\n%s", keep, got)
		}
	}
	var reported bool
	for _, c := range changes {
		if strings.Contains(c, "last_updated") {
			reported = true
		}
	}
	if !reported {
		t.Errorf("the drop was not reported to the practitioner: %v", changes)
	}
	// create_user_id must not be reported either, since it is not touched.
	for _, c := range changes {
		if strings.Contains(c, "create_user_id") {
			t.Errorf("create_user_id is Required in both providers and must be left alone: %v", changes)
		}
	}
}

// TestRewriteFile_dynamicBlocks covers the `dynamic` form of a repeatable block.
// Customers generate ownership blocks from a variable far more often than they
// write them out, so the projection has to collapse the whole
// dynamic/for_each/content construct into one for expression over the same
// for_each. There is no `dynamic` for an attribute.
func TestRewriteFile_dynamicBlocks(t *testing.T) {
	ups := map[string]Transform{
		"kion_user_group": {
			Project: map[string]ProjectRule{
				"owner_user_ids":       {From: "owner_users", Field: "id"},
				"owner_user_group_ids": {From: "owner_groups", Field: "id"},
			},
		},
		"kion_ou": {
			Project: map[string]ProjectRule{
				"owner_user_ids": {From: "owner_users", Field: "id"},
			},
		},
	}

	src := `resource "kion_user_group" "team" {
  name = "platform"

  dynamic "owner_users" {
    for_each = var.owner_user_ids
    content {
      id = owner_users.value
    }
  }

  dynamic "owner_groups" {
    for_each = var.owner_user_group_ids
    content {
      id = owner_groups.value
    }
  }
}

resource "kion_ou" "mixed" {
  owner_users {
    id = 5
  }

  dynamic "owner_users" {
    for_each = local.extra
    iterator = o
    content {
      id = o.value + o.key
    }
  }
}
`
	out, changes, actions, err := RewriteFile([]byte(src), ups)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	norm := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	ng := norm(got)

	for _, want := range []string{
		// The straight case: for_each carries straight through.
		"owner_user_ids = [for owner_users in var.owner_user_ids : owner_users]",
		"owner_user_group_ids = [for owner_groups in var.owner_user_group_ids : owner_groups]",
		// A custom iterator, and a body reading .key, which needs the
		// two-variable for form.
		"owner_user_ids = concat([5], [for o_key, o in local.extra : o + o_key])",
	} {
		if !strings.Contains(ng, norm(want)) {
			t.Errorf("output missing %q\n---\n%s", want, got)
		}
	}
	if strings.Contains(got, "dynamic ") {
		t.Errorf("dynamic blocks not removed:\n%s", got)
	}
	if len(actions) != 0 {
		t.Errorf("unexpected manual follow-ups: %v", actions)
	}
	if len(changes) != 3 {
		t.Errorf("changes = %v, want 3", changes)
	}
}

// TestRewriteFile_dynamicBlockToListAttr covers the same dynamic form for the
// object-bodied blocks (project_funding/budget), which become a list of objects
// rather than a list of ids.
func TestRewriteFile_dynamicBlockToListAttr(t *testing.T) {
	src := `resource "kion_project" "p" {
  dynamic "budget" {
    for_each = var.budgets
    content {
      amount         = budget.value.amount
      start_datecode = budget.value.start
    }
  }
}
`
	out, changes, actions, err := RewriteFile([]byte(src), map[string]Transform{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	norm := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	want := "budget = [for budget in var.budgets : { amount = budget.amount start_datecode = budget.start }]"
	if !strings.Contains(norm(got), norm(want)) {
		t.Errorf("output missing %q\n---\n%s", want, got)
	}
	if len(actions) != 0 {
		t.Errorf("unexpected manual follow-ups: %v", actions)
	}
	if len(changes) != 1 {
		t.Errorf("changes = %v, want 1", changes)
	}
}

// TestRewriteFile_dynamicUnconvertible asserts kmigrate reports rather than
// mangles a dynamic block it cannot read. A missing content field leaves the
// block alone and surfaces as a manual follow-up.
func TestRewriteFile_dynamicUnconvertible(t *testing.T) {
	ups := map[string]Transform{
		"kion_ou": {Project: map[string]ProjectRule{"owner_user_ids": {From: "owner_users", Field: "id"}}},
	}
	src := `resource "kion_ou" "o" {
  dynamic "owner_users" {
    for_each = var.ids
    content {
      user_id = owner_users.value
    }
  }
}
`
	out, changes, actions, err := RewriteFile([]byte(src), ups)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 0 {
		t.Errorf("changes = %v, want none", changes)
	}
	if len(actions) != 1 {
		t.Fatalf("actions = %v, want one manual follow-up", actions)
	}
	if !strings.Contains(string(out), `dynamic "owner_users"`) {
		t.Errorf("unconvertible block should be left in place:\n%s", out)
	}
}

// TestRewriteFile_dynamicMalformed feeds the rewriter dynamic blocks that are
// missing the parts it reads. kmigrate runs over files it did not write, so a
// half-written block must come back as a follow-up rather than a panic.
func TestRewriteFile_dynamicMalformed(t *testing.T) {
	ups := map[string]Transform{
		"kion_ou": {Project: map[string]ProjectRule{"owner_user_ids": {From: "owner_users", Field: "id"}}},
	}
	for name, src := range map[string]string{
		"no content": `resource "kion_ou" "o" {
  dynamic "owner_users" {
    for_each = var.ids
  }
}
`,
		"no for_each": `resource "kion_ou" "o" {
  dynamic "owner_users" {
    content {
      id = owner_users.value
    }
  }
}
`,
		"iterator is not an identifier": `resource "kion_ou" "o" {
  dynamic "owner_users" {
    for_each = var.ids
    iterator = "quoted"
    content {
      id = owner_users.value
    }
  }
}
`,
		// project_funding takes the same path but builds objects, not ids.
		"object-bodied block, no content": `resource "kion_project" "p" {
  dynamic "project_funding" {
    for_each = var.funding
  }
}
`,
	} {
		_, changes, actions, err := RewriteFile([]byte(src), ups)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if len(changes) != 0 {
			t.Errorf("%s: changes = %v, want none", name, changes)
		}
		if len(actions) != 1 {
			t.Errorf("%s: actions = %v, want one follow-up", name, actions)
		}
	}
}

// TestRewriteFile_mapToObjectList covers the cft tags rewrite: a map(string)
// attribute became a list of { tag_key, tag_value } objects, so the value itself
// has to be restructured, a literal map entry by entry, anything else as a for
// expression over the same value.
func TestRewriteFile_mapToObjectList(t *testing.T) {
	src := `resource "kion_aws_cloudformation_template" "literal" {
  tags = {
    env      = "coverage-test"
    "owner"  = var.owner
  }
}

resource "kion_aws_cloudformation_template" "computed" {
  tags = var.tags
}

resource "kion_aws_cloudformation_template" "already" {
  tags = [{ tag_key = "env", tag_value = "prod" }]
}
`
	// The rule comes from the shared state_upgrades transform, keyed by the new
	// primary type with old_type naming the alias customers still write.
	ups := map[string]Transform{
		"kion_cft": {
			OldType: "kion_aws_cloudformation_template",
			KVList:  map[string]KVListRule{"tags": {KeyField: "tag_key", ValField: "tag_value"}},
		},
	}
	out, changes, actions, err := RewriteFile([]byte(src), ups)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	norm := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	for _, want := range []string{
		`tags = [{ tag_key = "env", tag_value = "coverage-test" }, { tag_key = "owner", tag_value = var.owner }]`,
		`tags = [for k, v in var.tags : { tag_key = k, tag_value = v }]`,
		// Already migrated: left exactly as it was.
		`tags = [{ tag_key = "env", tag_value = "prod" }]`,
	} {
		if !strings.Contains(norm(got), norm(want)) {
			t.Errorf("output missing %q\n---\n%s", want, got)
		}
	}
	if len(actions) != 0 {
		t.Errorf("unexpected manual follow-ups: %v", actions)
	}
	// Two rewrites, not three: the already-migrated block must not be counted.
	if len(changes) != 2 {
		t.Errorf("changes = %v, want 2", changes)
	}
}

func TestRewriteFile_projectAndRename(t *testing.T) {
	ups := map[string]Transform{
		"kion_user_group": {
			Project: map[string]ProjectRule{
				"owner_user_ids": {From: "owner_users", Field: "id"},
				"user_ids":       {From: "users", Field: "id"},
			},
		},
		"kion_azure_account": {
			Rename: map[string]string{"name": "account_name"},
		},
	}

	src := `resource "kion_user_group" "team" {
  name = "platform"

  owner_users {
    id = 5
  }
  owner_users {
    id = kion_user.admin.id
  }
  users {
    id = 9
  }
}

resource "kion_azure_account" "prod" {
  name       = "prod"
  payer_id   = 3
}
`
	out, changes, _, err := RewriteFile([]byte(src), ups)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)

	for _, want := range []string{
		"owner_user_ids = [5, kion_user.admin.id]",
		"user_ids       = [9]", // hclwrite may align; check the value form loosely below
		"account_name = \"prod\"",
	} {
		// allow alignment whitespace differences on the "=" side
		norm := func(s string) string { return strings.Join(strings.Fields(s), " ") }
		if !strings.Contains(norm(got), norm(want)) {
			t.Errorf("output missing %q\n---\n%s", want, got)
		}
	}
	if strings.Contains(got, "owner_users {") || strings.Contains(got, "users {") {
		t.Errorf("source blocks not removed:\n%s", got)
	}
	// the azure_account block must no longer declare a bare `name` attribute
	azBlock := got[strings.Index(got, `"kion_azure_account"`):]
	if strings.Contains(azBlock, "\n  name ") || strings.Contains(azBlock, "\n  name=") {
		t.Errorf("azure_account name not renamed:\n%s", got)
	}
	if len(changes) == 0 {
		t.Error("no changes reported")
	}
}

// TestRewriteFile_providerConstraint covers the bump of the required_providers
// version pin. Found against a real configuration produced by the import tool:
// every resource migrated correctly, but the constraint still demanded the OLD
// provider, so `terraform init` resolved 0.3.34 and every rewrite kmigrate had
// just made became invalid config. A real tree has one of these per directory,
// eight in the case that surfaced it, so leaving it to the practitioner means
// leaving them to find all of them.
func TestRewriteFile_providerConstraint(t *testing.T) {
	src := `terraform {
  required_providers {
    kion = {
      source  = "kionsoftware/kion"
      version = "0.3.34"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.1.0"
    }
  }
}
`
	out, changes, _, err := RewriteFile([]byte(src), map[string]Transform{})
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)

	if !strings.Contains(got, `version = "`+NewProviderConstraint+`"`) {
		t.Errorf("kion constraint not bumped:\n%s", got)
	}
	if strings.Contains(got, `"0.3.34"`) {
		t.Errorf("old kion pin survived:\n%s", got)
	}
	// A different provider that merely sits alongside must not be touched.
	if !strings.Contains(got, `version = "3.1.0"`) {
		t.Errorf("unrelated provider was modified:\n%s", got)
	}
	if len(changes) != 1 {
		t.Errorf("want exactly 1 change, got %d: %v", len(changes), changes)
	}

	// Idempotent: a second pass reports nothing and changes nothing.
	out2, changes2, _, err := RewriteFile(out, map[string]Transform{})
	if err != nil {
		t.Fatal(err)
	}
	if len(changes2) != 0 {
		t.Errorf("second pass should be a no-op, got: %v", changes2)
	}
	if string(out2) != got {
		t.Errorf("second pass altered the file")
	}
}

// TestRewriteNamedFile_reportsRealFilename: a parse failure must name the file
// the practitioner has to fix. Every diagnostic used to say "config.tf", which
// over a directory named a file that existed nowhere in the tree.
func TestRewriteNamedFile_reportsRealFilename(t *testing.T) {
	// Invalid: a traversal cannot start with a digit. This is exactly what the
	// import tool emitted for a Kion resource named "800-53 Audit".
	bad := `output "x" {
  value = kion_cloud_rule.800-53_Audit.id
}
`
	_, _, _, err := RewriteNamedFile("cloud-rule/800-53_Audit.tf", []byte(bad), map[string]Transform{})
	if err == nil {
		t.Fatal("want a parse error, got nil")
	}
	if !strings.Contains(err.Error(), "cloud-rule/800-53_Audit.tf") {
		t.Errorf("error should name the real file, got: %v", err)
	}
	if strings.Contains(err.Error(), "config.tf") {
		t.Errorf("error still reports the placeholder filename: %v", err)
	}
}
