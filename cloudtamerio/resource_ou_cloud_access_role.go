package kion

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/cloudtamerio/internal/ctclient"
)

func resourceOUCloudAccessRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOUCloudAccessRoleCreate,
		ReadContext:   resourceOUCloudAccessRoleRead,
		UpdateContext: resourceOUCloudAccessRoleUpdate,
		DeleteContext: resourceOUCloudAccessRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceOUCloudAccessRoleRead(ctx, d, m)
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
			"aws_iam_path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"aws_iam_permissions_boundary": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"aws_iam_policies": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeList,
				Optional: true,
			},
			"aws_iam_role_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"long_term_access_keys": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ou_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"short_term_access_keys": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"user_groups": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeList,
				Optional: true,
			},
			"users": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeList,
				Optional: true,
			},
			"web_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceOUCloudAccessRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	post := hc.OUCloudAccessRoleCreate{
		AwsIamPath:                d.Get("aws_iam_path").(string),
		AwsIamPermissionsBoundary: hc.FlattenIntPointer(d, "aws_iam_permissions_boundary"),
		AwsIamPolicies:            hc.FlattenGenericIDPointer(d, "aws_iam_policies"),
		AwsIamRoleName:            d.Get("aws_iam_role_name").(string),
		LongTermAccessKeys:        d.Get("long_term_access_keys").(bool),
		Name:                      d.Get("name").(string),
		OUID:                      d.Get("ou_id").(int),
		ShortTermAccessKeys:       d.Get("short_term_access_keys").(bool),
		UserGroupIds:              hc.FlattenGenericIDPointer(d, "user_groups"),
		UserIds:                   hc.FlattenGenericIDPointer(d, "users"),
		WebAccess:                 d.Get("web_access").(bool),
	}

	resp, err := c.POST("/v3/ou-cloud-access-role", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create OUCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create OUCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceOUCloudAccessRoleRead(ctx, d, m)

	return diags
}

func resourceOUCloudAccessRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.OUCloudAccessRoleResponse)
	err := c.GET(fmt.Sprintf("/v3/ou-cloud-access-role/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read OUCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["aws_iam_path"] = item.OUCloudAccessRole.AwsIamPath
	if hc.InflateSingleObjectWithID(item.AwsIamPermissionsBoundary) != nil {
		data["aws_iam_permissions_boundary"] = hc.InflateSingleObjectWithID(item.AwsIamPermissionsBoundary)
	}
	if hc.InflateObjectWithID(item.AwsIamPolicies) != nil {
		data["aws_iam_policies"] = hc.InflateObjectWithID(item.AwsIamPolicies)
	}
	data["aws_iam_role_name"] = item.OUCloudAccessRole.AwsIamRoleName
	data["long_term_access_keys"] = item.OUCloudAccessRole.LongTermAccessKeys
	data["name"] = item.OUCloudAccessRole.Name
	data["ou_id"] = item.OUCloudAccessRole.OUID
	data["short_term_access_keys"] = item.OUCloudAccessRole.ShortTermAccessKeys
	if hc.InflateObjectWithID(item.UserGroups) != nil {
		data["user_groups"] = hc.InflateObjectWithID(item.UserGroups)
	}
	if hc.InflateObjectWithID(item.Users) != nil {
		data["users"] = hc.InflateObjectWithID(item.Users)
	}
	data["web_access"] = item.OUCloudAccessRole.WebAccess

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set OUCloudAccessRole",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceOUCloudAccessRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("long_term_access_keys",
		"name",
		"short_term_access_keys",
		"web_access") {
		hasChanged++
		req := hc.OUCloudAccessRoleUpdate{
			LongTermAccessKeys:  d.Get("long_term_access_keys").(bool),
			Name:                d.Get("name").(string),
			ShortTermAccessKeys: d.Get("short_term_access_keys").(bool),
			WebAccess:           d.Get("web_access").(bool),
		}

		err := c.PATCH(fmt.Sprintf("/v3/ou-cloud-access-role/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update OUCloudAccessRole",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Handle associations.
	if d.HasChanges("aws_iam_permissions_boundary",
		"aws_iam_policies",
		"user_groups",
		"users") {
		hasChanged++
		arrAddAwsIamPermissionsBoundary, arrRemoveAwsIamPermissionsBoundary, _, _ := hc.AssociationChangedInt(d, "aws_iam_permissions_boundary")
		arrAddAwsIamPolicies, arrRemoveAwsIamPolicies, _, _ := hc.AssociationChanged(d, "aws_iam_policies")
		arrAddUserGroupIds, arrRemoveUserGroupIds, _, _ := hc.AssociationChanged(d, "user_groups")
		arrAddUserIds, arrRemoveUserIds, _, _ := hc.AssociationChanged(d, "users")

		if arrAddAwsIamPermissionsBoundary != nil ||
			len(arrAddAwsIamPolicies) > 0 ||
			len(arrAddUserGroupIds) > 0 ||
			len(arrAddUserIds) > 0 {
			_, err := c.POST(fmt.Sprintf("/v3/ou-cloud-access-role/%s/association", ID), hc.OUCloudAccessRoleAssociationsAdd{
				AwsIamPermissionsBoundary: arrAddAwsIamPermissionsBoundary,
				AwsIamPolicies:            &arrAddAwsIamPolicies,
				UserGroupIds:              &arrAddUserGroupIds,
				UserIds:                   &arrAddUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on OUCloudAccessRole",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if arrRemoveAwsIamPermissionsBoundary != nil ||
			len(arrRemoveAwsIamPolicies) > 0 ||
			len(arrRemoveUserGroupIds) > 0 ||
			len(arrRemoveUserIds) > 0 {
			err := c.DELETE(fmt.Sprintf("/v3/ou-cloud-access-role/%s/association", ID), hc.OUCloudAccessRoleAssociationsRemove{
				AwsIamPermissionsBoundary: arrRemoveAwsIamPermissionsBoundary,
				AwsIamPolicies:            &arrRemoveAwsIamPolicies,
				UserGroupIds:              &arrRemoveUserGroupIds,
				UserIds:                   &arrRemoveUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on OUCloudAccessRole",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged > 0 {
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceOUCloudAccessRoleRead(ctx, d, m)
}

func resourceOUCloudAccessRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	err := c.DELETE(fmt.Sprintf("/v3/ou-cloud-access-role/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete OUCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
