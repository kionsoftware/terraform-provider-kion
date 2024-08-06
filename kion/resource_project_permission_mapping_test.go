package kion

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func TestAccResourceProjectPermissionsMapping(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckProjectPermissionsMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceProjectPermissionsMappingConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectPermissionsMappingExists(t, "kion_project_permission_mapping.test"),
					resource.TestCheckResourceAttr("kion_project_permission_mapping.test", "project_id", "11"),
					resource.TestCheckResourceAttr("kion_project_permission_mapping.test", "app_role_id", "4"),
					resource.TestCheckResourceAttr("kion_project_permission_mapping.test", "user_groups_ids.#", "2"),
					resource.TestCheckResourceAttr("kion_project_permission_mapping.test", "user_ids.#", "2"),
				),
			},
			{
				Config: testAccResourceProjectPermissionsMappingConfigUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectPermissionsMappingExists(t, "kion_project_permission_mapping.test"),
					resource.TestCheckResourceAttr("kion_project_permission_mapping.test", "app_role_id", "4"),
				),
			},
			{
				Config: testAccResourceProjectPermissionsMappingConfigEdge(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProjectPermissionsMappingExists(t, "kion_project_permission_mapping.test"),
					resource.TestCheckResourceAttr("kion_project_permission_mapping.test", "user_groups_ids.#", "0"),
					resource.TestCheckResourceAttr("kion_project_permission_mapping.test", "user_ids.#", "0"),
				),
			},
		},
	})
}

func testAccResourceProjectPermissionsMappingConfigUpdated() string {
	return fmt.Sprintf(`
provider "kion" {
  url    = "%s"
  apikey = "%s"
  apipath = ""
}

resource "kion_project_permission_mapping" "test" {
  project_id     = 11
  app_role_id    = 4
  user_groups_ids = [1, 5]
  user_ids       = [1,15]
}
`, os.Getenv("KION_URL"), os.Getenv("KION_APIKEY"))
}

func testAccResourceProjectPermissionsMappingConfigEdge() string {
	return fmt.Sprintf(`
provider "kion" {
  url    = "%s"
  apikey = "%s"
  apipath = ""
}

resource "kion_project_permission_mapping" "test" {
  project_id     = 11
  app_role_id    = 4
  user_groups_ids = []
  user_ids       = []
}
`, os.Getenv("KION_URL"), os.Getenv("KION_APIKEY"))
}

func testAccResourceProjectPermissionsMappingConfig() string {
	return fmt.Sprintf(`
provider "kion" {
  url    = "%s"
  apikey = "%s"
  apipath = ""
}

resource "kion_project_permission_mapping" "test" {
  project_id     = 11
  app_role_id    = 4
  user_groups_ids = [1, 5]
  user_ids       = [1,15]
}
`, os.Getenv("KION_URL"), os.Getenv("KION_APIKEY"))
}

func testAccCheckProjectPermissionsMappingExists(t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		t.Logf("Checking if resource %s exists...", n)
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			t.Errorf("Resource not found: %s", n)
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			t.Errorf("No project permissions mapping ID is set")
			return fmt.Errorf("no project permissions mapping ID is set")
		}

		client := testAccProvider.Meta().(*hc.Client)
		if client == nil {
			t.Fatalf("Kion client is nil. Provider not configured correctly.")
		}
		t.Logf("Client configured: %v", client)

		ids, err := hc.ParseResourceID(rs.Primary.ID, 2, "project_id", "app_role_id")
		if err != nil {
			t.Errorf("Error parsing resource ID: %v", err)
			return fmt.Errorf("error parsing resource ID: %v", err)
		}

		projectID, appRoleID := ids[0], ids[1]
		t.Logf("Fetching project permission mapping for project ID: %d and app role ID: %d", projectID, appRoleID)

		resp := new(hc.ProjectPermissionMappingListResponse)
		err = client.GET(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), resp)
		if err != nil {
			t.Errorf("Error fetching project permission mappings: %v", err)
			return fmt.Errorf("error fetching project permission mappings: %v", err)
		}

		for _, mapping := range resp.Data {
			if mapping.AppRoleID == appRoleID {
				t.Logf("Found project permission mapping with AppRoleID %d in project %d", appRoleID, projectID)
				return nil
			}
		}

		t.Errorf("Project permissions mapping with AppRoleID %d not found in project %d", appRoleID, projectID)
		return fmt.Errorf("project permissions mapping with AppRoleID %d not found in project %d", appRoleID, projectID)
	}
}

func testAccCheckProjectPermissionsMappingDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hc.Client)
	if client == nil {
		return fmt.Errorf("Kion client is nil. Provider not configured correctly during destroy phase.")
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "kion_project_permission_mapping" {
			continue
		}

		ids, err := hc.ParseResourceID(rs.Primary.ID, 2, "project_id", "app_role_id")
		if err != nil {
			return fmt.Errorf("error parsing resource ID: %v", err)
		}

		projectID, appRoleID := ids[0], ids[1]
		resp := new(hc.ProjectPermissionMappingListResponse)
		err = client.GET(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), resp)
		if err != nil {
			return fmt.Errorf("error fetching project permission mappings: %v", err)
		}

		for _, mapping := range resp.Data {
			if mapping.AppRoleID == appRoleID {
				return fmt.Errorf("project permissions mapping with AppRoleID %d still exists in project %d", appRoleID, projectID)
			}
		}
	}

	return nil
}
