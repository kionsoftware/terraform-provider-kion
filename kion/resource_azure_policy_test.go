package kion

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

const (
	resourceTypeAzurePolicy = "kion_azure_policy"
	resourceNameAzurePolicy = "azure_policy1"
)

var (
	initialTestAzurePolicy = hc.AzurePolicyCreate{
		AzurePolicy: struct {
			Description string `json:"description"`
			Name        string `json:"name"`
			Parameters  string `json:"parameters"`
			Policy      string `json:"policy"`
		}{
			Description: "Initial description",
			Name:        "Initial Policy",
			Parameters:  "{}",
			Policy:      "{}",
		},
		OwnerUserGroups: &[]int{1},
		OwnerUsers:      &[]int{1},
	}
	updatedTestAzurePolicy = hc.AzurePolicyCreate{
		AzurePolicy: struct {
			Description string `json:"description"`
			Name        string `json:"name"`
			Parameters  string `json:"parameters"`
			Policy      string `json:"policy"`
		}{
			Description: "Updated description",
			Name:        "Updated Policy",
			Parameters:  "{\"param\":\"value\"}",
			Policy:      "{\"rule\":\"new_rule\"}",
		},
		OwnerUserGroups: &[]int{2},
		OwnerUsers:      &[]int{2},
	}
)

func TestAccResourceAzurePolicyCreate(t *testing.T) {
	config := testAccAzurePolicyGenerateResourceDeclaration(&initialTestAzurePolicy)
	hc.PrintHCLConfig(config) // Print the generated HCL configuration

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAzurePolicyCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(testAccAzurePolicyCheckResource(&initialTestAzurePolicy)...),
			},
		},
	})
}

func TestAccResourceAzurePolicyUpdate(t *testing.T) {
	initialConfig := testAccAzurePolicyGenerateResourceDeclaration(&initialTestAzurePolicy)
	updatedConfig := testAccAzurePolicyGenerateResourceDeclaration(&updatedTestAzurePolicy)
	hc.PrintHCLConfig(initialConfig) // Print the initial HCL configuration
	hc.PrintHCLConfig(updatedConfig) // Print the updated HCL configuration

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAzurePolicyCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check:  resource.ComposeTestCheckFunc(testAccAzurePolicyCheckResource(&initialTestAzurePolicy)...),
			},
			{
				Config: updatedConfig,
				Check:  resource.ComposeTestCheckFunc(testAccAzurePolicyCheckResource(&updatedTestAzurePolicy)...),
			},
		},
	})
}

func TestAccResourceAzurePolicyDelete(t *testing.T) {
	config := testAccAzurePolicyGenerateResourceDeclaration(&initialTestAzurePolicy)
	hc.PrintHCLConfig(config) // Print the generated HCL configuration

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAzurePolicyCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(testAccAzurePolicyCheckResource(&initialTestAzurePolicy)...),
			},
			{
				Config:       "", // Apply an empty configuration to delete the resource
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// Ensure the resource is removed from the state
						for _, rs := range s.RootModule().Resources {
							if rs.Type == resourceTypeAzurePolicy {
								return fmt.Errorf("expected no resources, found %s", rs.Primary.ID)
							}
						}
						return nil
					},
				),
			},
		},
	})
}

// testAccAzurePolicyCheckResource returns a slice of functions that validate the test resource's fields
func testAccAzurePolicyCheckResource(policy *hc.AzurePolicyCreate) (funcs []resource.TestCheckFunc) {
	funcs = []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceTypeAzurePolicy+"."+resourceNameAzurePolicy, "description", policy.AzurePolicy.Description),
		resource.TestCheckResourceAttr(resourceTypeAzurePolicy+"."+resourceNameAzurePolicy, "name", policy.AzurePolicy.Name),
		resource.TestCheckResourceAttr(resourceTypeAzurePolicy+"."+resourceNameAzurePolicy, "parameters", policy.AzurePolicy.Parameters),
		resource.TestCheckResourceAttr(resourceTypeAzurePolicy+"."+resourceNameAzurePolicy, "policy", policy.AzurePolicy.Policy),
	}
	return
}

// testAccAzurePolicyGenerateResourceDeclaration generates a resource declaration string (a la main.tf)
func testAccAzurePolicyGenerateResourceDeclaration(policy *hc.AzurePolicyCreate) string {
	if policy == nil {
		return ""
	}

	return fmt.Sprintf(`
		resource "%v" "%v" {
			description = "%v"
			name        = "%v"
			parameters  = "%v"
			policy      = "%v"
			owner_user_groups = [%v]
			owner_users = [%v]
		}`,
		resourceTypeAzurePolicy, resourceNameAzurePolicy,
		policy.AzurePolicy.Description,
		policy.AzurePolicy.Name,
		policy.AzurePolicy.Parameters,
		policy.AzurePolicy.Policy,
		strings.Join(intSliceToStringSlice(*policy.OwnerUserGroups), ","),
		strings.Join(intSliceToStringSlice(*policy.OwnerUsers), ","),
	)
}

// testAccAzurePolicyCheckResourceDestroy verifies the resource has been destroyed
func testAccAzurePolicyCheckResourceDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	if meta == nil {
		return nil
	}

	client := meta.(*hc.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceTypeAzurePolicy {
			continue
		}

		resp := new(hc.AzurePolicyResponse)
		err := client.GET(fmt.Sprintf("/v3/azure-policy/%s", rs.Primary.ID), resp)
		if err == nil {
			if fmt.Sprint(resp.Data.AzurePolicy.ID) == rs.Primary.ID {
				return fmt.Errorf("Azure Policy (%s) still exists.", rs.Primary.ID)
			}
			return nil
		}

		if !strings.Contains(err.Error(), fmt.Sprintf("status: %d", http.StatusNotFound)) {
			return err
		}
	}

	return nil
}

func intSliceToStringSlice(intSlice []int) []string {
	strSlice := make([]string, len(intSlice))
	for i, val := range intSlice {
		strSlice[i] = fmt.Sprintf("%d", val)
	}
	return strSlice
}
