package migrate

import (
	"slices"
	"testing"
)

const (
	oldSnap = "../../../codegen/schema_snapshots/old.json"
	newSnap = "../../../codegen/schema_snapshots/new.json"
)

func TestLoadSchema_old(t *testing.T) {
	s, err := LoadSchema(oldSnap)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 33 {
		t.Fatalf("old schema: want 33 resources, got %d", len(s))
	}
	lbl, ok := s["kion_label"]
	if !ok {
		t.Fatal("kion_label missing")
	}
	for _, a := range []string{"id", "color", "key", "value"} {
		if _, ok := lbl.Attrs[a]; !ok {
			t.Errorf("kion_label missing attr %q", a)
		}
	}
}

func TestDiff_awsAccountIDTypeChanged(t *testing.T) {
	oldS, err := LoadSchema(oldSnap)
	if err != nil {
		t.Fatal(err)
	}
	newS, err := LoadSchema(newSnap)
	if err != nil {
		t.Fatal(err)
	}
	d := Diff(oldS, newS)
	if !d["kion_aws_account"].IDTypeChanged {
		t.Error("kion_aws_account id type change not detected")
	}
	// Only aws_account changes id type.
	for rt, delta := range d {
		if delta.IDTypeChanged && rt != "kion_aws_account" {
			t.Errorf("unexpected id-type change on %s", rt)
		}
	}
}

func TestDiff_userGroupOwnerUsersDropped(t *testing.T) {
	oldS, err := LoadSchema(oldSnap)
	if err != nil {
		t.Fatal(err)
	}
	newS, err := LoadSchema(newSnap)
	if err != nil {
		t.Fatal(err)
	}
	d := Diff(oldS, newS)["kion_user_group"]
	if !slices.Contains(d.Dropped, "owner_users") {
		t.Errorf("owner_users not in dropped: %v", d.Dropped)
	}
	if !slices.Contains(d.Added, "owner_user_ids") {
		t.Errorf("owner_user_ids not in added: %v", d.Added)
	}
}
