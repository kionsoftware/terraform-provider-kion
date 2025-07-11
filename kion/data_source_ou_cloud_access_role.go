package kion

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceOUCloudAccessRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOUCloudAccessRoleRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The ID of the OU cloud access role to look up.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"aws_iam_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_iam_permissions_boundary": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"aws_iam_policies": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Type:     schema.TypeSet,
				Computed: true,
			},
			"aws_iam_role_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"azure_role_definitions": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Type:     schema.TypeSet,
				Computed: true,
			},
			"gcp_iam_roles": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Type:     schema.TypeSet,
				Computed: true,
			},
			"long_term_access_keys": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ou_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"short_term_access_keys": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"user_groups": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Type:     schema.TypeSet,
				Computed: true,
			},
			"users": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Type:     schema.TypeSet,
				Computed: true,
			},
			"web_access": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceOUCloudAccessRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Get("id").(int)

	resp := new(hc.OUCloudAccessRoleResponse)
	err := client.GET(fmt.Sprintf("/v3/ou-cloud-access-role/%d", ID), resp)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read OU Cloud Access Role",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
	}
	item := resp.Data

	data := map[string]interface{}{
		"aws_iam_path": item.OUCloudAccessRole.AwsIamPath,
		"aws_iam_role_name": func() string {
			if item.OUCloudAccessRole.AwsIamRoleName != nil {
				return *item.OUCloudAccessRole.AwsIamRoleName
			}
			return ""
		}(),
		"aws_iam_permissions_boundary": hc.InflateSingleObjectWithID(item.AwsIamPermissionsBoundary),
		"long_term_access_keys":        item.OUCloudAccessRole.LongTermAccessKeys,
		"name":                         item.OUCloudAccessRole.Name,
		"ou_id":                        item.OUCloudAccessRole.OUID,
		"short_term_access_keys":       item.OUCloudAccessRole.ShortTermAccessKeys,
		"web_access":                   item.OUCloudAccessRole.WebAccess,
		"aws_iam_policies":             hc.InflateObjectWithID(item.AwsIamPolicies),
		"azure_role_definitions":       hc.InflateObjectWithID(item.AzureRoleDefinitions),
		"gcp_iam_roles":                hc.InflateObjectWithID(item.GCPIamRoles),
		"users":                        hc.InflateObjectWithID(item.Users),
		"user_groups":                  hc.InflateObjectWithID(item.UserGroups),
	}

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set OU Cloud Access Role",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
		}
	}

	// Set the resource ID to the same ID that was used for lookup
	d.SetId(strconv.Itoa(ID))

	return diags
}
