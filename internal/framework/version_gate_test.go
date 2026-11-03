package framework_test

import (
	"testing"

	"terraform-provider-kion/internal/conns"
	"terraform-provider-kion/internal/framework"

	"github.com/stretchr/testify/assert"
)

func TestRequireMinKionVersion(t *testing.T) {
	t.Parallel()

	minVer := conns.MustParseKionVersion("3.16.0")

	newClient := func(v string, detected bool) *conns.KionClient {
		c := &conns.KionClient{VersionDetected: detected}
		if v != "" {
			c.Version = conns.MustParseKionVersion(v)
		}
		return c
	}

	cases := []struct {
		name      string
		client    *conns.KionClient
		wantError bool
		wantWarn  bool
	}{
		{"supported exact", newClient("3.16.0", true), false, false},
		{"supported newer", newClient("3.17.2", true), false, false},
		{"unsupported older", newClient("3.15.4", true), true, false},
		{"unsupported much older", newClient("3.14.0", true), true, false},
		{"undetected warns and allows", newClient("", false), false, true},
		{"nil client warns and allows", nil, false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			diags := framework.RequireMinKionVersion(tc.client, minVer, "kion_ou_note")
			assert.Equal(t, tc.wantError, diags.HasError(), "HasError (%v)", diags)
			assert.Equal(t, tc.wantWarn, diags.WarningsCount() > 0, "hasWarning (%v)", diags)
		})
	}
}
