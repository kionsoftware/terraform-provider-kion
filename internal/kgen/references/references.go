// Package references records which schema attributes hold the id of another
// Kion resource.
//
// `terraform plan -generate-config-out` writes a foreign key as a bare integer
// (`ou_id = 11`) because Terraform cannot know the value points at anything.
// The resulting configuration is correct for the install it was imported from
// and wrong anywhere else, and it does not express its own dependency graph:
// destroy-and-reapply will not order correctly, and recreating an OU will not
// cascade. This package is the missing knowledge -- attribute to target type --
// that lets those literals be rewritten as `ou_id = kion_ou.engineering.id`.
//
// The map is authored in codegen/references.yaml, not derived, because the
// attribute name does not determine the target: payer_id is a billing source,
// ugroup_ids is a user group, and portfolio_id is an AWS id that must NOT be
// rewritten at all. Completeness is enforced by a test rather than by
// derivation -- see TestReferencesAreComplete.
package references

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// fkAttrRe matches the attribute names that must be classified: those ending in
// _id or _ids. `id` itself is the resource's own identity, never a reference.
var fkAttrRe = regexp.MustCompile(`_ids?$`)

// IsForeignKeyShaped reports whether an attribute name is one the completeness
// test requires an entry for.
func IsForeignKeyShaped(attr string) bool {
	return attr != "id" && fkAttrRe.MatchString(attr)
}

// Map is the loaded reference table.
type Map struct {
	// refs is attribute name -> target tf_type ("ou_id" -> "kion_ou").
	refs map[string]string
	// notRefs is attribute name -> why it is not a reference.
	notRefs map[string]string
	// overrides is tf_type -> attribute -> target, for a resource whose
	// attribute disagrees with the global table.
	overrides map[string]map[string]string
	// families is target -> the tf_types sharing its id space.
	families map[string][]string
}

type file struct {
	References    map[string]string            `yaml:"references"`
	NotReferences map[string]string            `yaml:"not_references"`
	Overrides     map[string]map[string]string `yaml:"overrides"`
	Families      map[string][]string          `yaml:"families"`
}

// Parse loads the reference table from codegen/references.yaml content.
func Parse(raw []byte) (*Map, error) {
	var f file
	if err := yaml.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("parsing references: %w", err)
	}
	m := &Map{
		refs:      f.References,
		notRefs:   f.NotReferences,
		overrides: f.Overrides,
		families:  f.Families,
	}
	if m.refs == nil {
		m.refs = map[string]string{}
	}
	if m.notRefs == nil {
		m.notRefs = map[string]string{}
	}
	if m.overrides == nil {
		m.overrides = map[string]map[string]string{}
	}
	// An attribute in both tables is a contradiction, and the resulting
	// behavior would depend on lookup order. Refuse it.
	var both []string
	for attr := range m.refs {
		if _, ok := m.notRefs[attr]; ok {
			both = append(both, attr)
		}
	}
	if len(both) > 0 {
		sort.Strings(both)
		return nil, fmt.Errorf("attributes in both references and not_references: %s", strings.Join(both, ", "))
	}
	for _, target := range m.refs {
		if !strings.HasPrefix(target, "kion_") {
			return nil, fmt.Errorf("reference target %q is not a tf_type (want a kion_ prefix)", target)
		}
	}
	for family, members := range m.families {
		for _, member := range members {
			if !strings.HasPrefix(member, "kion_") {
				return nil, fmt.Errorf("family %s member %q is not a tf_type (want a kion_ prefix)", family, member)
			}
		}
	}
	return m, nil
}

// Members returns the tf_types a reference to target may resolve to: its
// family when it has one, else target alone.
func (m *Map) Members(target string) []string {
	if members, ok := m.families[target]; ok {
		return members
	}
	return []string{target}
}

// Target returns the tf_type an attribute points at, for the given resource.
// ok is false when the attribute is not a reference.
func (m *Map) Target(tfType, attr string) (target string, ok bool) {
	if byAttr, has := m.overrides[tfType]; has {
		if t, has := byAttr[attr]; has {
			return t, true
		}
	}
	t, has := m.refs[attr]
	return t, has
}

// Classified reports whether the attribute appears in either table. Used by the
// completeness test; a false here means references.yaml needs an entry.
func (m *Map) Classified(attr string) bool {
	if _, ok := m.refs[attr]; ok {
		return true
	}
	_, ok := m.notRefs[attr]
	return ok
}

// Attrs returns every attribute known to be a reference, sorted.
func (m *Map) Attrs() []string {
	out := make([]string, 0, len(m.refs))
	for a := range m.refs {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}
