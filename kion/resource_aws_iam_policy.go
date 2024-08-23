package kion

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceAwsIamPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsIamPolicyCreate,
		ReadContext:   resourceAwsIamPolicyRead,
		UpdateContext: resourceAwsIamPolicyUpdate,
		DeleteContext: resourceAwsIamPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceAwsIamPolicyRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			// Notice there is no 'id' field specified because it will be created.
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_iam_path": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"aws_managed_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
				Type:         schema.TypeSet,
				Optional:     true,
				Description:  "Must provide at least the owner_user_groups field or the owner_users field.",
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
				Type:         schema.TypeSet,
				Optional:     true,
				Description:  "Must provide at least the owner_user_groups field or the owner_users field.",
				AtLeastOneOf: []string{"owner_user_groups", "owner_users"},
			},
			"path_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"system_managed_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceAwsIamPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	post := hc.IAMPolicyCreate{
		AwsIamPath:        d.Get("aws_iam_path").(string),
		Description:       d.Get("description").(string),
		Name:              d.Get("name").(string),
		OwnerUserGroupIds: hc.FlattenGenericIDPointer(d, "owner_user_groups"),
		OwnerUserIds:      hc.FlattenGenericIDPointer(d, "owner_users"),
		Policy:            d.Get("policy").(string),
	}

	resp, err := client.POST("/v3/iam-policy", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create AwsIamPolicy",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create AwsIamPolicy",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceAwsIamPolicyRead(ctx, d, m)

	return diags
}

func resourceAwsIamPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.IAMPolicyResponse)
	err := client.GET(fmt.Sprintf("/v3/iam-policy/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read AwsIamPolicy",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["aws_iam_path"] = item.IamPolicy.AwsIamPath
	data["aws_managed_policy"] = item.IamPolicy.AwsManagedPolicy
	data["description"] = item.IamPolicy.Description
	data["name"] = item.IamPolicy.Name
	if hc.InflateObjectWithID(item.OwnerUserGroups) != nil {
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
	}
	if hc.InflateObjectWithID(item.OwnerUsers) != nil {
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
	}
	data["path_suffix"] = item.IamPolicy.PathSuffix
	data["policy"] = item.IamPolicy.Policy
	data["system_managed_policy"] = item.IamPolicy.SystemManagedPolicy

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set AwsIamPolicy",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceAwsIamPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("description",
		"name",
		"policy") {
		hasChanged++
		req := hc.IAMPolicyUpdate{
			Description: d.Get("description").(string),
			Name:        d.Get("name").(string),
			Policy:      d.Get("policy").(string),
		}

		err := client.PATCH(fmt.Sprintf("/v3/iam-policy/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update AwsIamPolicy",
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
			_, err := client.POST(fmt.Sprintf("/v3/iam-policy/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on AwsIamPolicy",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 ||
			len(arrRemoveOwnerUserIds) > 0 {
			err := client.DELETE(fmt.Sprintf("/v3/iam-policy/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrRemoveOwnerUserGroupIds,
				OwnerUserIds:      &arrRemoveOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on AwsIamPolicy",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged > 0 {
		diags = append(diags, hc.SafeSet(d, "last_updated", time.Now().Format(time.RFC850), "Failed to set last_updated")...)
		if len(diags) > 0 {
			return diags
		}
	}

	return resourceAwsIamPolicyRead(ctx, d, m)
}

func resourceAwsIamPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/iam-policy/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete AwsIamPolicy",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
