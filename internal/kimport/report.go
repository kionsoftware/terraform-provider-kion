package kimport

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"terraform-provider-kion/internal/kgen/importmanifest"
)

// LoadManifest parses the embedded (or supplied) import manifest.
func LoadManifest(data []byte) (*importmanifest.Manifest, error) {
	var m importmanifest.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse import manifest: %w", err)
	}
	if m.Version != importmanifest.ManifestVersion {
		return nil, fmt.Errorf("import manifest version %d, want %d",
			m.Version, importmanifest.ManifestVersion)
	}
	return &m, nil
}

// FormatReport summarizes a run: totals, then only the rows a human must act on.
// Listing all 50 healthy kinds would bury the handful that failed.
func FormatReport(results []Result) string {
	counts := map[string]int{}
	total := 0
	for _, r := range results {
		counts[r.Status]++
		total += len(r.Records)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "\nCoverage: %d resource types, %d records\n", len(results), total)

	statuses := make([]string, 0, len(counts))
	for s := range counts {
		statuses = append(statuses, s)
	}
	sort.Strings(statuses)
	parts := make([]string, 0, len(statuses))
	for _, s := range statuses {
		parts = append(parts, fmt.Sprintf("%s: %d", s, counts[s]))
	}
	fmt.Fprintf(&b, "  %s\n", strings.Join(parts, ", "))

	// See partition's doc comment: gaps go under "Not imported", caveats --
	// a successful read that still carries a real gap in Reason -- are
	// reported separately so a partial read never looks clean. Mirrors
	// RenderImports' "Not imported"/"Read with caveats" sections in emit.go.
	gaps, caveats := partition(results)
	if len(gaps) > 0 {
		sort.Slice(gaps, func(i, j int) bool { return gaps[i].TFType < gaps[j].TFType })
		fmt.Fprintf(&b, "\nNot imported (%d):\n", len(gaps))
		for _, r := range gaps {
			fmt.Fprintf(&b, "  %-45s %-12s %s\n", r.TFType, r.Status, r.Reason)
		}
	}
	if len(caveats) > 0 {
		sort.Slice(caveats, func(i, j int) bool { return caveats[i].TFType < caveats[j].TFType })
		fmt.Fprintf(&b, "\nRead with caveats (%d):\n", len(caveats))
		for _, r := range caveats {
			fmt.Fprintf(&b, "  %-45s %-12s %s\n", r.TFType, r.Status, r.Reason)
		}
	}

	if ws := NotPortable(results); len(ws) > 0 {
		fmt.Fprintf(&b, "\nWill not apply to another install as-is (%d):\n", len(ws))
		for _, w := range ws {
			fmt.Fprintf(&b, "  %s %q (%s) %s\n", w.TFType, w.Name, w.ID, w.Msg)
		}
	}

	return b.String()
}
