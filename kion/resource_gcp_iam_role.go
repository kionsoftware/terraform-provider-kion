package kion

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/ctclient"
)

func resourceGcpIamRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGcpIamRoleCreate,
		ReadContext:   resourceGcpIamRoleRead,
		UpdateContext: resourceGcpIamRoleUpdate,
		DeleteContext: resourceGcpIamRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceGcpIamRoleRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			// Notice there is no 'id' field specified because it will be created.
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_permissions": {
				Elem:     &schema.Schema{Type: schema.TypeString},
				Type:     schema.TypeList,
				Required: true,
			},
			"gcp_role_launch_stage": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"gcp_managed_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"gcp_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"system_managed_policy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"owner_user_groups": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:         schema.TypeList,
				Optional:     true,
				AtLeastOneOf: []string{"owner_user_groups", "owner_users"},
			},
			"owner_users": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:         schema.TypeList,
				Optional:     true,
				AtLeastOneOf: []string{"owner_user_groups", "owner_users"},
			},
		},
	}
}

func resourceGcpIamRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	post := hc.GCPRoleCreate{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		RolePermissions:    hc.FlattenStringArray(d.Get("role_permissions").([]interface{})),
		OwnerUserIDs:       hc.FlattenGenericIDPointer(d, "owner_users"),
		OwnerUGroupIDs:     hc.FlattenGenericIDPointer(d, "owner_user_groups"),
		GCPRoleLaunchStage: d.Get("gcp_role_launch_stage").(int),
	}

	resp, err := c.POST("/v3/gcp-iam-role", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create GcpIamRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create GcpIamRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceGcpIamRoleRead(ctx, d, m)

	return diags
}

func resourceGcpIamRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.GCPRoleResponseWithOwners)
	err := c.GET(fmt.Sprintf("/v3/gcp-iam-role/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read GcpIamRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["name"] = item.GcpRole.Name
	data["description"] = item.GcpRole.Description
	if hc.InflateObjectWithID(item.OwnerUserGroups) != nil {
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
	}
	if hc.InflateObjectWithID(item.OwnerUsers) != nil {
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
	}
	data["gcp_id"] = item.GcpRole.GCPID
	data["gcp_role_launch_stage"] = item.GcpRole.GCPRoleLaunchStage
	data["role_permissions"] = hc.FilterStringArray(item.GcpRole.RolePermissions)
	data["gcp_managed_policy"] = item.GcpRole.GCPManagedPolicy
	data["system_managed_policy"] = item.GcpRole.SystemManagedPolicy

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set GcpIamRole",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v\nData: %+v\nPayload: %+v", err.Error(), ID, data, resp),
			})
			return diags
		}
	}

	return diags
}

func resourceGcpIamRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges(
		"description",
		"name",
		"gcp_role_launch_stage",
		"role_permissions") {
		hasChanged++
		req := hc.GCPRoleUpdate{
			Description:        d.Get("description").(string),
			Name:               d.Get("name").(string),
			GCPRoleLaunchStage: d.Get("gcp_role_launch_stage").(int),
			RolePermissions:    hc.FlattenStringArray(d.Get("role_permissions").([]interface{})),
		}

		err := c.PATCH(fmt.Sprintf("/v3/gcp-iam-role/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update GcpIamRole",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Determine if the owners have changed.
	if d.HasChanges("owner_user_groups",
		"owner_users") {
		hasChanged++
		arrAddOwnerUserGroupIds, arrRemoveOwnerUserGroupIds, _, _ := hc.AssociationChanged(d, "owner_user_groups")
		arrAddOwnerUserIds, arrRemoveOwnerUserIds, _, _ := hc.AssociationChanged(d, "owner_users")

		if len(arrAddOwnerUserGroupIds) > 0 ||
			len(arrAddOwnerUserIds) > 0 {
			_, err := c.POST(fmt.Sprintf("/v3/gcp-iam-role/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on GcpIamRole",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 ||
			len(arrRemoveOwnerUserIds) > 0 {
			err := c.DELETE(fmt.Sprintf("/v3/gcp-iam-role/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrRemoveOwnerUserGroupIds,
				OwnerUserIds:      &arrRemoveOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on GcpIamRole",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged > 0 {
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceGcpIamRoleRead(ctx, d, m)
}

func resourceGcpIamRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	err := c.DELETE(fmt.Sprintf("/v3/gcp-iam-role/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete GcpIamRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
