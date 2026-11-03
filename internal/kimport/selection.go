package kimport

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"terraform-provider-kion/internal/kgen/importmanifest"
)

// Selection narrows a run to some of the resource types the manifest knows.
//
// Importing everything is rarely what an operator wants: a real install carries
// tens of thousands of records in Kion's shipped policy and compliance catalogs,
// which dwarf the resources anyone manages in Terraform.
//
// Both spellings exist because they suit different uses. Flags are quicker for a
// one-off; a file is what you keep beside the configuration and re-run, and it
// takes comments explaining why a type is in or out.
type Selection struct {
	// Include, when non-empty, is the entire set to enumerate. Empty means every
	// type the manifest can read.
	Include []string `yaml:"include"`
	// Exclude drops types, and is applied after Include.
	Exclude []string `yaml:"exclude"`
}

// LoadSelection reads a selection file. An absent path yields an empty Selection.
func LoadSelection(path string) (Selection, error) {
	if path == "" {
		return Selection{}, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Selection{}, fmt.Errorf("reading selection %s: %w", path, err)
	}
	var s Selection
	dec := yaml.NewDecoder(strings.NewReader(string(raw)))
	dec.KnownFields(true)
	if err := dec.Decode(&s); err != nil {
		return Selection{}, fmt.Errorf("parsing selection %s: %w", path, err)
	}
	return s, nil
}

// Merge folds command-line values into a file-based selection. Flags append
// rather than replace, so a file can carry the standing policy and a flag can
// add one more exclusion for a single run.
func (s Selection) Merge(include, exclude []string) Selection {
	s.Include = append(append([]string(nil), s.Include...), include...)
	s.Exclude = append(append([]string(nil), s.Exclude...), exclude...)
	return s
}

// Apply returns the manifest rows to enumerate.
//
// A name that matches nothing is an error rather than a silent no-op. Typing
// kion_labels for kion_label should not quietly produce an import of everything.
func (s Selection) Apply(rs []importmanifest.Resource) ([]importmanifest.Resource, error) {
	known := make(map[string]bool, len(rs))
	for _, r := range rs {
		known[r.TFType] = true
	}
	var unknown []string
	for _, n := range append(append([]string(nil), s.Include...), s.Exclude...) {
		if !known[n] {
			unknown = append(unknown, n)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return nil, fmt.Errorf("unknown resource type(s): %s\nrun with --list-types to see the %d the manifest knows",
			strings.Join(unknown, ", "), len(rs))
	}

	inc := toSet(s.Include)
	exc := toSet(s.Exclude)
	out := make([]importmanifest.Resource, 0, len(rs))
	for _, r := range rs {
		if len(inc) > 0 && !inc[r.TFType] {
			continue
		}
		if exc[r.TFType] {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func toSet(xs []string) map[string]bool {
	if len(xs) == 0 {
		return nil
	}
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[strings.TrimSpace(x)] = true
	}
	return m
}
