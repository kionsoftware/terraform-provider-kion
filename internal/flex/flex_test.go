package flex_test

import (
	"context"
	"testing"

	"terraform-provider-kion/internal/flex"

	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
	"github.com/stretchr/testify/assert"
)

// --- String ---

func TestStringValueFromFramework(t *testing.T) {
	assert.Equal(t, "hello", flex.StringValueFromFramework(types.StringValue("hello")))
	assert.Equal(t, "", flex.StringValueFromFramework(types.StringValue("")))
}

func TestOptStringFromFramework(t *testing.T) {
	opt := flex.OptStringFromFramework(types.StringValue("hello"))
	assert.True(t, opt.IsSet())
	assert.Equal(t, "hello", opt.Value)

	opt = flex.OptStringFromFramework(types.StringNull())
	assert.False(t, opt.IsSet())

	opt = flex.OptStringFromFramework(types.StringUnknown())
	assert.False(t, opt.IsSet())
}

func TestStringToFramework(t *testing.T) {
	assert.Equal(t, types.StringValue("hello"), flex.StringToFramework("hello"))
}

func TestOptStringToFramework(t *testing.T) {
	assert.Equal(t, types.StringValue("hello"), flex.OptStringToFramework(generated.OptString{Value: "hello", Set: true}))
	assert.Equal(t, types.StringNull(), flex.OptStringToFramework(generated.OptString{}))
}

// A create_only secret is captured through this: a public OAuth client answers
// "" for its secret, which must store as null.
func TestOptStringToFrameworkNullIfEmpty(t *testing.T) {
	assert.Equal(t, types.StringValue("s3cr3t"), flex.OptStringToFrameworkNullIfEmpty(generated.OptString{Value: "s3cr3t", Set: true}))
	assert.Equal(t, types.StringNull(), flex.OptStringToFrameworkNullIfEmpty(generated.OptString{Value: "", Set: true}))
	assert.Equal(t, types.StringNull(), flex.OptStringToFrameworkNullIfEmpty(generated.OptString{}))
	assert.Equal(t, types.StringNull(), flex.OptStringToFrameworkNullIfEmpty(generated.NewOptString("")))
}

// --- Int64 ---

func TestInt64ValueFromFramework(t *testing.T) {
	assert.Equal(t, int64(42), flex.Int64ValueFromFramework(types.Int64Value(42)))
}

func TestUint64FromFramework(t *testing.T) {
	assert.Equal(t, uint64(42), flex.Uint64FromFramework(types.Int64Value(42)))
}

func TestOptInt64FromFramework(t *testing.T) {
	opt := flex.OptInt64FromFramework(types.Int64Value(42))
	assert.True(t, opt.IsSet())
	assert.Equal(t, int64(42), opt.Value)

	opt = flex.OptInt64FromFramework(types.Int64Null())
	assert.False(t, opt.IsSet())
}

func TestOptUint64FromFramework(t *testing.T) {
	opt := flex.OptUint64FromFramework(types.Int64Value(42))
	assert.True(t, opt.IsSet())
	assert.Equal(t, uint64(42), opt.Value)

	opt = flex.OptUint64FromFramework(types.Int64Null())
	assert.False(t, opt.IsSet())
}

func TestInt64ToFramework(t *testing.T) {
	assert.Equal(t, types.Int64Value(42), flex.Int64ToFramework(42))
}

func TestUint64ToFramework(t *testing.T) {
	assert.Equal(t, types.Int64Value(42), flex.Uint64ToFramework(42))
}

func TestOptInt64ToFramework(t *testing.T) {
	assert.Equal(t, types.Int64Value(42), flex.OptInt64ToFramework(generated.OptInt64{Value: 42, Set: true}))
	assert.Equal(t, types.Int64Null(), flex.OptInt64ToFramework(generated.OptInt64{}))
}

