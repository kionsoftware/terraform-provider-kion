// Package provider implements the Kion Terraform provider.
package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"time"

	"terraform-provider-kion/internal/conns"

	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = (*kionProvider)(nil)

// New returns a new instance of the Kion provider.
func New() func() provider.Provider {
	return func() provider.Provider {
		return &kionProvider{}
	}
}

type kionProvider struct{}

type kionProviderModel struct {
	APIURL            types.String `tfsdk:"api_url"`
	APIKey            types.String `tfsdk:"api_key"`
	AuthToken         types.String `tfsdk:"auth_token"`
	SkipSSLValidation types.Bool   `tfsdk:"skip_ssl_validation"`
	// Backward-compatible aliases from the old provider
	URL     types.String `tfsdk:"url"`
	Apikey  types.String `tfsdk:"apikey"`
	Apipath types.String `tfsdk:"apipath"`
}

func (p *kionProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Kion cloud governance platform.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Description: "Kion API base URL. Can be set with KION_API_URL environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "Kion API key. Can be set with KION_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"auth_token": schema.StringAttribute{
				Description: "Kion authentication token. Can be set with KION_AUTH_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"skip_ssl_validation": schema.BoolAttribute{
				Description: "Skip SSL certificate verification. Can be set with KION_SKIP_SSL_VALIDATION environment variable.",
				Optional:    true,
			},
			// apipath is carried over from the old provider with the same
			// meaning, not deprecated: it is the only way to talk to a Kion
			// whose API is not under /api.
			"apipath": schema.StringAttribute{
				Description: "Base path of the Kion API, appended to api_url. Defaults to /api; set to \"\" for an API mounted at the root.",
				Optional:    true,
			},
			// Backward-compatible aliases from the old provider
			"url": schema.StringAttribute{
				Description: "Deprecated: use api_url instead. Kion API base URL.",
				Optional:    true,
			},
			"apikey": schema.StringAttribute{
				Description: "Deprecated: use api_key instead. Kion API key.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *kionProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config kionProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve and validate the configuration (new names > deprecated aliases >
	// environment variables). Extracted into a helper so the resolution logic is
	// unit-testable without the framework plumbing or a live API.
	resolved, diags := resolveProviderConfig(config, osGetenv)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey := resolved.apiKey
	authToken := resolved.authToken
	skipSSL := resolved.skipSSL
	serverURL := resolved.serverURL

	// Build an HTTP client honoring skip SSL validation.
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected HTTP transport",
			"The default HTTP transport is not *http.Transport, so the provider cannot configure the HTTP client.",
		)
		return
	}
	transport := defaultTransport.Clone()
	if skipSSL {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // user-requested
		}
	}
	httpClient := &http.Client{
		Transport: transport,
		// Whole-request, and a filtered data source fetches its entire
		// collection: 30s could not read /v3/azure-policy (51 MB, ~35s) at all.
		// Matches kion-import's 60s.
		Timeout: 60 * time.Second,
	}

	// Initialize the generated SDK client directly so we control the server URL
	// without the auto-appended "/api" suffix from kion.NewClient.
	sec := &securitySource{apiKey: apiKey, authToken: authToken}
	sdkClient, err := generated.NewClient(serverURL, sec,
		generated.WithClient(httpClient),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create Kion Client",
			"Could not create Kion SDK client: "+err.Error(),
		)
		return
	}

	kionClient := &conns.KionClient{
		Client:     sdkClient,
		APIURL:     serverURL,
		APIKey:     apiKey,
		AuthToken:  authToken,
		HTTPClient: httpClient,
	}

	// Best-effort: detect the Kion version (GET /api/version) so resources can
	// gate on a minimum version. Failure is non-fatal, gating degrades to a
	// warning and the API enforces support on its own.
	if err := kionClient.DetectVersion(ctx); err != nil {
		tflog.Warn(ctx, "could not detect Kion version", map[string]any{"error": err.Error()})
	} else {
		tflog.Info(ctx, "detected Kion version", map[string]any{"version": kionClient.Version.String()})
	}

	// Best-effort: detect whether the install funds projects with budgets or
	// spend plans (GET /v3/app-config), which decides the kion_project create
	// endpoint. Reading app-config needs a global settings permission, so a
	// failure is expected for a restricted token and leaves the mode to the
	// configuration.
	if err := kionClient.DetectFinancialMode(ctx); err != nil {
		tflog.Debug(ctx, "could not detect Kion financial mode", map[string]any{"error": err.Error()})
	} else {
		tflog.Info(ctx, "detected Kion financial mode", map[string]any{"budget_mode": kionClient.BudgetMode})
	}

	// Make the client available to resources and data sources
	resp.ResourceData = kionClient
	resp.DataSourceData = kionClient
}

