package conns

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	kion "github.com/kionsoftware/kion-sdk-go"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
)

var (
	sharedClient     *KionClient
	sharedClientOnce sync.Once
	errSharedClient  error
)

// SharedClient returns a KionClient built from environment variables. It is
// safe for concurrent use and caches the client after the first call. Tests
// use this in CheckDestroy / CheckExists helpers and sweepers that need to
// call the SDK directly rather than going through the provider.
func SharedClient() (*KionClient, error) {
	sharedClientOnce.Do(func() {
		sharedClient, errSharedClient = buildSharedClient()
	})
	return sharedClient, errSharedClient
}

// buildSharedClient reads the environment and constructs a KionClient. It holds
// no global state, so tests can exercise it directly with t.Setenv without the
// sync.Once caching in SharedClient getting in the way.
func buildSharedClient() (*KionClient, error) {
	apiURL := os.Getenv("KION_API_URL")
	if apiURL == "" {
		return nil, fmt.Errorf("KION_API_URL must be set")
	}

	apiKey := os.Getenv("KION_API_KEY")
	authToken := os.Getenv("KION_AUTH_TOKEN")
	if apiKey == "" && authToken == "" {
		return nil, fmt.Errorf("KION_API_KEY or KION_AUTH_TOKEN must be set")
	}

	var opts []kion.Option
	if apiKey != "" {
		opts = append(opts, kion.WithAPIKey(apiKey))
	}
	if authToken != "" {
		opts = append(opts, kion.WithBearerToken(authToken))
	}
	skipVerify := false
	if v := os.Getenv("KION_SKIP_SSL_VALIDATION"); v == "true" || v == "1" {
		skipVerify = true
		opts = append(opts, kion.WithSkipVerify(true))
	}
	_ = opts // the SDK client is built directly below; opts is kept for callers that still read it

	// generated.New routes through kion.NormalizeServerURL, which appends "/api"
	// unconditionally -- so it can never address an install that serves its API
	// at the root. That is exactly a local development instance, and it made
	// every acceptance-test check fail with "not found" while the provider,
	// which honors KION_APIPATH, created the record perfectly well: two
	// clients in one test disagreeing about where the API lives.
	//
	// Build the server URL here on the same rule the provider uses and hand it
	// to NewClient, which takes it verbatim.
	serverURL := sharedServerURL(apiURL, os.LookupEnv)

	sdkClient, err := generated.NewClient(serverURL, &sharedSecurity{apiKey: apiKey, authToken: authToken},
		generated.WithClient(kion.BuildHTTPClient(skipVerify, 0)))
	if err != nil {
		return nil, fmt.Errorf("creating shared Kion client: %w", err)
	}

	// APIURL must be the API root, matching what the provider's Configure
	// stores, so the raw helpers and the SDK agree on where the API lives.
	//
	// The credentials are carried separately because the raw helpers build
	// their own requests and read them from these fields, not from the SDK's
	// security source. Leaving them empty made every raw call from a test or
	// sweeper unauthenticated -- a 401 that reads like the record is gone.
	return &KionClient{
		Client:     sdkClient,
		APIURL:     serverURL,
		HTTPClient: kion.BuildHTTPClient(skipVerify, 0),
		APIKey:     apiKey,
		AuthToken:  authToken,
	}, nil
}

// sharedServerURL applies the provider's own apipath rule: default "/api",
// overridden by KION_APIPATH, where set-but-empty means the API is served at
// the root. lookup is injected so this is testable without touching the
// process environment.
func sharedServerURL(apiURL string, lookup func(string) (string, bool)) string {
	apiPath := "/api"
	if v, ok := lookup("KION_APIPATH"); ok {
		apiPath = v
	}
	return strings.TrimRight(apiURL, "/") + strings.TrimRight(apiPath, "/")
}

// sharedSecurity is the acceptance-test client's bearer source. generated.New
// builds one internally, but it is unexported, and NewClient (the constructor
// that does not rewrite the URL) requires the caller to supply it.
type sharedSecurity struct {
	apiKey    string
	authToken string
}

func (s *sharedSecurity) Token(_ context.Context, _ generated.OperationName) (generated.Token, error) {
	key := s.authToken
	if key == "" {
		key = s.apiKey
	}
	return generated.Token{APIKey: "Bearer " + key}, nil
}