func TestOptUint64ToFramework(t *testing.T) {
	assert.Equal(t, types.Int64Value(42), flex.OptUint64ToFramework(generated.OptUint64{Value: 42, Set: true}))
	assert.Equal(t, types.Int64Null(), flex.OptUint64ToFramework(generated.OptUint64{}))
}

// --- Bool ---

func TestBoolValueFromFramework(t *testing.T) {
	assert.Equal(t, true, flex.BoolValueFromFramework(types.BoolValue(true)))
	assert.Equal(t, false, flex.BoolValueFromFramework(types.BoolValue(false)))
}

func TestOptBoolFromFramework(t *testing.T) {
	opt := flex.OptBoolFromFramework(types.BoolValue(true))
	assert.True(t, opt.IsSet())
	assert.True(t, opt.Value)

	opt = flex.OptBoolFromFramework(types.BoolNull())
	assert.False(t, opt.IsSet())
}

func TestBoolToFramework(t *testing.T) {
	assert.Equal(t, types.BoolValue(true), flex.BoolToFramework(true))
}

func TestOptBoolToFramework(t *testing.T) {
	assert.Equal(t, types.BoolValue(true), flex.OptBoolToFramework(generated.OptBool{Value: true, Set: true}))
	assert.Equal(t, types.BoolNull(), flex.OptBoolToFramework(generated.OptBool{}))
}

// --- List ---

func TestUint64SliceFromFramework(t *testing.T) {
	ctx := context.Background()

	list, diags := types.ListValueFrom(ctx, types.Int64Type, []int64{1, 2, 3})
	assert.False(t, diags.HasError())

	result, diags := flex.Uint64SliceFromFramework(ctx, list)
	assert.False(t, diags.HasError())
	assert.Equal(t, []uint64{1, 2, 3}, result)

	result, diags = flex.Uint64SliceFromFramework(ctx, types.ListNull(types.Int64Type))
	assert.False(t, diags.HasError())
	assert.Nil(t, result)
}

func TestStringSliceFromFramework(t *testing.T) {
	ctx := context.Background()

	list, diags := types.ListValueFrom(ctx, types.StringType, []string{"a", "b"})
	assert.False(t, diags.HasError())

	result, diags := flex.StringSliceFromFramework(ctx, list)
	assert.False(t, diags.HasError())
	assert.Equal(t, []string{"a", "b"}, result)

	result, diags = flex.StringSliceFromFramework(ctx, types.ListNull(types.StringType))
	assert.False(t, diags.HasError())
	assert.Nil(t, result)
}

func TestUint64SliceToFramework(t *testing.T) {
	ctx := context.Background()

	result, diags := flex.Uint64SliceToFramework(ctx, []uint64{1, 2, 3})
	assert.False(t, diags.HasError())
	assert.False(t, result.IsNull())

	result, diags = flex.Uint64SliceToFramework(ctx, nil)
	assert.False(t, diags.HasError())
	assert.True(t, result.IsNull())
}

func TestStringSliceToFramework(t *testing.T) {
	ctx := context.Background()

	result, diags := flex.StringSliceToFramework(ctx, []string{"a", "b"})
	assert.False(t, diags.HasError())
	assert.False(t, result.IsNull())

	result, diags = flex.StringSliceToFramework(ctx, nil)
	assert.False(t, diags.HasError())
	assert.True(t, result.IsNull())
}

// Kion's [""] and null both mean "no elements", and must read as an empty set.
func TestStringSliceToFrameworkSetOrEmpty(t *testing.T) {
	ctx := context.Background()

	for _, in := range [][]string{nil, {""}} {
		result, diags := flex.StringSliceToFrameworkSetOrEmpty(ctx, in)
		assert.False(t, diags.HasError())
		assert.False(t, result.IsNull())
		assert.Empty(t, result.Elements())
	}

	result, diags := flex.StringSliceToFrameworkSetOrEmpty(ctx, []string{"us-east-1", ""})
	assert.False(t, diags.HasError())
	assert.Len(t, result.Elements(), 1)
}

