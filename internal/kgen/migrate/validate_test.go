package migrate

import "testing"

// TestUpgrades_targetsAndSourcesExist validates every transform in
// codegen/state_upgrades.yaml against the real schema snapshots: each transform
// must WRITE an attribute the new schema actually has, and READ an attribute the
// old schema actually had. A projection into a non-existent new attribute (or
// from a non-existent old block) silently no-ops in the generator and ships
// broken migration state, exactly the azure_policy bug this guards against.
func TestUpgrades_targetsAndSourcesExist(t *testing.T) {
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

	for rt, tr := range ups {
		newR, ok := newS[rt]
		if !ok {
			t.Errorf("%s: in state_upgrades.yaml but not in the new schema snapshot", rt)
			continue
		}
		// Old-state sources live under the alias's old type when old_type is set.
		oldKey := rt
		if tr.OldType != "" {
			oldKey = tr.OldType
		}
		oldR := oldS[oldKey] // may be absent; sub-checks handle empties

		// Projections: target (new attr) and source block (old attr) must exist.
		for target, pr := range tr.Project {
			if _, ok := newR.Attrs[target]; !ok {
				t.Errorf("%s: project target %q is not an attribute of the new schema", rt, target)
			}
			if _, ok := oldR.Attrs[pr.From]; !ok {
				t.Errorf("%s: project source %q is not an attribute/block of the old schema", rt, pr.From)
			}
		}
		// Renames: new name must exist in new schema, old name in old schema.
		for oldA, newA := range tr.Rename {
			if _, ok := newR.Attrs[newA]; !ok {
				t.Errorf("%s: rename target %q is not an attribute of the new schema", rt, newA)
			}
			if _, ok := oldR.Attrs[oldA]; !ok {
				t.Errorf("%s: rename source %q is not an attribute of the old schema", rt, oldA)
			}
		}
		// Unwraps: the attribute must exist in both (same name, block → single).
		for _, u := range tr.Unwrap {
			if _, ok := newR.Attrs[u]; !ok {
				t.Errorf("%s: unwrap target %q is not an attribute of the new schema", rt, u)
			}
			if _, ok := oldR.Attrs[u]; !ok {
				t.Errorf("%s: unwrap source %q is not an attribute of the old schema", rt, u)
			}
		}
	}
}
