package migrate

import (
	"strings"
	"testing"
)

// TestGenerateUpgrade_userGroup verifies the generated upgrader for the vertical
// slice: projected id-lists, null added attrs, string/bool/int passthrough, and
// that the whole thing is valid formatted Go (format.Source inside GenerateUpgrade).
func TestGenerateUpgrade_userGroup(t *testing.T) {
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
	tf := ups["kion_user_group"]
	out, err := GenerateUpgrade("../../..", "kion_user_group", tf, oldS["kion_user_group"], newS["kion_user_group"])
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{
		"package user_group",
		"var _ resource.ResourceWithUpgradeState = &user_groupResource{}",
		`migratehelper.ProjectIDs(old["owner_users"], "id")`,
		`migratehelper.ProjectIDs(old["owner_groups"], "id")`,
		`migratehelper.ProjectIDs(old["users"], "id")`,
		`"viewer_user_ids":`,
		`"add_self_as_viewer":`,
		`migratehelper.OrNull(old["name"])`,
		`migratehelper.OrNull(old["id"])`,
		"tftypes.ValueFromJSON",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("generated upgrader missing: %s\n---\n%s", want, s)
		}
	}
}

// TestGenerateUpgrade_padsNewNestedFields: a block whose new nested object gained
// a field must have the key added as null, since tftypes rejects a missing key.
func TestGenerateUpgrade_padsNewNestedFields(t *testing.T) {
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
	out, err := GenerateUpgrade("../../..", "kion_project", ups["kion_project"], oldS["kion_project"], newS["kion_project"])
	if err != nil {
		t.Fatal(err)
	}
	want := `migratehelper.PadObjectKeys(old["move_ou_settings"], "spend_plan_setting")`
	if !strings.Contains(string(out), want) {
		t.Errorf("generated upgrader missing: %s\n---\n%s", want, out)
	}
}
