package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// env returns a getenv func backed by a map, so resolution can be tested
// without touching the real process environment.
func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestResolveProviderConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  kionProviderModel
		env     map[string]string
		wantErr string // substring of the first diagnostic summary; "" means no error
		want    resolvedConfig
	}{
		{
			name: "new attribute names",
			config: kionProviderModel{
				APIURL: types.StringValue("https://kion.example.com"),
				APIKey: types.StringValue("key123"),
			},
			want: resolvedConfig{
				apiKey:    "key123",
				serverURL: "https://kion.example.com/api",
			},
		},
		{
			name: "deprecated aliases",
			config: kionProviderModel{
				URL:    types.StringValue("https://kion.example.com"),
				Apikey: types.StringValue("legacykey"),
			},
			want: resolvedConfig{
				apiKey:    "legacykey",
				serverURL: "https://kion.example.com/api",
			},
		},
		{
			name: "environment fallback",
			env: map[string]string{
				"KION_API_URL": "https://env.example.com",
				"KION_API_KEY": "envkey",
			},
			want: resolvedConfig{
				apiKey:    "envkey",
				serverURL: "https://env.example.com/api",
			},
		},
		{
			name: "auth token instead of api key",
			config: kionProviderModel{
				APIURL:    types.StringValue("https://kion.example.com"),
				AuthToken: types.StringValue("tok"),
			},
			want: resolvedConfig{
				authToken: "tok",
				serverURL: "https://kion.example.com/api",
			},
		},
		{
			name: "new names win over aliases and env",
			config: kionProviderModel{
				APIURL: types.StringValue("https://new.example.com"),
				URL:    types.StringValue("https://old.example.com"),
				APIKey: types.StringValue("newkey"),
				Apikey: types.StringValue("oldkey"),
			},
			env: map[string]string{
				"KION_API_URL": "https://env.example.com",
				"KION_API_KEY": "envkey",
			},
			want: resolvedConfig{
				apiKey:    "newkey",
				serverURL: "https://new.example.com/api",
			},
		},
		{
			name:    "missing api url",
			config:  kionProviderModel{APIKey: types.StringValue("key")},
			wantErr: "Missing API URL",
		},
		{
			name:    "missing authentication",
			config:  kionProviderModel{APIURL: types.StringValue("https://kion.example.com")},
			wantErr: "Missing Authentication",
		},
		{
			name: "skip ssl via attribute",
			config: kionProviderModel{
				APIURL:            types.StringValue("https://kion.example.com"),
				APIKey:            types.StringValue("key"),
				SkipSSLValidation: types.BoolValue(true),
			},
			want: resolvedConfig{
				apiKey:    "key",
				skipSSL:   true,
				serverURL: "https://kion.example.com/api",
			},
		},
		{
			name: "skip ssl via env",
			config: kionProviderModel{
				APIURL: types.StringValue("https://kion.example.com"),
				APIKey: types.StringValue("key"),
			},
			env: map[string]string{"KION_SKIP_SSL_VALIDATION": "1"},
			want: resolvedConfig{
				apiKey:    "key",
				skipSSL:   true,
				serverURL: "https://kion.example.com/api",
			},
		},
		{
			name: "empty apipath disables /api suffix",
			config: kionProviderModel{
				APIURL:  types.StringValue("https://kion.example.com"),
				APIKey:  types.StringValue("key"),
				Apipath: types.StringValue(""),
			},
			want: resolvedConfig{
				apiKey:    "key",
				serverURL: "https://kion.example.com",
			},
		},
		{
			name: "custom apipath and trailing slash trimming",
			config: kionProviderModel{
				APIURL:  types.StringValue("https://kion.example.com/"),
				APIKey:  types.StringValue("key"),
				Apipath: types.StringValue("/custom/"),
			},
			want: resolvedConfig{
				apiKey:    "key",
				serverURL: "https://kion.example.com/custom",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, diags := resolveProviderConfig(tt.config, env(tt.env))

			if tt.wantErr != "" {
				require.True(t, diags.HasError(), "expected an error diagnostic")
				assert.Contains(t, diags[0].Summary(), tt.wantErr)
				return
			}

			require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSecuritySourceToken(t *testing.T) {
	t.Parallel()

	// Auth token is preferred over the API key when both are present.
	tok, err := (&securitySource{apiKey: "key", authToken: "tok"}).Token(context.Background(), generated.OperationName(""))
	require.NoError(t, err)
	assert.Equal(t, "Bearer tok", tok.APIKey)

	// Falls back to the API key when no auth token is set.
	tok, err = (&securitySource{apiKey: "key"}).Token(context.Background(), generated.OperationName(""))
	require.NoError(t, err)
	assert.Equal(t, "Bearer key", tok.APIKey)
}

func TestProviderMetadata(t *testing.T) {
	t.Parallel()

	resp := &provider.MetadataResponse{}
	New()().Metadata(context.Background(), provider.MetadataRequest{}, resp)
	assert.Equal(t, "kion", resp.TypeName)
}

func TestProviderResourcesAndDataSources(t *testing.T) {
	t.Parallel()

	// Collecting the factories exercises ServicePackages() and the registration
	// loops without needing a configured client (the factories aren't invoked).
	p := New()()
	ctx := context.Background()
	assert.NotEmpty(t, p.Resources(ctx), "expected at least one resource factory")
	assert.NotEmpty(t, p.DataSources(ctx), "expected at least one data source factory")
}

func TestProviderSchema(t *testing.T) {
	t.Parallel()

	resp := &provider.SchemaResponse{}
	New()().Schema(context.Background(), provider.SchemaRequest{}, resp)

	require.False(t, resp.Diagnostics.HasError())
	for _, attr := range []string{"api_url", "api_key", "auth_token", "skip_ssl_validation", "url", "apikey", "apipath"} {
		assert.Contains(t, resp.Schema.Attributes, attr, "schema missing attribute %q", attr)
	}
}
