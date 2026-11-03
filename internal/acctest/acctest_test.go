package acctest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckPreCheck(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"missing url", map[string]string{"KION_API_KEY": "k"}, "KION_API_URL"},
		{"missing creds", map[string]string{"KION_API_URL": "https://x"}, "KION_API_KEY or KION_AUTH_TOKEN"},
		{"url + key ok", map[string]string{"KION_API_URL": "https://x", "KION_API_KEY": "k"}, ""},
		{"url + token ok", map[string]string{"KION_API_URL": "https://x", "KION_AUTH_TOKEN": "t"}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			getenv := func(k string) string { return tc.env[k] }
			err := checkPreCheck(getenv)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestRandomWithPrefix(t *testing.T) {
	t.Parallel()

	a := RandomWithPrefix(ResourcePrefix)
	b := RandomWithPrefix(ResourcePrefix)

	require.True(t, strings.HasPrefix(a, ResourcePrefix+"-"))
	require.Len(t, a, len(ResourcePrefix)+1+8) // prefix + "-" + 8 random chars
	require.NotEqual(t, a, b)
}

func TestProtoV6ProviderFactories_Kion(t *testing.T) {
	t.Parallel()

	factory, ok := ProtoV6ProviderFactories["kion"]
	require.True(t, ok)

	srv, err := factory()
	require.NoError(t, err)
	require.NotNil(t, srv)
}