// envAPIPath is the API path override; envAPIPathSet is the companion key
// os.Getenv cannot express. An empty KION_APIPATH is a real choice -- "serve at
// the root" -- and os.Getenv returns "" for that and for unset alike, so
// osGetenv answers envAPIPathSet with "1" when the variable is present at all.
const (
	envAPIPath    = "KION_APIPATH"
	envAPIPathSet = "KION_APIPATH__isset"
)

// osGetenv adapts os.Getenv to the injected getenv seam, answering the
// synthetic envAPIPathSet key from os.LookupEnv.
func osGetenv(key string) string {
	if key == envAPIPathSet {
		if _, ok := os.LookupEnv(envAPIPath); ok {
			return "1"
		}
		return ""
	}
	return os.Getenv(key)
}

// resolvedConfig holds the effective provider configuration after resolving
// attribute values, deprecated aliases, and environment variables.
type resolvedConfig struct {
	apiKey    string
	authToken string
	skipSSL   bool
	serverURL string
}

// resolveProviderConfig resolves the effective provider configuration from the
// schema model, preferring new attribute names over deprecated aliases over
// environment variables, and validates that the required values are present.
// getenv is injected (os.Getenv in production) so the resolution is testable
// without touching the real environment.
func resolveProviderConfig(config kionProviderModel, getenv func(string) string) (resolvedConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Resolve configuration, prefer new names, fall back to old aliases.
	apiURL := config.APIURL.ValueString()
	if apiURL == "" {
		apiURL = config.URL.ValueString()
	}
	if apiURL == "" {
		apiURL = getenv("KION_API_URL")
	}

	apiKey := config.APIKey.ValueString()
	if apiKey == "" {
		apiKey = config.Apikey.ValueString()
	}
	if apiKey == "" {
		apiKey = getenv("KION_API_KEY")
	}

	authToken := config.AuthToken.ValueString()
	if authToken == "" {
		authToken = getenv("KION_AUTH_TOKEN")
	}

	// Validate that API URL is set.
	if apiURL == "" {
		diags.AddError(
			"Missing API URL",
			"The provider cannot be configured without an API URL. "+
				"Set the api_url attribute or KION_API_URL environment variable.",
		)
		return resolvedConfig{}, diags
	}

	// Validate that at least one authentication method is provided.
	if apiKey == "" && authToken == "" {
		diags.AddError(
			"Missing Authentication",
			"The provider requires either an API key or auth token. "+
				"Set the api_key/auth_token attribute or KION_API_KEY/KION_AUTH_TOKEN environment variable.",
		)
		return resolvedConfig{}, diags
	}

	// Resolve skip SSL validation.
	skipSSL := config.SkipSSLValidation.ValueBool()
	if !skipSSL {
		if v := getenv("KION_SKIP_SSL_VALIDATION"); v == "true" || v == "1" {
			skipSSL = true
		}
	}

	// Resolve the API path. The old provider supported `apipath = ""` to disable
	// the default "/api" suffix (used for local development where the API is
	// mounted at the root). Default to "/api" to match the kion-sdk-go behavior.
	//
	// KION_APIPATH exists for the acceptance tests, which configure the provider
	// entirely from the environment and write no provider block -- so without it
	// they cannot reach an install serving at the root, and every test 404s.
	//
	// An empty value is meaningful here rather than absent, which is why this
	// reads the sentinel "" through lookupEnv semantics: KION_APIPATH= (set,
	// empty) selects the root, while leaving it unset keeps "/api".
	apiPath := "/api"
	switch {
	case !config.Apipath.IsNull() && !config.Apipath.IsUnknown():
		apiPath = config.Apipath.ValueString()
	case getenv(envAPIPathSet) != "":
		apiPath = getenv(envAPIPath)
	}

	// Build the full server URL by joining the base URL with the API path.
	serverURL := strings.TrimRight(apiURL, "/") + strings.TrimRight(apiPath, "/")

	return resolvedConfig{
		apiKey:    apiKey,
		authToken: authToken,
		skipSSL:   skipSSL,
		serverURL: serverURL,
	}, diags
}

// securitySource implements generated.SecuritySource. The Kion API uses
// "Bearer <token>" for both API keys and auth tokens.
type securitySource struct {
	apiKey    string
	authToken string
}

func (s *securitySource) Token(_ context.Context, _ generated.OperationName) (generated.Token, error) {
	key := s.authToken
	if key == "" {
		key = s.apiKey
	}
	return generated.Token{APIKey: "Bearer " + key}, nil
}

func (p *kionProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kion"
}

func (p *kionProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	var dataSources []func() datasource.DataSource

	for _, sp := range ServicePackages() {
		for _, ds := range sp.DataSources(ctx) {
			dataSources = append(dataSources, ds.Factory)
		}
	}

	return dataSources
}

func (p *kionProvider) Resources(ctx context.Context) []func() resource.Resource {
	var resources []func() resource.Resource

	for _, sp := range ServicePackages() {
		for _, r := range sp.Resources(ctx) {
			resources = append(resources, r.Factory)
		}
	}

	return resources
}
