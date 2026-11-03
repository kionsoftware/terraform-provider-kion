// Package acctest provides helpers for acceptance testing.
package acctest

import (
	"context"
	"fmt"
	"os"
	"testing"

	"terraform-provider-kion/internal/conns"
	"terraform-provider-kion/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

// ResourcePrefix is the standard prefix for resources created in acceptance tests.
const ResourcePrefix = "test-acc"

// RandomWithPrefix returns a unique name with the given prefix, suitable for
// resource names in acceptance tests. The format is "<prefix>-<random>".
func RandomWithPrefix(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlphaNum))
}

// Context returns a context.Context for use in acceptance tests.
func Context(_ *testing.T) context.Context {
	return context.Background()
}

// ProtoV6ProviderFactories returns provider factories for acceptance testing.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"kion": providerserver.NewProtocol6WithError(provider.New()()),
}

// SharedClient returns a KionClient built from environment variables.
// Delegates to conns.SharedClient().
func SharedClient() (*conns.KionClient, error) {
	return conns.SharedClient()
}

// PreCheck validates that required environment variables are set for
// acceptance tests, failing the test if not.
func PreCheck(t *testing.T) {
	t.Helper()

	if err := checkPreCheck(os.Getenv); err != nil {
		t.Fatal(err)
	}
}

// checkPreCheck is the pure, testable core of PreCheck: it reports the first
// missing-precondition error (if any) given an environment lookup function.
func checkPreCheck(getenv func(string) string) error {
	if getenv("KION_API_URL") == "" {
		return fmt.Errorf("KION_API_URL must be set for acceptance tests")
	}
	if getenv("KION_API_KEY") == "" && getenv("KION_AUTH_TOKEN") == "" {
		return fmt.Errorf("KION_API_KEY or KION_AUTH_TOKEN must be set for acceptance tests")
	}
	return nil
}
