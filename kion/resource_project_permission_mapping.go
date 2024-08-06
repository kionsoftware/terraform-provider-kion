package kion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// resourceProjectPermissionsMapping returns a schema.Resource for managing project permission mappings in Kion.
func resourceProjectPermissionsMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectPermissionsMappingCreate,
		ReadContext:   resourceProjectPermissionsMappingRead,
		UpdateContext: resourceProjectPermissionsMappingUpdate,
		DeleteContext: resourceProjectPermissionsMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectPermissionsMappingImport,
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the project to manage permission mappings for.",
			},
			"app_role_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Application role ID for the permission mapping.",
			},
			"user_groups_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Set of user group IDs for the permission mapping (must be provided in numerical order).",
			},
			"user_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Set of user IDs for the permission mapping (must be provided in numerical order).",
			},
		},
	}
}

// resourceProjectPermissionsMappingCreate handles the creation of the resource
func resourceProjectPermissionsMappingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	projectID := d.Get("project_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	// Create a ProjectPermissionsMapping object using the provided data
	mapping := hc.ProjectPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List()),
		UserIDs:       hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List()),
	}

	// Log the mapping details
	fmt.Printf("Creating project permission mapping: project_id=%d, app_role_id=%d, user_groups_ids=%v, user_ids=%v\n", projectID, appRoleID, mapping.UserGroupsIDs, mapping.UserIDs)

	// Make a PATCH request to the Kion API to create the permission mapping
	err := client.PATCH(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), []hc.ProjectPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the resource ID using the projectID and appRoleID
	d.SetId(fmt.Sprintf("%d-%d", projectID, appRoleID))

	// Ensure the state reflects the provided list
	return resourceProjectPermissionsMappingRead(ctx, d, m)
}

// resourceProjectPermissionsMappingRead retrieves the state of the resource from the Kion API
func resourceProjectPermissionsMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Use the generic ParseResourceID function to extract projectID and appRoleID
	ids, err := hc.ParseResourceID(d.Id(), 2, "project_id", "app_role_id")
	if err != nil {
		return diag.FromErr(err)
	}

	projectID, appRoleID := ids[0], ids[1]

	fmt.Printf("Reading project permission mapping: project_id=%d, app_role_id=%d\n", projectID, appRoleID)

	resp := new(hc.ProjectPermissionMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	found := false
	for _, mapping := range resp.Data {
		if mapping.AppRoleID == appRoleID {
			fmt.Printf("Found mapping: project_id=%d, app_role_id=%d, user_groups_ids=%v, user_ids=%v\n", projectID, appRoleID, mapping.UserGroupsIDs, mapping.UserIDs)
			diags = append(diags, hc.SafeSet(d, "project_id", projectID)...)
			diags = append(diags, hc.SafeSet(d, "app_role_id", appRoleID)...)
			diags = append(diags, hc.SafeSet(d, "user_groups_ids", mapping.UserGroupsIDs)...)
			diags = append(diags, hc.SafeSet(d, "user_ids", mapping.UserIDs)...)
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Mapping not found for app_role_id=%d in project_id=%d, setting ID to empty\n", appRoleID, projectID)
		d.SetId("")
	}

	return diags
}

// resourceProjectPermissionsMappingUpdate handles updating the resource
func resourceProjectPermissionsMappingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	projectID := d.Get("project_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	fmt.Printf("Updating project permission mapping: project_id=%d, app_role_id=%d\n", projectID, appRoleID)

	// Check if the app_role_id has changed
	if d.HasChange("app_role_id") {
		// Fetch the old app_role_id
		oldAppRoleID, _ := d.GetChange("app_role_id")

		// Remove the old mapping
		oldMapping := hc.ProjectPermissionMapping{
			AppRoleID:     oldAppRoleID.(int),
			UserGroupsIDs: []int{},
			UserIDs:       []int{},
		}

		fmt.Printf("Removing old mapping for app_role_id=%d\n", oldAppRoleID)

		err := client.PATCH(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), []hc.ProjectPermissionMapping{oldMapping})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Create an updated ProjectPermissionsMapping object using the provided data
	updatedMapping := hc.ProjectPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List()),
		UserIDs:       hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List()),
	}

	// Fetch existing mappings from the API
	resp := new(hc.ProjectPermissionMappingListResponse)
	err := client.GET(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	existingMappings := resp.Data
	updatedMappings := make([]hc.ProjectPermissionMapping, 0)
	found := false

	// Iterate through existing mappings to update the matching one
	for _, existing := range existingMappings {
		if existing.AppRoleID == appRoleID {
			fmt.Printf("Updating existing mapping: project_id=%d, app_role_id=%d\n", projectID, appRoleID)
			// Replace the existing mapping with the updated one
			updatedMappings = append(updatedMappings, updatedMapping)
			found = true
		} else {
			updatedMappings = append(updatedMappings, existing)
		}
	}

	// If the mapping wasn't found, add it as a new mapping
	if !found {
		fmt.Printf("Adding new mapping: project_id=%d, app_role_id=%d\n", projectID, appRoleID)
		updatedMappings = append(updatedMappings, updatedMapping)
	}

	// Send the updated mappings to the Kion API
	err = client.PATCH(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), updatedMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ensure the state reflects the provided list
	return resourceProjectPermissionsMappingRead(ctx, d, m)
}

// resourceProjectPermissionsMappingDelete handles the deletion of the resource
func resourceProjectPermissionsMappingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Use the generic ParseResourceID function to extract projectID and appRoleID
	ids, err := hc.ParseResourceID(d.Id(), 2, "project_id", "app_role_id")
	if err != nil {
		return diag.FromErr(err)
	}

	projectID, appRoleID := ids[0], ids[1]

	fmt.Printf("Deleting project permission mapping: project_id=%d, app_role_id=%d\n", projectID, appRoleID)

	// Create a mapping with empty user IDs and user group IDs to effectively delete it
	mapping := hc.ProjectPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: []int{},
		UserIDs:       []int{},
	}

	// Send the delete request to the Kion API
	err = client.PATCH(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), []hc.ProjectPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	// Remove the resource ID to indicate it has been deleted
	d.SetId("")
	fmt.Printf("Successfully deleted project permission mapping: project_id=%d, app_role_id=%d\n", projectID, appRoleID)

	return nil
}

// resourceProjectPermissionsMappingImport handles the import of existing resources into Terraform
func resourceProjectPermissionsMappingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Use the generic ParseResourceID function to extract projectID and appRoleID
	ids, err := hc.ParseResourceID(d.Id(), 2, "project_id", "app_role_id")
	if err != nil {
		return nil, err
	}

	projectID, appRoleID := ids[0], ids[1]

	fmt.Printf("Importing project permission mapping: project_id=%d, app_role_id=%d\n", projectID, appRoleID)

	// Set the project_id and app_role_id fields in the resource data
	if err := d.Set("project_id", projectID); err != nil {
		return nil, err
	}
	if err := d.Set("app_role_id", appRoleID); err != nil {
		return nil, err
	}

	// Return the resource data for importing
	return []*schema.ResourceData{d}, nil
}
