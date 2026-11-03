package flex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Expand: TF → SDK ---

// Uint64SliceFromFramework converts a types.List of Int64 to []uint64.
// Null or unknown lists return nil.
func Uint64SliceFromFramework(ctx context.Context, v types.List) ([]uint64, diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}

	var elems []types.Int64
	diags := v.ElementsAs(ctx, &elems, false)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]uint64, len(elems))
	for i, e := range elems {
		result[i] = uint64(e.ValueInt64())
	}
	return result, nil
}

// StringSliceFromFramework converts a types.List of String to []string.
// Null or unknown lists return nil.
func StringSliceFromFramework(ctx context.Context, v types.List) ([]string, diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}

	var elems []types.String
	diags := v.ElementsAs(ctx, &elems, false)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]string, len(elems))
	for i, e := range elems {
		result[i] = e.ValueString()
	}
	return result, nil
}

// StringSliceFromFrameworkSet converts a types.Set of String to []string.
// Null or unknown sets return nil.
func StringSliceFromFrameworkSet(ctx context.Context, v types.Set) ([]string, diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}

	var elems []types.String
	diags := v.ElementsAs(ctx, &elems, false)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]string, len(elems))
	for i, e := range elems {
		result[i] = e.ValueString()
	}
	return result, nil
}

// --- Flatten: SDK → TF ---

// Uint64SliceToFramework converts a []uint64 to a types.List of Int64.
// Nil slices produce a null types.List.
func Uint64SliceToFramework(ctx context.Context, v []uint64) (types.List, diag.Diagnostics) {
	if v == nil {
		return types.ListNull(types.Int64Type), nil
	}

	elems := make([]types.Int64, len(v))
	for i, val := range v {
		elems[i] = types.Int64Value(int64(val))
	}
	return types.ListValueFrom(ctx, types.Int64Type, elems)
}

// StringSliceToFramework converts a []string to a types.List of String.
// Nil slices produce a null types.List.
func StringSliceToFramework(ctx context.Context, v []string) (types.List, diag.Diagnostics) {
	if v == nil {
		return types.ListNull(types.StringType), nil
	}

	elems := make([]types.String, len(v))
	for i, val := range v {
		elems[i] = types.StringValue(val)
	}
	return types.ListValueFrom(ctx, types.StringType, elems)
}

// StringSliceToFrameworkSet converts a []string to a types.Set of String.
// Nil slices produce a null types.Set. Repeats are collapsed; see the note on
// Uint64SliceToFrameworkSet.
func StringSliceToFrameworkSet(ctx context.Context, v []string) (types.Set, diag.Diagnostics) {
	if v == nil {
		return types.SetNull(types.StringType), nil
	}

	seen := make(map[string]bool, len(v))
	elems := make([]types.String, 0, len(v))
	for _, val := range v {
		if seen[val] {
			continue
		}
		seen[val] = true
		elems = append(elems, types.StringValue(val))
	}
	return types.SetValueFrom(ctx, types.StringType, elems)
}

// Uint64SliceFromFrameworkSet converts a types.Set (of int64) to []uint64.
// Null or unknown sets produce a nil slice.
func Uint64SliceFromFrameworkSet(ctx context.Context, v types.Set) ([]uint64, diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	var elems []types.Int64
	diags := v.ElementsAs(ctx, &elems, false)
	if diags.HasError() {
		return nil, diags
	}
	result := make([]uint64, len(elems))
	for i, e := range elems {
		result[i] = uint64(e.ValueInt64())
	}
	return result, nil
}

// Uint64SliceToFrameworkSet converts a []uint64 to a types.Set (of int64).
// A nil slice produces a null set.
//
// Repeats are collapsed. A types.Set rejects duplicate elements outright, and
// the API does return them: importing a real install produced three "Duplicate
// Set Element" failures on ids 13361, 9 and 16564. These attributes were lists
// until association collections became sets, which is what made a repeat fatal
// rather than merely odd. Collapsing here keeps the read faithful to set
// semantics, where a member is present or not.
func Uint64SliceToFrameworkSet(ctx context.Context, v []uint64) (types.Set, diag.Diagnostics) {
	if v == nil {
		return types.SetNull(types.Int64Type), nil
	}
	seen := make(map[uint64]bool, len(v))
	elems := make([]types.Int64, 0, len(v))
	for _, val := range v {
		if seen[val] {
			continue
		}
		seen[val] = true
		elems = append(elems, types.Int64Value(int64(val)))
	}
	return types.SetValueFrom(ctx, types.Int64Type, elems)
}

