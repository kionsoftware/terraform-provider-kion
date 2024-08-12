package kion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// resourceFundingSourcePermissionsMapping returns a schema.Resource for managing funding source permission mappings in Kion.
func resourceFundingSourcePermissionsMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFundingSourcePermissionsMappingCreate,
		ReadContext:   resourceFundingSourcePermissionsMappingRead,
		UpdateContext: resourceFundingSourcePermissionsMappingUpdate,
		DeleteContext: resourceFundingSourcePermissionsMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceFundingSourcePermissionsMappingImport,
		},
		Schema: map[string]*schema.Schema{
			"funding_source_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the Funding Source to manage permission mappings for.",
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

// resourceFundingSourcePermissionsMappingCreate handles the creation of the resource
func resourceFundingSourcePermissionsMappingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	fundingSourceID := d.Get("funding_source_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	// Convert user_groups_ids and user_ids from interface{} to int slices
	userGroupsIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List())
	if err != nil {
		return diag.Errorf("failed to convert user_groups_ids: %v", err)
	}

	userIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List())
	if err != nil {
		return diag.Errorf("failed to convert user_ids: %v", err)
	}

	// Create a FundingSourcePermissionsMapping object using the provided data
	mapping := hc.FundingSourcePermissionsMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: userGroupsIDs,
		UserIDs:       userIDs,
	}

	// Make a PATCH request to the Kion API to create the permission mapping
	err = client.PATCH(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), []hc.FundingSourcePermissionsMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the resource ID using the fundingSourceID and appRoleID
	d.SetId(fmt.Sprintf("%d-%d", fundingSourceID, appRoleID))

	// Ensure the state reflects the provided list
	return resourceFundingSourcePermissionsMappingRead(ctx, d, m)
}

// resourceFundingSourcePermissionsMappingRead retrieves the state of the resource from the Kion API
func resourceFundingSourcePermissionsMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Use the generic ParseResourceID function to extract fundingSourceID and appRoleID
	ids, err := hc.ParseResourceID(d.Id(), 2, "funding_source_id", "app_role_id")
	if err != nil {
		return diag.FromErr(err)
	}

	fundingSourceID, appRoleID := ids[0], ids[1]

	resp := new(hc.FundingSourcePermissionsMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	found := false
	for _, mapping := range resp.Data {
		if mapping.AppRoleID == appRoleID {
			diags = append(diags, hc.SafeSet(d, "funding_source_id", fundingSourceID, "Failed to set funding_source_id")...)
			diags = append(diags, hc.SafeSet(d, "app_role_id", appRoleID, "Failed to se app_role_id")...)
			diags = append(diags, hc.SafeSet(d, "user_groups_ids", mapping.UserGroupsIDs, "Failed to set user_groups_ids")...)
			diags = append(diags, hc.SafeSet(d, "user_ids", mapping.UserIDs, "Failed to set user_ids")...)
			found = true
			break
		}
	}

	if !found {
		d.SetId("")
	}

	return diags
}

// resourceFundingSourcePermissionsMappingUpdate handles updating the resource
func resourceFundingSourcePermissionsMappingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	fundingSourceID := d.Get("funding_source_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	// Check if the app_role_id has changed
	if d.HasChange("app_role_id") {
		// Fetch the old app_role_id
		oldAppRoleID, _ := d.GetChange("app_role_id")

		// Remove the old mapping
		oldMapping := hc.FundingSourcePermissionsMapping{
			AppRoleID:     oldAppRoleID.(int),
			UserGroupsIDs: []int{},
			UserIDs:       []int{},
		}

		err := client.PATCH(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), []hc.FundingSourcePermissionsMapping{oldMapping})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Convert user_groups_ids and user_ids from interface{} to int slices
	userGroupsIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List())
	if err != nil {
		return diag.Errorf("failed to convert user_groups_ids: %v", err)
	}

	userIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List())
	if err != nil {
		return diag.Errorf("failed to convert user_ids: %v", err)
	}

	// Create an updated FundingSourcePermissionsMapping object using the provided data
	updatedMapping := hc.FundingSourcePermissionsMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: userGroupsIDs,
		UserIDs:       userIDs,
	}

	// Fetch existing mappings from the API
	resp := new(hc.FundingSourcePermissionsMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	existingMappings := resp.Data
	updatedMappings := make([]hc.FundingSourcePermissionsMapping, 0)
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
	err = client.PATCH(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), updatedMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ensure the state reflects the provided list
	return resourceFundingSourcePermissionsMappingRead(ctx, d, m)
}

// resourceFundingSourcePermissionsMappingDelete handles the deletion of the resource
func resourceFundingSourcePermissionsMappingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Use the generic ParseResourceID function to extract fundingSourceID and appRoleID
	ids, err := hc.ParseResourceID(d.Id(), 2, "funding_source_id", "app_role_id")
	if err != nil {
		return diag.FromErr(err)
	}

	fundingSourceID, appRoleID := ids[0], ids[1]

	// Create a mapping with empty user IDs and user group IDs to effectively delete it
	mapping := hc.FundingSourcePermissionsMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: []int{},
		UserIDs:       []int{},
	}

	// Send the delete request to the Kion API
	err = client.PATCH(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), []hc.FundingSourcePermissionsMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	// Remove the resource ID to indicate it has been deleted
	d.SetId("")

	return nil
}

// resourceFundingSourcePermissionsMappingImport handles the import of existing resources into Terraform
func resourceFundingSourcePermissionsMappingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Use the generic ParseResourceID function to extract fundingSourceID and appRoleID
	ids, err := hc.ParseResourceID(d.Id(), 2, "funding_source_id", "app_role_id")
	if err != nil {
		return nil, err
	}

	fundingSourceID, appRoleID := ids[0], ids[1]

	// Set the funding_source_id and app_role_id fields in the resource data
	if err := d.Set("funding_source_id", fundingSourceID); err != nil {
		return nil, err
	}
	if err := d.Set("app_role_id", appRoleID); err != nil {
		return nil, err
	}

	// Return the resource data for importing
	return []*schema.ResourceData{d}, nil
}
