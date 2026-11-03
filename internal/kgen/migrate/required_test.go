package migrate

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

// computeRequiredAdditions derives, from the schema snapshots, the attributes
// the new schema requires that the old schema did not: absent, or optional.
// Rename targets are excluded: the rewrite supplies those from the old name, so
// they are not something the practitioner must add.
func computeRequiredAdditions(t *testing.T) map[string][]string {
	t.Helper()
	oldS, err := LoadSchema(oldSnap)
	if err != nil {
		t.Fatal(err)
	}
	newS, err := LoadSchema(newSnap)
	if err != nil {
		t.Fatal(err)
	}
	ups, err := LoadUpgrades("../../../codegen/state_upgrades.yaml")
	if err != nil {
		t.Fatal(err)
	}

	// Renamed types map old name -> new via the transform keyed under the new
	// type with old_type set. Derived rather than hand-listed: a hand-written
	// copy is one more place to forget when an alias is added.
	oldToNew := map[string]string{}
	for newType, tr := range ups {
		if tr.OldType != "" {
			oldToNew[tr.OldType] = newType
		}
	}

	out := map[string][]string{}
	for rtype, oldR := range oldS {
		newType := rtype
		if mapped, ok := oldToNew[rtype]; ok {
			newType = mapped
		}
		newR, ok := newS[newType]
		if !ok {
			t.Errorf("%s: no counterpart in the new provider", rtype)
			continue
		}

		// A rename's target is filled in by RewriteFile, so it is not missing.
		renamed := map[string]bool{}
		for _, tr := range []Transform{ups[newType], ups[rtype]} {
			for _, to := range tr.Rename {
				renamed[to] = true
			}
		}

		var missing []string
		for name, a := range newR.Attrs {
			if !a.Required || renamed[name] {
				continue
			}
			if oa, existed := oldR.Attrs[name]; !existed || !oa.Required {
				missing = append(missing, name)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			out[rtype] = missing
		}
	}
	return out
}

// TestRequiredAdditions_matchSchemas is the guarantee behind the map: if a future
// spec makes an existing resource require something the old provider never had,
// this fails until RequiredAdditions records it. So kmigrate keeps telling
// practitioners about breaks it cannot fix for them.
func TestRequiredAdditions_matchSchemas(t *testing.T) {
	want := computeRequiredAdditions(t)

	got := map[string][]string{}
	for k, v := range RequiredAdditions {
		c := append([]string(nil), v...)
		sort.Strings(c)
		got[k] = c
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("RequiredAdditions is out of step with the schema snapshots:\n  have: %v\n  want: %v", got, want)
	}
}

func TestRewriteFile_reportsMissingRequired(t *testing.T) {
	// A kion_user block as the old provider allowed it: no arguments at all.
	src := `resource "kion_user" "adopted" {
}
`
	_, changes, actions, err := RewriteFile([]byte(src), map[string]Transform{})
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 0 {
		t.Errorf("nothing is rewritable here, got changes: %v", changes)
	}
	if len(actions) != 1 {
		t.Fatalf("expected one follow-up, got %v", actions)
	}
	for _, want := range []string{"kion_user.adopted", "email", "username", "idms_id"} {
		if !strings.Contains(actions[0], want) {
			t.Errorf("follow-up should name %q, got: %s", want, actions[0])
		}
	}
}

func TestRewriteFile_silentWhenRequiredArePresent(t *testing.T) {
	src := `resource "kion_user" "full" {
  email      = "a@b.c"
  first_name = "A"
  idms_id    = 1
  last_name  = "B"
  username   = "ab"
}
`
	_, _, actions, err := RewriteFile([]byte(src), map[string]Transform{})
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 0 {
		t.Errorf("a complete block needs no follow-up, got: %v", actions)
	}
}

// The account resources gained account_name, but as a rename of the old `name`
// RewriteFile fills it in, so reporting it would send practitioners chasing a
// value they already have.
func TestRequiredAdditions_excludesRenameTargets(t *testing.T) {
	for _, rtype := range []string{"kion_azure_account", "kion_custom_account", "kion_gcp_account"} {
		for _, a := range RequiredAdditions[rtype] {
			if a == "account_name" {
				t.Errorf("%s: account_name is a rename of `name`, not an addition the user must make", rtype)
			}
		}
	}
}

// A cached account has no project, so "add project_id" alone would move it;
// the follow-up must point at kion_account_cache.
func TestRewriteFile_cachedAccountHint(t *testing.T) {
	src := `resource "kion_custom_account" "cached" {
  name           = "c"
  account_number = "123"
  payer_id       = 1
}
`
	_, _, actions, err := RewriteFile([]byte(src), map[string]Transform{})
	if err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected one follow-up, got %v", actions)
	}
	for _, want := range []string{"project_id", "start_datecode", "kion_account_cache"} {
		if !strings.Contains(actions[0], want) {
			t.Errorf("follow-up should name %q, got: %s", want, actions[0])
		}
	}
}
