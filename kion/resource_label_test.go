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
	resourceTypeLabel = "kion_label"
	resourceNameLabel = "label1"
)

var (
	initialTestLabel = hc.LabelCreate{
		Color: "#123abc",
		Key:   "environment",
		Value: "production",
	}
	updatedTestLabel = hc.LabelCreate{
		Color: "#abcdef",
		Key:   "env",
		Value: "prod",
	}
)

func TestAccResourceLabelCreate(t *testing.T) {
	config := testAccLabelGenerateResourceDeclaration(&initialTestLabel)
	printHCLConfig(config) // Print the generated HCL configuration

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccLabelCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(testAccLabelCheckResource(&initialTestLabel)...),
			},
		},
	})
}

func TestAccResourceLabelUpdate(t *testing.T) {
	initialConfig := testAccLabelGenerateResourceDeclaration(&initialTestLabel)
	updatedConfig := testAccLabelGenerateResourceDeclaration(&updatedTestLabel)
	printHCLConfig(initialConfig) // Print the initial HCL configuration
	printHCLConfig(updatedConfig) // Print the updated HCL configuration

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccLabelCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check:  resource.ComposeTestCheckFunc(testAccLabelCheckResource(&initialTestLabel)...),
			},
			{
				Config: updatedConfig,
				Check:  resource.ComposeTestCheckFunc(testAccLabelCheckResource(&updatedTestLabel)...),
			},
		},
	})
}

func TestAccResourceLabelDelete(t *testing.T) {
	config := testAccLabelGenerateResourceDeclaration(&initialTestLabel)
	printHCLConfig(config) // Print the generated HCL configuration

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccLabelCheckResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.ComposeTestCheckFunc(testAccLabelCheckResource(&initialTestLabel)...),
			},
			{
				Config:       "", // Apply an empty configuration to delete the resource
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						// Ensure the resource is removed from the state
						for _, rs := range s.RootModule().Resources {
							if rs.Type == resourceTypeLabel {
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

func printHCLConfig(config string) {
	fmt.Println("Generated HCL configuration:")
	fmt.Println(config)
}

// testAccLabelCheckResource returns a slice of functions that validate the test resource's fields
func testAccLabelCheckResource(label *hc.LabelCreate) (funcs []resource.TestCheckFunc) {
	funcs = []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceTypeLabel+"."+resourceNameLabel, "color", label.Color),
		resource.TestCheckResourceAttr(resourceTypeLabel+"."+resourceNameLabel, "key", label.Key),
		resource.TestCheckResourceAttr(resourceTypeLabel+"."+resourceNameLabel, "value", label.Value),
	}
	return
}

// testAccLabelGenerateResourceDeclaration generates a resource declaration string (a la main.tf)
func testAccLabelGenerateResourceDeclaration(label *hc.LabelCreate) string {
	if label == nil {
		return ""
	}

	return fmt.Sprintf(`
		resource "%v" "%v" {
			color = "%v"
			key   = "%v"
			value = "%v"
		}`,
		resourceTypeLabel, resourceNameLabel,
		label.Color,
		label.Key,
		label.Value,
	)
}

// testAccLabelCheckResourceDestroy verifies the resource has been destroyed
func testAccLabelCheckResourceDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	if meta == nil {
		return nil
	}

	client := meta.(*hc.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceTypeLabel {
			continue
		}

		resp := new(hc.LabelResponse)
		err := client.GET(fmt.Sprintf("/v3/label/%s", rs.Primary.ID), resp)
		if err == nil {
			if fmt.Sprint(resp.Data.ID) == rs.Primary.ID {
				return fmt.Errorf("Label (%s) still exists.", rs.Primary.ID)
			}
			return nil
		}

		if !strings.Contains(err.Error(), fmt.Sprintf("status: %d", http.StatusNotFound)) {
			return err
		}
	}

	return nil
}
