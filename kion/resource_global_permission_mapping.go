package kion

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

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
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "List of user group IDs for the permission mapping.",
			},
			"user_ids": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "List of user IDs for the permission mapping.",
			},
		},
	}
}

func resourceGlobalPermissionMappingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	appRoleID := d.Get("app_role_id").(int)

	mapping := hc.GlobalPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").([]interface{})),
		UserIDs:       hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").([]interface{})),
	}

	_, err := client.POST("/v3/global/permission-mapping", []hc.GlobalPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("global-%d", appRoleID))

	return resourceGlobalPermissionMappingRead(ctx, d, m)
}

func resourceGlobalPermissionMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	parts := strings.Split(d.Id(), "-")
	if len(parts) != 2 {
		return diag.Errorf("invalid resource ID format, expected global-{app_role_id}")
	}

	appRoleID, err := strconv.Atoi(parts[1])
	if err != nil {
		return diag.FromErr(err)
	}

	resp := new(hc.GlobalPermissionMappingListResponse)
	err = client.GET("/v3/global/permission-mapping", resp)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	found := false
	for _, mapping := range resp.Data {
		if mapping.AppRoleID == appRoleID {
			sort.Ints(mapping.UserGroupsIDs)
			sort.Ints(mapping.UserIDs)

			diags = append(diags, hc.SafeSet(d, "app_role_id", mapping.AppRoleID)...)
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

func resourceGlobalPermissionMappingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	appRoleID := d.Get("app_role_id").(int)

	mapping := hc.GlobalPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").([]interface{})),
		UserIDs:       hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").([]interface{})),
	}

	resp := new(hc.GlobalPermissionMappingListResponse)
	err := client.GET("/v3/global/permission-mapping", resp)
	if err != nil {
		return diag.FromErr(err)
	}

	existingMappings := resp.Data
	updatedMappings := make([]hc.GlobalPermissionMapping, 0)
	found := false
	for _, existing := range existingMappings {
		if existing.AppRoleID == appRoleID {
			updatedMappings = append(updatedMappings, hc.GlobalPermissionMapping(existing))
			found = true
		} else {
			updatedMappings = append(updatedMappings, hc.GlobalPermissionMapping(existing))
		}
	}

	if !found {
		updatedMappings = append(updatedMappings, mapping)
	}

	err = client.PATCH("/v3/global/permission-mapping", updatedMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceGlobalPermissionMappingRead(ctx, d, m)
}

func resourceGlobalPermissionMappingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	parts := strings.Split(d.Id(), "-")
	if len(parts) != 2 {
		return diag.Errorf("invalid resource ID format, expected global-{app_role_id}")
	}

	appRoleID, err := strconv.Atoi(parts[1])
	if err != nil {
		return diag.FromErr(err)
	}

	resp := new(hc.GlobalPermissionMappingListResponse)
	err = client.GET("/v3/global/permission-mapping", resp)
	if err != nil {
		return diag.FromErr(err)
	}

	remainingMappings := make([]hc.GlobalPermissionMapping, 0)
	for _, existing := range resp.Data {
		if existing.AppRoleID != appRoleID {
			remainingMappings = append(remainingMappings, hc.GlobalPermissionMapping(existing))
		}
	}

	err = client.PATCH("/v3/global/permission-mapping", remainingMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceGlobalPermissionMappingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ID format, expected global-{app_role_id}")
	}

	appRoleID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid app role ID, must be an integer")
	}

	d.Set("app_role_id", appRoleID)

	return []*schema.ResourceData{d}, nil
}
