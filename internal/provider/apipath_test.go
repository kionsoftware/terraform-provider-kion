package provider

import (
	"maps"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The acceptance tests configure the provider entirely from the environment and
// write no provider block, so without KION_APIPATH they cannot reach an install
// that serves its API at the root -- every request goes to /api and 404s.
//
// "set but empty" is the case that matters and the one os.Getenv cannot express
// on its own: it selects the root, where unset keeps the "/api" default.
func TestResolveProviderConfig_APIPath(t *testing.T) {
	t.Parallel()

	base := map[string]string{"KION_API_URL": "http://localhost:8081", "KION_API_KEY": "k"}

	cases := []struct {
		name    string
		env     map[string]string
		apipath types.String
		want    string
	}{
		{"unset keeps /api", nil, types.StringNull(), "http://localhost:8081/api"},
		{
			"set empty selects root",
			map[string]string{"KION_APIPATH__isset": "1", "KION_APIPATH": ""},
			types.StringNull(),
			"http://localhost:8081",
		},
		{
			"set non-empty is used",
			map[string]string{"KION_APIPATH__isset": "1", "KION_APIPATH": "/gateway"},
			types.StringNull(),
			"http://localhost:8081/gateway",
		},
		// An explicit provider block still wins over the environment, matching
		// every other attribute's precedence.
		{
			"config beats env",
			map[string]string{"KION_APIPATH__isset": "1", "KION_APIPATH": "/gateway"},
			types.StringValue(""),
			"http://localhost:8081",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{}
			maps.Copy(env, base)
			maps.Copy(env, tc.env)
			got, diags := resolveProviderConfig(
				kionProviderModel{Apipath: tc.apipath},
				func(k string) string { return env[k] },
			)
			require.False(t, diags.HasError(), "%v", diags)
			assert.Equal(t, tc.want, got.serverURL)
		})
	}
}