// unknown lists and element-type mismatches exercise the branches the basic
// value/null cases above don't reach.
func TestUint64SliceFromFramework_unknownAndError(t *testing.T) {
	ctx := context.Background()

	result, diags := flex.Uint64SliceFromFramework(ctx, types.ListUnknown(types.Int64Type))
	assert.False(t, diags.HasError())
	assert.Nil(t, result)

	// A String-typed list can't decode into []types.Int64, so ElementsAs errors.
	mismatch, d := types.ListValueFrom(ctx, types.StringType, []string{"a"})
	assert.False(t, d.HasError())
	_, diags = flex.Uint64SliceFromFramework(ctx, mismatch)
	assert.True(t, diags.HasError())
}

func TestStringSliceFromFramework_unknownAndError(t *testing.T) {
	ctx := context.Background()

	result, diags := flex.StringSliceFromFramework(ctx, types.ListUnknown(types.StringType))
	assert.False(t, diags.HasError())
	assert.Nil(t, result)

	mismatch, d := types.ListValueFrom(ctx, types.Int64Type, []int64{1})
	assert.False(t, d.HasError())
	_, diags = flex.StringSliceFromFramework(ctx, mismatch)
	assert.True(t, diags.HasError())
}

// --- Set ---

func TestStringSliceFromFrameworkSet(t *testing.T) {
	ctx := context.Background()

	set, diags := types.SetValueFrom(ctx, types.StringType, []string{"a", "b"})
	assert.False(t, diags.HasError())

	result, diags := flex.StringSliceFromFrameworkSet(ctx, set)
	assert.False(t, diags.HasError())
	assert.ElementsMatch(t, []string{"a", "b"}, result)

	result, diags = flex.StringSliceFromFrameworkSet(ctx, types.SetNull(types.StringType))
	assert.False(t, diags.HasError())
	assert.Nil(t, result)

	result, diags = flex.StringSliceFromFrameworkSet(ctx, types.SetUnknown(types.StringType))
	assert.False(t, diags.HasError())
	assert.Nil(t, result)

	mismatch, d := types.SetValueFrom(ctx, types.Int64Type, []int64{1})
	assert.False(t, d.HasError())
	_, diags = flex.StringSliceFromFrameworkSet(ctx, mismatch)
	assert.True(t, diags.HasError())
}

func TestStringSliceToFrameworkSet(t *testing.T) {
	ctx := context.Background()

	result, diags := flex.StringSliceToFrameworkSet(ctx, []string{"a", "b"})
	assert.False(t, diags.HasError())
	assert.False(t, result.IsNull())

	result, diags = flex.StringSliceToFrameworkSet(ctx, nil)
	assert.False(t, diags.HasError())
	assert.True(t, result.IsNull())
}

// --- OptNilBool ---

func TestOptNilBoolFromFramework(t *testing.T) {
	v, ok := flex.OptNilBoolFromFramework(types.BoolValue(true)).Get()
	assert.True(t, ok)
	assert.True(t, v)

	_, ok = flex.OptNilBoolFromFramework(types.BoolNull()).Get()
	assert.False(t, ok)

	_, ok = flex.OptNilBoolFromFramework(types.BoolUnknown()).Get()
	assert.False(t, ok)
}

func TestOptNilBoolToFramework(t *testing.T) {
	assert.Equal(t, types.BoolValue(true), flex.OptNilBoolToFramework(generated.NewOptNilBool(true)))
	assert.Equal(t, types.BoolNull(), flex.OptNilBoolToFramework(generated.OptNilBool{}))
}

// --- OptNilUint64 / NilUint64 ---

func TestOptNilUint64FromFramework(t *testing.T) {
	v, ok := flex.OptNilUint64FromFramework(types.Int64Value(42)).Get()
	assert.True(t, ok)
	assert.Equal(t, uint64(42), v)

	_, ok = flex.OptNilUint64FromFramework(types.Int64Null()).Get()
	assert.False(t, ok)

	_, ok = flex.OptNilUint64FromFramework(types.Int64Unknown()).Get()
	assert.False(t, ok)
}

