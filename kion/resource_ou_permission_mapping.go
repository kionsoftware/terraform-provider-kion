package kion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// resourceOUPermissionMapping returns a schema.Resource for managing OU permission mappings in Kion.
func resourceOUPermissionsMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOUPermissionMappingCreate,
		ReadContext:   resourceOUPermissionMappingRead,
		UpdateContext: resourceOUPermissionMappingUpdate,
		DeleteContext: resourceOUPermissionMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOUPermissionMappingImport,
		},
		Schema: map[string]*schema.Schema{
			"ou_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the OU to manage permission mappings for.",
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

// resourceOUPermissionMappingCreate handles the creation of the resource
func resourceOUPermissionMappingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	ouID := d.Get("ou_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	// Create an OUPermissionMapping object using the provided data
	mapping := hc.OUPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List()),
		UserIDs:       hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List()),
	}

	// Make a PATCH request to the Kion API to create the permission mapping
	err := client.PATCH(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), []hc.OUPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the resource ID using the ouID and appRoleID
	d.SetId(fmt.Sprintf("%d-%d", ouID, appRoleID))

	// Ensure the state reflects the provided list
	return resourceOUPermissionMappingRead(ctx, d, m)
}

// resourceOUPermissionMappingRead retrieves the state of the resource from the Kion API
func resourceOUPermissionMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Use the generic ParseResourceID function to extract ouID and appRoleID
	ids, err := hc.ParseResourceID(d.Id(), 2, "ou_id", "app_role_id")
	if err != nil {
		return diag.FromErr(err)
	}

	ouID, appRoleID := ids[0], ids[1]

	resp := new(hc.OUPermissionMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	found := false
	for _, mapping := range resp.Data {
		if mapping.AppRoleID == appRoleID {
			diags = append(diags, hc.SafeSet(d, "ou_id", ouID)...)
			diags = append(diags, hc.SafeSet(d, "app_role_id", appRoleID)...)
			diags = append(diags, hc.SafeSet(d, "user_groups_ids", mapping.UserGroupsIDs)...)
			diags = append(diags, hc.SafeSet(d, "user_ids", mapping.UserIDs)...)
			found = true
			break
		}
	}

	if !found {
		d.SetId("")
	}

	return diags
}

// resourceOUPermissionMappingUpdate handles updating the resource
func resourceOUPermissionMappingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	ouID := d.Get("ou_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	// Create an updated OUPermissionMapping object using the provided data
	updatedMapping := hc.OUPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List()),
		UserIDs:       hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List()),
	}

	// Fetch existing mappings from the API
	resp := new(hc.OUPermissionMappingListResponse)
	err := client.GET(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	existingMappings := resp.Data
	updatedMappings := make([]hc.OUPermissionMapping, 0)
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
	err = client.PATCH(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), updatedMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ensure the state reflects the provided list
	return resourceOUPermissionMappingRead(ctx, d, m)
}

// resourceOUPermissionMappingDelete handles the deletion of the resource
func resourceOUPermissionMappingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Use the generic ParseResourceID function to extract ouID and appRoleID
	ids, err := hc.ParseResourceID(d.Id(), 2, "ou_id", "app_role_id")
	if err != nil {
		return diag.FromErr(err)
	}

	ouID, appRoleID := ids[0], ids[1]

	// Create a mapping with empty user IDs and user group IDs to effectively delete it
	mapping := hc.OUPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: []int{},
		UserIDs:       []int{},
	}

	// Send the delete request to the Kion API
	err = client.PATCH(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), []hc.OUPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	// Remove the resource ID to indicate it has been deleted
	d.SetId("")

	return nil
}

// resourceOUPermissionMappingImport handles the import of existing resources into Terraform
func resourceOUPermissionMappingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Use the generic ParseResourceID function to extract ouID and appRoleID
	ids, err := hc.ParseResourceID(d.Id(), 2, "ou_id", "app_role_id")
	if err != nil {
		return nil, err
	}

	ouID, appRoleID := ids[0], ids[1]

	// Set the ou_id and app_role_id fields in the resource data
	if err := d.Set("ou_id", ouID); err != nil {
		return nil, err
	}
	if err := d.Set("app_role_id", appRoleID); err != nil {
		return nil, err
	}

	// Return the resource data for importing
	return []*schema.ResourceData{d}, nil
}
