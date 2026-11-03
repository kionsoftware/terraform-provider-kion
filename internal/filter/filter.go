// Package filter provides a shared filter+match implementation for data
// sources that need to support the legacy `filter { name=... values=[...] }`
// block syntax from terraform-provider-kion (the SDK v2 provider).
//
// This is a backwards-compatibility layer: existing terraform configurations
// that look up resources by attribute (rather than by id) will continue to
// work against the framework-based provider unchanged.
//
// Usage in a data source:
//
//	type fooDataSourceModel struct {
//	    Filter []filter.Model `tfsdk:"filter"`
//	    List   types.List           `tfsdk:"list"`
//	    // ... id-mode scalar fields ...
//	}
//
//	func (d *fooDataSource) Schema(...) {
//	    resp.Schema = schema.Schema{
//	        Attributes: map[string]schema.Attribute{ ... },
//	        Blocks: map[string]schema.Block{
//	            "filter": filter.Schema(),
//	        },
//	    }
//	}
//
//	// In Read(), after fetching all rows from the API:
//	rows := []map[string]any{ ... }
//	for _, row := range rows {
//	    matched, diags := filter.Match(ctx, model.Filter, row)
//	    if diags.HasError() { ... }
//	    if matched { keep(row) }
//	}
package filter

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-kion/internal/flex"
)

// Model is the canonical Go representation of a single `filter` block.
// All filterable data sources should embed `[]Model` in their tfsdk
// model under the field tag `tfsdk:"filter"`.
type Model struct {
	Name   types.String `tfsdk:"name"`
	Values types.List   `tfsdk:"values"`
	Regex  types.Bool   `tfsdk:"regex"`
}

// Schema returns the standard ListNestedBlock schema for a filter block.
// Use it inside a data source's `Blocks` map under the key `"filter"` so all
// filterable data sources expose an identical filter syntax.
func Schema() schema.Block {
	return schema.ListNestedBlock{
		Description: "Filter results by attribute. Multiple filter blocks are AND'd; values within a single filter are OR'd.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Description: "The field name whose values you wish to filter by. Supports dotted paths (e.g. \"owner.id\") to traverse nested objects.",
					Required:    true,
				},
				"values": schema.ListAttribute{
					Description: "The values of the field name you specified. A row matches if its field equals any of these values.",
					Required:    true,
					ElementType: types.StringType,
				},
				"regex": schema.BoolAttribute{
					Description: "When true, the values are treated as regular expressions instead of exact matches.",
					Optional:    true,
				},
			},
		},
	}
}

// Match returns true if the given row satisfies every filter in the slice.
// Empty/nil filter slices match every row by default.
//
// Within a single filter, the row matches if the row's field equals (or, when
// regex=true, matches as a regex) any of the supplied values. Across filters,
// every filter must match (logical AND).
//
// Filter `name` supports dotted paths (e.g. "owner.id") that walk nested
// `map[string]any` and `[]any` structures recursively, mirroring the legacy
// behavior from terraform-provider-kion's kionclient.Filterable.
func Match(ctx context.Context, filters []Model, row map[string]any) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(filters) == 0 {
		return true, diags
	}

	for _, f := range filters {
		if f.Name.IsNull() || f.Name.IsUnknown() {
			diags.AddError("invalid filter", "filter.name is required")
			return false, diags
		}

		values, valDiags := flex.StringSliceFromFramework(ctx, f.Values)
		diags.Append(valDiags...)
		if diags.HasError() {
			return false, diags
		}
		if len(values) == 0 {
			diags.AddError("invalid filter", fmt.Sprintf("filter %q must have at least one value", f.Name.ValueString()))
			return false, diags
		}

		useRegex := !f.Regex.IsNull() && !f.Regex.IsUnknown() && f.Regex.ValueBool()
		keys := strings.Split(f.Name.ValueString(), ".")

		anyMatched := false
		for _, v := range values {
			matched, mDiags := deepMatch(keys, row, v, useRegex, f.Name.ValueString())
			diags.Append(mDiags...)
			if diags.HasError() {
				return false, diags
			}
			if matched {
				anyMatched = true
				break
			}
		}
		if !anyMatched {
			return false, diags
		}
	}

	return true, diags
}

// deepMatch walks the row map by the dotted key path and compares the leaf
// value against filterValue. When the path passes through a slice, every
// element is tried (any match wins).
func deepMatch(keys []string, m map[string]any, filterValue string, useRegex bool, fullKey string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(keys) == 0 {
		return false, diags
	}

	val, ok := m[keys[0]]
	if !ok {
		// Missing key is a non-match (rather than an error) so a single
		// filter expression can be applied to heterogeneous rows that may
		// not all carry the same fields.
		return false, diags
	}

	if len(keys) == 1 {
		// Leaf key. Refuse to compare against an array, the legacy provider
		// errored out in this case to surface configuration mistakes.
		if _, isSlice := val.([]any); isSlice {
			diags.AddError("invalid filter target",
				fmt.Sprintf("filter key %q references an array; use a dotted path to address a scalar field on the array elements", fullKey))
			return false, diags
		}
		if useRegex {
			re, err := regexp.Compile(filterValue)
			if err != nil {
				diags.AddError("invalid regular expression",
					fmt.Sprintf("filter %q value %q is not a valid regex: %v", fullKey, filterValue, err))
				return false, diags
			}
			return re.MatchString(fmt.Sprint(val)), diags
		}
		return fmt.Sprint(val) == filterValue, diags
	}

	// Non-leaf key: descend into the value.
	switch next := val.(type) {
	case map[string]any:
		return deepMatch(keys[1:], next, filterValue, useRegex, fullKey)
	case []any:
		for _, item := range next {
			if itemMap, ok := item.(map[string]any); ok {
				matched, mDiags := deepMatch(keys[1:], itemMap, filterValue, useRegex, fullKey)
				diags.Append(mDiags...)
				if diags.HasError() {
					return false, diags
				}
				if matched {
					return true, diags
				}
			}
		}
	}

	return false, diags
}
