package migrate

import "sort"

// Delta is the old→new attribute-level change set for one resource type. Renames
// cannot be auto-detected (they are semantic), they surface here as a Dropped
// old name + an Added new name, and are resolved in codegen/state_upgrades.yaml.
type Delta struct {
	TypeChanged   []string // attr present in both but its type/kind changed
	Dropped       []string // old attr absent from new
	Added         []string // new attr absent from old
	IDTypeChanged bool     // id present in both but its type changed (str→num)
}

// sameShape reports whether two attributes have the same migration-relevant
// shape, container kind, type, and block nesting. It deliberately ignores the
// optional/required/computed flags, which change constantly across a major
// version but do not, on their own, require a state upgrader.
func sameShape(a, b Attr) bool {
	return a.Kind == b.Kind && a.TypeJSON == b.TypeJSON && a.Nesting == b.Nesting
}

// Diff computes the delta for every old resource type that has a same-named new
// type. Types with no same-named counterpart are omitted (there are none in
// practice; renames are handled separately).
func Diff(oldS, newS map[string]Resource) map[string]Delta {
	out := map[string]Delta{}
	for rtype, oldR := range oldS {
		newR, ok := newS[rtype]
		if !ok {
			continue
		}
		var d Delta
		for name, oa := range oldR.Attrs {
			na, ok := newR.Attrs[name]
			if !ok {
				d.Dropped = append(d.Dropped, name)
				continue
			}
			if !sameShape(oa, na) {
				d.TypeChanged = append(d.TypeChanged, name)
			}
		}
		for name := range newR.Attrs {
			if _, ok := oldR.Attrs[name]; !ok {
				d.Added = append(d.Added, name)
			}
		}
		if oa, ok := oldR.Attrs["id"]; ok {
			if na, ok := newR.Attrs["id"]; ok && !sameShape(oa, na) {
				d.IDTypeChanged = true
			}
		}
		sort.Strings(d.TypeChanged)
		sort.Strings(d.Dropped)
		sort.Strings(d.Added)
		out[rtype] = d
	}
	return out
}
