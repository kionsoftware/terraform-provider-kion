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

func resourceOUPermissionMapping() *schema.Resource {
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
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "List of user group IDs for the permission mapping (must be provided in numerical order).",
			},
			"user_ids": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "List of user IDs for the permission mapping (must be provided in numerical order).",
			},
		},
	}
}

func resourceOUPermissionMappingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	ouID := d.Get("ou_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	userGroupsIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").([]interface{}))
	userIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").([]interface{}))

	mapping := hc.OUPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: userGroupsIDs,
		UserIDs:       userIDs,
	}

	err := client.PATCH(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), []hc.OUPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("OU-%d-%d", ouID, appRoleID))

	// Ensure the state reflects the provided list
	d.Set("user_groups_ids", userGroupsIDs)
	d.Set("user_ids", userIDs)

	return resourceOUPermissionMappingRead(ctx, d, m)
}

func resourceOUPermissionMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	parts := strings.Split(d.Id(), "-")
	if len(parts) != 3 {
		return diag.Errorf("invalid resource ID format, expected OU-{ou_id}-{app_role_id}")
	}

	ouID, err := strconv.Atoi(parts[1])
	if err != nil {
		return diag.FromErr(err)
	}
	appRoleID, err := strconv.Atoi(parts[2])
	if err != nil {
		return diag.FromErr(err)
	}

	resp := new(hc.OUPermissionMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	found := false
	for _, mapping := range resp.Data {
		if mapping.AppRoleID == appRoleID {
			// Set lists to the state as provided
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

func resourceOUPermissionMappingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	ouID := d.Get("ou_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	userGroupsIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").([]interface{}))
	userIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").([]interface{}))

	mapping := hc.OUPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: userGroupsIDs,
		UserIDs:       userIDs,
	}

	resp := new(hc.OUPermissionMappingListResponse)
	err := client.GET(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	updatedMappings := make([]hc.OUPermissionMapping, 0)
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

	err = client.PATCH(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), updatedMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ensure the state reflects the provided list
	d.Set("user_groups_ids", userGroupsIDs)
	d.Set("user_ids", userIDs)

	return resourceOUPermissionMappingRead(ctx, d, m)
}

func resourceOUPermissionMappingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	parts := strings.Split(d.Id(), "-")
	if len(parts) != 3 {
		return diag.Errorf("invalid resource ID format, expected OU-{ou_id}-{app_role_id}")
	}

	ouID, err := strconv.Atoi(parts[1])
	if err != nil {
		return diag.FromErr(err)
	}
	appRoleID, err := strconv.Atoi(parts[2])
	if err != nil {
		return diag.FromErr(err)
	}

	resp := new(hc.OUPermissionMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	remainingMappings := make([]hc.OUPermissionMapping, 0)
	for _, existing := range resp.Data {
		if existing.AppRoleID != appRoleID {
			remainingMappings = append(remainingMappings, hc.OUPermissionMapping(existing))
		}
	}

	err = client.PATCH(fmt.Sprintf("/v3/ou/%d/permission-mapping", ouID), remainingMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceOUPermissionMappingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID format, expected OU-{ou_id}-{app_role_id}")
	}

	ouID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid OU ID, must be an integer")
	}
	appRoleID, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid app role ID, must be an integer")
	}

	d.Set("ou_id", ouID)
	d.Set("app_role_id", appRoleID)

	return []*schema.ResourceData{d}, nil
}