func TestOptNilUint64ToFramework(t *testing.T) {
	assert.Equal(t, types.Int64Value(42), flex.OptNilUint64ToFramework(generated.NewOptNilUint64(42)))
	assert.Equal(t, types.Int64Null(), flex.OptNilUint64ToFramework(generated.OptNilUint64{}))
}

func TestNilUint64FromFramework(t *testing.T) {
	assert.Equal(t, generated.NilUint64{Value: 42}, flex.NilUint64FromFramework(types.Int64Value(42)))
	assert.Equal(t, generated.NilUint64{Null: true}, flex.NilUint64FromFramework(types.Int64Null()))
	assert.Equal(t, generated.NilUint64{Null: true}, flex.NilUint64FromFramework(types.Int64Unknown()))
}

func TestNilUint64ToFramework(t *testing.T) {
	assert.Equal(t, types.Int64Value(42), flex.NilUint64ToFramework(generated.NilUint64{Value: 42}))
	assert.Equal(t, types.Int64Null(), flex.NilUint64ToFramework(generated.NilUint64{Null: true}))
}

// --- OptStringToFrameworkLegacy ---

func TestOptStringToFrameworkLegacy(t *testing.T) {
	assert.Equal(t, types.StringValue("hello"), flex.OptStringToFrameworkLegacy(generated.OptString{Value: "hello", Set: true}))
	// Legacy semantics: unset -> empty string, not null.
	assert.Equal(t, types.StringValue(""), flex.OptStringToFrameworkLegacy(generated.OptString{}))
}

// --- NilBool ---

func TestNilBoolFromFramework(t *testing.T) {
	assert.Equal(t, generated.NilBool{Value: true}, flex.NilBoolFromFramework(types.BoolValue(true)))
	assert.Equal(t, generated.NilBool{Null: true}, flex.NilBoolFromFramework(types.BoolNull()))
	assert.Equal(t, generated.NilBool{Null: true}, flex.NilBoolFromFramework(types.BoolUnknown()))
}

func TestNilBoolToFramework(t *testing.T) {
	assert.Equal(t, types.BoolValue(true), flex.NilBoolToFramework(generated.NilBool{Value: true}))
	assert.Equal(t, types.BoolNull(), flex.NilBoolToFramework(generated.NilBool{Null: true}))
}

// --- Normalized (JSON) <-> jx.Raw ---

func TestNormalizedFromFramework(t *testing.T) {
	assert.Equal(t, jx.Raw(`{"a":1}`), flex.NormalizedFromFramework(jsontypes.NewNormalizedValue(`{"a":1}`)))
	assert.Nil(t, flex.NormalizedFromFramework(jsontypes.NewNormalizedNull()))
	assert.Nil(t, flex.NormalizedFromFramework(jsontypes.NewNormalizedUnknown()))
}

func TestNormalizedToFramework(t *testing.T) {
	assert.Equal(t, jsontypes.NewNormalizedValue(`{"a":1}`), flex.NormalizedToFramework(jx.Raw(`{"a":1}`)))
	assert.Equal(t, jsontypes.NewNormalizedNull(), flex.NormalizedToFramework(jx.Raw(nil)))
}

// --- GCPRoleLaunchStage (enum) ---

func TestOptGCPRoleLaunchStageFromFramework(t *testing.T) {
	assert.Equal(t, generated.OptGCPRoleLaunchStage{Value: generated.GCPRoleLaunchStage(2), Set: true}, flex.OptGCPRoleLaunchStageFromFramework(types.Int64Value(2)))
	assert.False(t, flex.OptGCPRoleLaunchStageFromFramework(types.Int64Null()).Set)
}

func TestOptGCPRoleLaunchStageToFramework(t *testing.T) {
	assert.Equal(t, types.Int64Value(2), flex.OptGCPRoleLaunchStageToFramework(generated.OptGCPRoleLaunchStage{Value: 2, Set: true}))
	assert.Equal(t, types.Int64Null(), flex.OptGCPRoleLaunchStageToFramework(generated.OptGCPRoleLaunchStage{}))
}

