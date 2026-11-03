package filter_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/filter"
)

// stringList builds a types.List of strings for filter Values.
func stringList(t *testing.T, vals ...string) types.List {
	t.Helper()
	elems := make([]types.String, len(vals))
	for i, v := range vals {
		elems[i] = types.StringValue(v)
	}
	list, diags := types.ListValueFrom(t.Context(), types.StringType, elems)
	require.False(t, diags.HasError(), "stringList diags: %v", diags)
	return list
}

func TestMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		filters   func(t *testing.T) []filter.Model
		row       map[string]any
		wantMatch bool
		wantErr   bool
	}{
		{
			name:      "no filters matches everything",
			filters:   func(*testing.T) []filter.Model { return nil },
			row:       map[string]any{"key": "anything"},
			wantMatch: true,
		},
		{
			name: "single filter exact hit",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("key"), Values: stringList(t, "Owner Email")}}
			},
			row:       map[string]any{"key": "Owner Email", "value": "alice@example.com"},
			wantMatch: true,
		},
		{
			name: "single filter miss",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("key"), Values: stringList(t, "Application Name")}}
			},
			row:       map[string]any{"key": "Owner Email"},
			wantMatch: false,
		},
		{
			name: "multi-value OR",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("key"), Values: stringList(t, "Owner Email", "Regulatory", "App")}}
			},
			row:       map[string]any{"key": "Regulatory"},
			wantMatch: true,
		},
		{
			name: "multi-filter AND all hit",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{
					{Name: types.StringValue("key"), Values: stringList(t, "Regulatory")},
					{Name: types.StringValue("value"), Values: stringList(t, "GxP")},
				}
			},
			row:       map[string]any{"key": "Regulatory", "value": "GxP"},
			wantMatch: true,
		},
		{
			name: "multi-filter AND one misses",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{
					{Name: types.StringValue("key"), Values: stringList(t, "Regulatory")},
					{Name: types.StringValue("value"), Values: stringList(t, "Non-GxP")},
				}
			},
			row:       map[string]any{"key": "Regulatory", "value": "GxP"},
			wantMatch: false,
		},
		{
			name: "regex hit",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("value"), Values: stringList(t, `.*@biogen\.com$`), Regex: types.BoolValue(true)}}
			},
			row:       map[string]any{"value": "alice@biogen.com"},
			wantMatch: true,
		},
		{
			name: "regex miss",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("value"), Values: stringList(t, `.*@biogen\.com$`), Regex: types.BoolValue(true)}}
			},
			row:       map[string]any{"value": "bob@example.org"},
			wantMatch: false,
		},
		{
			name: "regex invalid pattern errors",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("value"), Values: stringList(t, "[unterminated"), Regex: types.BoolValue(true)}}
			},
			row:     map[string]any{"value": "anything"},
			wantErr: true,
		},
		{
			name: "dotted key into map",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("owner.id"), Values: stringList(t, "42")}}
			},
			row:       map[string]any{"owner": map[string]any{"id": "42", "name": "alice"}},
			wantMatch: true,
		},
		{
			name: "dotted key walks array",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("groups.name"), Values: stringList(t, "admins")}}
			},
			row:       map[string]any{"groups": []any{map[string]any{"name": "developers"}, map[string]any{"name": "admins"}}},
			wantMatch: true,
		},
		{
			name: "leaf is array errors",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("tags"), Values: stringList(t, "a")}}
			},
			row:     map[string]any{"tags": []any{"a", "b"}},
			wantErr: true,
		},
		{
			name: "missing key is a miss",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("nonexistent"), Values: stringList(t, "x")}}
			},
			row:       map[string]any{"key": "Owner Email"},
			wantMatch: false,
		},
		{
			name: "numeric value compared as string",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("id"), Values: stringList(t, "42")}}
			},
			row:       map[string]any{"id": int64(42)},
			wantMatch: true,
		},
		{
			name: "empty values errors",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("key"), Values: stringList(t)}}
			},
			row:     map[string]any{"key": "anything"},
			wantErr: true,
		},
		{
			name: "null filter name errors",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringNull(), Values: stringList(t, "x")}}
			},
			row:     map[string]any{"key": "anything"},
			wantErr: true,
		},
		{
			name: "unknown filter name errors",
			filters: func(t *testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringUnknown(), Values: stringList(t, "x")}}
			},
			row:     map[string]any{"key": "anything"},
			wantErr: true,
		},
		{
			name: "null values errors",
			filters: func(*testing.T) []filter.Model {
				return []filter.Model{{Name: types.StringValue("key"), Values: types.ListNull(types.StringType)}}
			},
			row:     map[string]any{"key": "anything"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			matched, diags := filter.Match(t.Context(), tc.filters(t), tc.row)
			if tc.wantErr {
				require.True(t, diags.HasError(), "expected diagnostics error")
				return
			}
			require.False(t, diags.HasError(), "unexpected diags: %v", diags)
			assert.Equal(t, tc.wantMatch, matched)
		})
	}
}
