package kion

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testAccProvider and testAccProviders are used to manage provider instances for testing
var testAccProvider *schema.Provider
var testAccProviders map[string]*schema.Provider

// Initialize provider instances and factories for testing
func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"kion": testAccProvider,
	}
}

// testAccProviderFactories sets up provider instances for testing
var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"kion": func() (*schema.Provider, error) {
		p := Provider()
		err := p.InternalValidate() // Ensure provider configuration is validated
		if err != nil {
			return nil, err
		}
		return p, nil
	},
}

// testAccPreCheck verifies that required environment variables are set
// before running acceptance tests.
func testAccPreCheck(t *testing.T) {
	url := os.Getenv("KION_URL")
	apiKey := os.Getenv("KION_APIKEY")

	if url == "" {
		t.Fatal("KION_URL must be set for acceptance tests")
	}
	if apiKey == "" {
		t.Fatal("KION_APIKEY must be set for acceptance tests")
	}

	t.Logf("Using KION_URL: %s", url)
	t.Logf("Using KION_APIKEY: %s", apiKey)
}

// TestProvider checks if the provider can be validated correctly
func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// TestProvider_impl ensures the provider implements the schema.Provider interface
func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}