// --- OptNullString / OptNullTime (SQL-null wrappers) ---

func TestOptNullStringRoundTrip(t *testing.T) {
	got := flex.OptNullStringFromFramework(types.StringValue("hi"))
	assert.True(t, got.Set)
	assert.Equal(t, types.StringValue("hi"), flex.OptNullStringToFramework(got))
	assert.False(t, flex.OptNullStringFromFramework(types.StringNull()).Set)
	assert.Equal(t, types.StringNull(), flex.OptNullStringToFramework(generated.OptNullString{}))
}

func TestOptNullTimeRoundTrip(t *testing.T) {
	got := flex.OptNullTimeFromFramework(types.StringValue("2025-01-02T03:04:05Z"))
	assert.True(t, got.Set)
	assert.Equal(t, types.StringValue("2025-01-02T03:04:05Z"), flex.OptNullTimeToFramework(got))
	assert.False(t, flex.OptNullTimeFromFramework(types.StringValue("not-a-time")).Set)
	assert.Equal(t, types.StringNull(), flex.OptNullTimeToFramework(generated.OptNullTime{}))
}

func TestDatecodeToFramework(t *testing.T) {
	t.Parallel()
	// The private reads return 202406 where the public schema validates
	// ^\d{4}-(?:0[1-9]|1[0-2])$, so a plain decimal format could never satisfy
	// the provider's own validator.
	assert.Equal(t, "2024-06", flex.DatecodeToFramework(202406).ValueString())
	assert.Equal(t, "2019-01", flex.DatecodeToFramework(201901).ValueString())
	assert.Equal(t, "2024-12", flex.DatecodeToFramework(202412).ValueString())
	// 0 is unset, not the year 0.
	assert.True(t, flex.DatecodeToFramework(0).IsNull())
	// Not a datecode: hand it back so the validator reports the real value
	// rather than a shape we invented.
	assert.Equal(t, "42", flex.DatecodeToFramework(42).ValueString())
	assert.Equal(t, "202413", flex.DatecodeToFramework(202413).ValueString())
}

// TestSliceToFrameworkSetDedupes covers the API returning a member twice. A
// types.Set rejects duplicates, so a faithful copy of the wire order made the
// resource unimportable; three resources failed this way on a real install.
func TestSliceToFrameworkSetDedupes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	got, diags := flex.Uint64SliceToFrameworkSet(ctx, []uint64{9, 13361, 9, 16564, 13361})
	assert.False(t, diags.HasError())
	assert.Len(t, got.Elements(), 3)

	strs, diags := flex.StringSliceToFrameworkSet(ctx, []string{"a", "b", "a"})
	assert.False(t, diags.HasError())
	assert.Len(t, strs.Elements(), 2)

	// Nil still means null, not empty.
	null, _ := flex.Uint64SliceToFrameworkSet(ctx, nil)
	assert.True(t, null.IsNull())
}

// TestCanonicalJSONMatchesJsonencode covers the byte equality a generated
// configuration needs: terraform's jsonencode emits compact JSON with sorted
// keys, and the read must produce the same bytes or every plan reports a change.
func TestCanonicalJSONMatchesJsonencode(t *testing.T) {
	t.Parallel()
	got := flex.NormalizedStringToFramework("{\n  \"b\": 2,\n  \"a\": [1, 2]\n}")
	assert.Equal(t, `{"a":[1,2],"b":2}`, got.ValueString(), "keys sorted, whitespace removed")

	// Not JSON: returned untouched. CloudFormation templates are often YAML.
	yaml := "---\nAWSTemplateFormatVersion: 2010-09-09\n"
	assert.Equal(t, yaml, flex.NormalizedStringToFramework(yaml).ValueString())

	// Already canonical: unchanged.
	assert.Equal(t, `{"a":1}`, flex.NormalizedStringToFramework(`{"a":1}`).ValueString())
}
