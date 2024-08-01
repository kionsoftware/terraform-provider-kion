package kion

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

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

func resourceFundingSourcePermissionsMappingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	fundingSourceID := d.Get("funding_source_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	userGroupsIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List())
	userIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List())

	mapping := hc.FundingSourcePermissionsMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: userGroupsIDs,
		UserIDs:       userIDs,
	}

	err := client.PATCH(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), []hc.FundingSourcePermissionsMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d-%d", fundingSourceID, appRoleID))

	// Ensure the state reflects the provided list
	d.Set("user_groups_ids", userGroupsIDs)
	d.Set("user_ids", userIDs)

	return resourceFundingSourcePermissionsMappingRead(ctx, d, m)
}

func resourceFundingSourcePermissionsMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	parts := strings.Split(d.Id(), "-")
	if len(parts) != 2 {
		return diag.Errorf("invalid resource ID format, expected {funding_source_id}-{app_role_id}")
	}

	fundingSourceID, err := strconv.Atoi(parts[0])
	if err != nil {
		return diag.FromErr(err)
	}
	appRoleID, err := strconv.Atoi(parts[1])
	if err != nil {
		return diag.FromErr(err)
	}

	resp := new(hc.FundingSourcePermissionsMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	found := false
	for _, mapping := range resp.Data {
		if mapping.AppRoleID == appRoleID {
			// Set lists to the state as provided
			diags = append(diags, hc.SafeSet(d, "funding_source_id", fundingSourceID)...)
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

func resourceFundingSourcePermissionsMappingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	fundingSourceID := d.Get("funding_source_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	userGroupsIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List())
	userIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List())

	mapping := hc.FundingSourcePermissionsMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: userGroupsIDs,
		UserIDs:       userIDs,
	}

	resp := new(hc.FundingSourcePermissionsMappingListResponse)
	err := client.GET(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	updatedMappings := make([]hc.FundingSourcePermissionsMapping, 0)
	found := false
	for _, existing := range resp.Data {
		if existing.AppRoleID == appRoleID {
			existing.UserGroupsIDs = userGroupsIDs
			existing.UserIDs = userIDs
			updatedMappings = append(updatedMappings, existing)
			found = true
		} else {
			updatedMappings = append(updatedMappings, existing)
		}
	}

	if !found {
		updatedMappings = append(updatedMappings, mapping)
	}

	err = client.PATCH(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), updatedMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ensure the state reflects the provided list
	d.Set("user_groups_ids", userGroupsIDs)
	d.Set("user_ids", userIDs)

	return resourceFundingSourcePermissionsMappingRead(ctx, d, m)
}

func resourceFundingSourcePermissionsMappingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Expecting ID format "{funding_source_id}-{app_role_id}"
	parts := strings.Split(d.Id(), "-")
	if len(parts) != 2 {
		return diag.Errorf("invalid resource ID format, expected {funding_source_id}-{app_role_id}")
	}

	fundingSourceID, err := strconv.Atoi(parts[0])
	if err != nil {
		return diag.FromErr(err)
	}
	appRoleID, err := strconv.Atoi(parts[1])
	if err != nil {
		return diag.FromErr(err)
	}

	// Create the mapping with empty user IDs and user group IDs
	mapping := hc.FundingSourcePermissionsMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: []int{},
		UserIDs:       []int{},
	}

	err = client.PATCH(fmt.Sprintf("/v3/funding-source/%d/permission-mapping", fundingSourceID), []hc.FundingSourcePermissionsMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceFundingSourcePermissionsMappingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ID format, expected {funding_source_id}-{app_role_id}")
	}

	fundingSourceID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid Funding Source ID, must be an integer")
	}
	appRoleID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid app role ID, must be an integer")
	}

	d.Set("funding_source_id", fundingSourceID)
	d.Set("app_role_id", appRoleID)

	return []*schema.ResourceData{d}, nil
}
