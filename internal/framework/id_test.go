package framework_test

import (
	"testing"

	"terraform-provider-kion/internal/framework"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToInt64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		in      string
		want    int64
		wantErr bool
	}{
		{"valid", "42", 42, false},
		{"negative", "-7", -7, false},
		{"non-numeric", "abc", 0, true},
		{"overflow", "99999999999999999999", 0, true},
		{"empty", "", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, diags := framework.StringToInt64(tc.in)
			if tc.wantErr {
				require.True(t, diags.HasError())
				return
			}
			require.False(t, diags.HasError(), "diags: %v", diags)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestInt64ToString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "42", framework.Int64ToString(42))
	assert.Equal(t, "-7", framework.Int64ToString(-7))
}

func TestUint64ToString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "42", framework.Uint64ToString(42))
	assert.Equal(t, "0", framework.Uint64ToString(0))
}
