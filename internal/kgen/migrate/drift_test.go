package migrate

import (
	"testing"
)

// structuralKind reports the migration-relevant kind of an attribute: its
// container shape (attr type or block nesting). Two attributes with different
// kinds cannot be decoded from one another's state without an upgrader.
func structuralKind(r Resource, name string) string {
	a, ok := r.Attrs[name]
	if !ok {
		return ""
	}
	if a.Kind == "block" {
		return "block:" + a.Nesting
	}
	return "attr:" + a.TypeJSON
}

// TestDrift_everyStructuralChangeIsCovered fails if any old resource has a
// structural change (attribute container/type change, or id string→number) that
// is NOT declared in codegen/state_upgrades.yaml. Scalar-only add/drop is
// tolerated by Terraform and needs no upgrader. The documented alias exception is
// allow-listed. This is the completeness guarantee for customer state migration.
func TestDrift_everyStructuralChangeIsCovered(t *testing.T) {
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

	// Alias resources are covered by a transform keyed under their new primary
	// type with old_type set; index those so the alias old type counts as covered.
	allowed := map[string]bool{}
	for _, tr := range ups {
		if tr.OldType != "" {
			allowed[tr.OldType] = true
		}
	}

	for rt, oldR := range oldS {
		newR, ok := newS[rt]
		if !ok {
			continue
		}
		var structural []string
		for name := range oldR.Attrs {
			if _, in := newR.Attrs[name]; !in {
				continue // dropped scalar/block, tolerated
			}
			if structuralKind(oldR, name) != structuralKind(newR, name) {
				structural = append(structural, name)
			}
		}
		idOld, idNew := oldR.Attrs["id"], newR.Attrs["id"]
		idChanged := idOld.Kind == "attr" && idNew.Kind == "attr" && idOld.TypeJSON != idNew.TypeJSON

		if len(structural) == 0 && !idChanged {
			continue
		}
		if _, covered := ups[rt]; covered || allowed[rt] {
			continue
		}
		t.Errorf("%s has structural changes %v (id-changed=%v) but no codegen/state_upgrades.yaml entry; customer state will fail to decode", rt, structural, idChanged)
	}
}
