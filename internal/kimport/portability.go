package kimport

import (
	"sort"
	"strings"
)

// Warning is a record that imports cleanly where it was read but whose
// generated configuration will not create it on another install as-is.
type Warning struct {
	TFType string
	ID     string
	Name   string
	Msg    string
}

// portabilityRules are checks on a record's list payload, keyed by type. Each
// was found applying one install's configuration to another; see
// docs/guides/importing.md "Applying to another install".
var portabilityRules = map[string]func(raw map[string]any) string{
	"kion_funding_source": func(raw map[string]any) string {
		if v, ok := raw["ou_id"].(float64); ok && v > 0 {
			return ""
		}
		return "has no OU, and Kion refuses to create a funding source without one (\"ou id required\"); set ou_id"
	},
	"kion_permission_scheme": func(raw map[string]any) string {
		if b, ok := raw["system_managed"].(bool); ok && b {
			return "is built in to every install; there, reference the install's own scheme instead of creating this one"
		}
		return ""
	},
	"kion_user_group": func(raw map[string]any) string {
		if n, ok := raw["name"].(string); ok && n == "Administrators" {
			return "is built in to every install, so creating it fails (\"User Group name already in use\"); reference the install's own group"
		}
		return ""
	},
	"kion_custom_variable": func(raw map[string]any) string {
		if n, ok := raw["name"].(string); ok && strings.HasPrefix(n, "system-") {
			return "is Kion-managed, and the system- prefix is reserved, so it cannot be created; drop it"
		}
		return ""
	},
}

// NotPortable returns the records whose configuration will fail to apply on
// another install, sorted by type then id.
func NotPortable(results []Result) []Warning {
	var out []Warning
	for _, res := range results {
		rule := portabilityRules[res.TFType]
		if rule == nil {
			continue
		}
		for _, rec := range res.Records {
			if msg := rule(rec.Raw); msg != "" {
				out = append(out, Warning{TFType: res.TFType, ID: rec.ID, Name: rec.Name, Msg: msg})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TFType != out[j].TFType {
			return out[i].TFType < out[j].TFType
		}
		return out[i].ID < out[j].ID
	})
	return out
}
