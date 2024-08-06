package kion

import (
	"fmt"
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
		return Provider(), nil
	},
}

// TestMain is the entry point for all acceptance tests
// It sets up the environment and runs the tests.
func TestMain(m *testing.M) {
	// Ensure necessary environment variables are set for testing
	if os.Getenv("KION_URL") == "" || os.Getenv("KION_APIKEY") == "" {
		// Skip all tests if necessary environment variables are not set
		fmt.Println("Skipping tests due to missing environment variables")
		os.Exit(0)
	}

	os.Exit(m.Run())
}

// testAccPreCheck verifies that required environment variables are set
// before running acceptance tests.
func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("KION_URL"); v == "" {
		t.Fatal("KION_URL must be set for acceptance tests")
	}
	if v := os.Getenv("KION_APIKEY"); v == "" {
		t.Fatal("KION_APIKEY must be set for acceptance tests")
	}
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
