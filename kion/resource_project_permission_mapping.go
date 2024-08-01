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

func resourceProjectPermissionMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectPermissionMappingCreate,
		ReadContext:   resourceProjectPermissionMappingRead,
		UpdateContext: resourceProjectPermissionMappingUpdate,
		DeleteContext: resourceProjectPermissionMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectPermissionMappingImport,
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

func resourceProjectPermissionMappingCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	projectID := d.Get("project_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	userGroupsIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List())
	userIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List())

	mapping := hc.ProjectPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: userGroupsIDs,
		UserIDs:       userIDs,
	}

	err := client.PATCH(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), []hc.ProjectPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("project-%d-%d", projectID, appRoleID))

	// Ensure the state reflects the provided set
	d.Set("user_groups_ids", userGroupsIDs)
	d.Set("user_ids", userIDs)

	return resourceProjectPermissionMappingRead(ctx, d, m)
}

func resourceProjectPermissionMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	parts := strings.Split(d.Id(), "-")
	if len(parts) != 3 {
		return diag.Errorf("invalid resource ID format, expected project-{project_id}-{app_role_id}")
	}

	projectID, err := strconv.Atoi(parts[1])
	if err != nil {
		return diag.FromErr(err)
	}
	appRoleID, err := strconv.Atoi(parts[2])
	if err != nil {
		return diag.FromErr(err)
	}

	resp := new(hc.ProjectPermissionMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	found := false
	for _, mapping := range resp.Data {
		if mapping.AppRoleID == appRoleID {
			// Set sets to the state as provided
			diags = append(diags, hc.SafeSet(d, "project_id", projectID)...)
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

func resourceProjectPermissionMappingUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	projectID := d.Get("project_id").(int)
	appRoleID := d.Get("app_role_id").(int)

	userGroupsIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_groups_ids").(*schema.Set).List())
	userIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("user_ids").(*schema.Set).List())

	mapping := hc.ProjectPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: userGroupsIDs,
		UserIDs:       userIDs,
	}

	resp := new(hc.ProjectPermissionMappingListResponse)
	err := client.GET(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	updatedMappings := make([]hc.ProjectPermissionMapping, 0)
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

	err = client.PATCH(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), updatedMappings)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ensure the state reflects the provided set
	d.Set("user_groups_ids", userGroupsIDs)
	d.Set("user_ids", userIDs)

	return resourceProjectPermissionMappingRead(ctx, d, m)
}

func resourceProjectPermissionMappingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	parts := strings.Split(d.Id(), "-")
	if len(parts) != 3 {
		return diag.Errorf("invalid resource ID format, expected project-{project_id}-{app_role_id}")
	}

	projectID, err := strconv.Atoi(parts[1])
	if err != nil {
		return diag.FromErr(err)
	}
	appRoleID, err := strconv.Atoi(parts[2])
	if err != nil {
		return diag.FromErr(err)
	}

	// Create the mapping with empty user IDs and user group IDs
	mapping := hc.ProjectPermissionMapping{
		AppRoleID:     appRoleID,
		UserGroupsIDs: []int{},
		UserIDs:       []int{},
	}

	err = client.PATCH(fmt.Sprintf("/v3/project/%d/permission-mapping", projectID), []hc.ProjectPermissionMapping{mapping})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}

func resourceProjectPermissionMappingImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid ID format, expected project-{project_id}-{app_role_id}")
	}

	projectID, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid project ID, must be an integer")
	}
	appRoleID, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid app role ID, must be an integer")
	}

	d.Set("project_id", projectID)
	d.Set("app_role_id", appRoleID)

	return []*schema.ResourceData{d}, nil
}
