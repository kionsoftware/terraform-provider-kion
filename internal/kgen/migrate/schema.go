// Package migrate holds the old→new state-migration codegen: it reads the two
// providers' `terraform providers schema -json` snapshots, diffs them, and (with
// codegen/state_upgrades.yaml) generates per-resource Terraform state upgraders
// plus drives the kmigrate HCL config rewriter.
package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Attr is one attribute or nested block of a resource schema.
type Attr struct {
	Kind      string // "attr" | "block"
	TypeJSON  string // attr: the cty type as JSON (e.g. `"string"`, `["list",...]`)
	Nesting   string // block: nesting_mode (e.g. "list", "set", "single")
	Optional  bool   // attr may be set in config
	Required  bool   // attr must be set in config
	Computed  bool   // attr is set by the provider
	NestedObj bool   // attr: has a nested_type (object/collection of objects)
	// NestedMode/NestedAttrs are the nesting mode and field names of the object
	// this attribute or block carries, so a rule that names fields inside a
	// nested object (cft's tag_key/tag_value) can be checked against the schema,
	// and an old block whose state JSON passes straight through into a new
	// nested-object attribute can be checked field-for-field.
	NestedMode  string
	NestedAttrs []string
	// NestedComputed holds the nested-object fields the provider can set, which
	// a rule may leave null for the next refresh to fill.
	NestedComputed map[string]bool
}

// Settable reports whether a user can write this attribute in config (so a
// dropped settable attribute is a config-migration concern, whereas a dropped
// computed-only attribute never appears in a customer's .tf).
func (a Attr) Settable() bool { return a.Optional || a.Required }

// Resource is a resource type's attribute set.
type Resource struct {
	Attrs map[string]Attr
}

// LoadSchema parses a `terraform providers schema -json` file into resources
// keyed by type (e.g. "kion_label").
func LoadSchema(path string) (map[string]Resource, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		ProviderSchemas map[string]struct {
			ResourceSchemas map[string]struct {
				Block struct {
					Attributes map[string]struct {
						Type       json.RawMessage `json:"type"`
						NestedType json.RawMessage `json:"nested_type"`
						Optional   bool            `json:"optional"`
						Required   bool            `json:"required"`
						Computed   bool            `json:"computed"`
					} `json:"attributes"`
					BlockTypes map[string]struct {
						NestingMode string `json:"nesting_mode"`
						Block       struct {
							Attributes map[string]json.RawMessage `json:"attributes"`
							BlockTypes map[string]json.RawMessage `json:"block_types"`
						} `json:"block"`
					} `json:"block_types"`
				} `json:"block"`
			} `json:"resource_schemas"`
		} `json:"provider_schemas"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := map[string]Resource{}
	for _, ps := range doc.ProviderSchemas {
		for rtype, rs := range ps.ResourceSchemas {
			r := Resource{Attrs: map[string]Attr{}}
			for name, a := range rs.Block.Attributes {
				attr := Attr{
					Kind:      "attr",
					TypeJSON:  string(a.Type),
					Optional:  a.Optional,
					Required:  a.Required,
					Computed:  a.Computed,
					NestedObj: len(a.NestedType) > 0,
				}
				if attr.NestedObj {
					var nt struct {
						NestingMode string `json:"nesting_mode"`
						Attributes  map[string]struct {
							Computed bool `json:"computed"`
						} `json:"attributes"`
					}
					if err := json.Unmarshal(a.NestedType, &nt); err != nil {
						return nil, fmt.Errorf("parse %s: nested_type of %s.%s: %w", path, rtype, name, err)
					}
					attr.NestedMode = nt.NestingMode
					attr.NestedComputed = map[string]bool{}
					for n, na := range nt.Attributes {
						attr.NestedAttrs = append(attr.NestedAttrs, n)
						if na.Computed {
							attr.NestedComputed[n] = true
						}
					}
					sort.Strings(attr.NestedAttrs)
				}
				r.Attrs[name] = attr
			}
			for name, b := range rs.Block.BlockTypes {
				a := Attr{Kind: "block", Nesting: b.NestingMode, NestedMode: b.NestingMode}
				for n := range b.Block.Attributes {
					a.NestedAttrs = append(a.NestedAttrs, n)
				}
				for n := range b.Block.BlockTypes {
					a.NestedAttrs = append(a.NestedAttrs, n)
				}
				sort.Strings(a.NestedAttrs)
				r.Attrs[name] = a
			}
			out[rtype] = r
		}
		break // the snapshot contains a single provider
	}
	return out, nil
}