// Uint64SliceToFrameworkSetOrEmpty is Uint64SliceToFrameworkSet with a nil
// slice mapped to an EMPTY set instead of a null one.
//
// Use it for a membership collection whose API cannot express null. Kion's
// permission mappings store an empty member list and return it as JSON null, so
// null and empty are one state on the wire. Preserving the distinction in state
// is what made `user_groups_ids = []` in a configuration never converge: the
// read turned the API's null back into a null set, which never equals the
// configured empty set, so every plan re-proposed the same write. Collections
// the API does distinguish keep using Uint64SliceToFrameworkSet.
func Uint64SliceToFrameworkSetOrEmpty(ctx context.Context, v []uint64) (types.Set, diag.Diagnostics) {
	if v == nil {
		v = []uint64{}
	}
	return Uint64SliceToFrameworkSet(ctx, v)
}

// StringSliceToFrameworkSetOrEmpty is StringSliceToFrameworkSet with a nil
// slice mapped to an EMPTY set and blank elements dropped. Kion stores an empty
// compliance check region list as [""], so a configured `[]` came back with one
// element and failed the apply.
func StringSliceToFrameworkSetOrEmpty(ctx context.Context, v []string) (types.Set, diag.Diagnostics) {
	kept := make([]string, 0, len(v))
	for _, s := range v {
		if s != "" {
			kept = append(kept, s)
		}
	}
	return StringSliceToFrameworkSet(ctx, kept)
}

// --- Membership diff: for associations synced via paired add/remove endpoints ---

// Uint64SetDiff extracts the old (state) and new (plan) uint64 id sets and
// returns the ids to add (in new, not old) and to remove (in old, not new).
// Used by resources whose owner/member associations are updated through
// separate add/remove endpoints rather than the main update body.
func Uint64SetDiff(ctx context.Context, old, cur types.Set, diags *diag.Diagnostics) (add, remove []uint64) {
	o, d := Uint64SliceFromFrameworkSet(ctx, old)
	diags.Append(d...)
	n, d2 := Uint64SliceFromFrameworkSet(ctx, cur)
	diags.Append(d2...)
	return subtractUint64(n, o), subtractUint64(o, n)
}

// subtractUint64 returns the elements of a that are not in b (set difference).
func subtractUint64(a, b []uint64) []uint64 {
	inB := make(map[uint64]struct{}, len(b))
	for _, x := range b {
		inB[x] = struct{}{}
	}
	var out []uint64
	for _, x := range a {
		if _, ok := inB[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

// Int64SetDiff is the []int64 analog of Uint64SetDiff, for member endpoints that
// take []int64 (e.g. UpdateUGroupUsers).
func Int64SetDiff(ctx context.Context, old, cur types.Set, diags *diag.Diagnostics) (add, remove []int64) {
	o, d := Uint64SliceFromFrameworkSet(ctx, old)
	diags.Append(d...)
	n, d2 := Uint64SliceFromFrameworkSet(ctx, cur)
	diags.Append(d2...)
	return toInt64(subtractUint64(n, o)), toInt64(subtractUint64(o, n))
}

func toInt64(v []uint64) []int64 {
	if v == nil {
		return nil
	}
	out := make([]int64, len(v))
	for i, x := range v {
		out[i] = int64(x)
	}
	return out
}

// DedupeInt64 removes repeats while preserving first-seen order.
//
// A types.Set refuses duplicate elements, and Kion returns the same owner id
// twice on a few records, which failed the import outright rather than merely
// looking odd.
func DedupeInt64(v []int64) []int64 {
	if len(v) < 2 {
		return v
	}
	seen := make(map[int64]bool, len(v))
	out := v[:0:0]
	for _, x := range v {
		if seen[x] {
			continue
		}
		seen[x] = true
		out = append(out, x)
	}
	return out
}
