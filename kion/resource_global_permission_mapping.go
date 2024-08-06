package kion

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// resourceGlobalPermissionMapping returns a schema.Resource for managing global permission mappings in Kion.
func resourceGlobalPermissionMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGlobalPermissionMappingCreate,
		ReadContext:   resourceGlobalPermissionMappingRead,
		UpdateContext: resourceGlobalPermissionMappingUpdate,
		DeleteContext: resourceGlobalPermissionMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceGlobalPermissionMappingImport,
		},

		Schema: map[string]*schema.Schema{
			"app_role_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Application role ID for the permission mapping.",
			},
			"user_groups_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Set of user group IDs for the permission mapping.",
			},
			"user_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Set of user IDs for the permission mapping.",
			},
		},
	}
}

// resourceGlobalPermissionMappingCreate handles the creation of the resource
func resourceGlobalPermissionMappingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	appRoleID := d.Get("app_role_id").(int) // Retrieve app_role_id from the resource data

	// Create a GlobalPermissionMapping object using the provided data
	mapping := hc.GlobalPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List()),
		UserIDs:       hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List()),
	}

	// Make a POST request to the Kion API to create the permission mapping
	_, err := client.POST("/v3/global/permission-mapping", []hc.GlobalPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err) // Return an error diagnostic if the request fails
	}

	// Set the resource ID using the appRoleID
	d.SetId(fmt.Sprintf("%d", appRoleID))

	// Read and update the state of the newly created resource
	return resourceGlobalPermissionMappingRead(ctx, d, m)
}

// resourceGlobalPermissionMappingRead retrieves the state of the resource from the Kion API
func resourceGlobalPermissionMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Convert the resource ID to an integer
	appRoleID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Make a GET request to the Kion API to fetch all global permission mappings
	resp := new(hc.GlobalPermissionMappingListResponse)
	err = client.GET("/v3/global/permission-mapping", resp)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics
	found := false

	// Iterate through the retrieved mappings to find the one matching the appRoleID
	for _, mapping := range resp.Data {
		if mapping.AppRoleID == appRoleID {
			// Sort the user group IDs and user IDs for consistent ordering
			sort.Ints(mapping.UserGroupsIDs)
			sort.Ints(mapping.UserIDs)

			diags = append(diags, hc.SafeSet(d, "app_role_id", mapping.AppRoleID)...)
			diags = append(diags, hc.SafeSet(d, "user_groups_ids", mapping.UserGroupsIDs)...)
			diags = append(diags, hc.SafeSet(d, "user_ids", mapping.UserIDs)...)
			found = true
			break
		}
	}

	// If the mapping is not found, it implies the resource has been deleted externally
	if !found {
		d.SetId("") // Remove the ID to indicate the resource no longer exists
	}

	return diags
}

// resourceGlobalPermissionMappingUpdate handles updating the resource
func resourceGlobalPermissionMappingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	appRoleID := d.Get("app_role_id").(int) // Retrieve app_role_id from the resource data

	// Create an updated GlobalPermissionMapping object using the provided data
	updatedMapping := hc.GlobalPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List()),
		UserIDs:       hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List()),
	}

	// Fetch existing mappings from the API
	resp := new(hc.GlobalPermissionMappingListResponse)
	err := client.GET("/v3/global/permission-mapping", resp)
	if err != nil {
		return diag.FromErr(err)
	}

	existingMappings := resp.Data
	updatedMappings := make([]hc.GlobalPermissionMapping, 0)
	found := false

	// Iterate through existing mappings to update the matching one
	for _, existing := range existingMappings {
		if existing.AppRoleID == appRoleID {
			// Replace the existing mapping with the updated one
			updatedMappings = append(updatedMappings, updatedMapping)
			found = true
		} else {
			updatedMappings = append(updatedMappings, existing)
		}
	}

	// If the mapping wasn't found, add it as a new mapping
	if !found {
		updatedMappings = append(updatedMappings, updatedMapping)
	}

	// Send the updated mappings to the Kion API
	err = client.PATCH("/v3/global/permission-mapping", updatedMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	// Read and update the state of the updated resource
	return resourceGlobalPermissionMappingRead(ctx, d, m)
}

// resourceGlobalPermissionMappingDelete handles the deletion of the resource
func resourceGlobalPermissionMappingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Convert the resource ID to an integer
	appRoleID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Create a mapping with empty user IDs and user group IDs to effectively delete it
	mapping := hc.GlobalPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: []int{},
		UserIDs:       []int{},
	}

	// Send the delete request to the Kion API
	err = client.PATCH("/v3/global/permission-mapping", []hc.GlobalPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	// Remove the resource ID to indicate it has been deleted
	d.SetId("")

	return nil
}

// resourceGlobalPermissionMappingImport handles the import of existing resources into Terraform
func resourceGlobalPermissionMappingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Convert the resource ID to an integer
	appRoleID, err := strconv.Atoi(d.Id())
	if err != nil {
		return nil, fmt.Errorf("invalid app role ID, must be an integer")
	}

	// Set the app_role_id field in the resource data
	if err := d.Set("app_role_id", appRoleID); err != nil {
		return nil, err
	}

	// Return the resource data for importing
	return []*schema.ResourceData{d}, nil
}
