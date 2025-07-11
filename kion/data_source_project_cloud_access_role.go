package kion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// dataSourceProjectCloudAccessRole returns a schema.Resource for reading a specific project cloud access role by ID from Kion.
func dataSourceProjectCloudAccessRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProjectCloudAccessRoleRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the project cloud access role to retrieve.",
			},
			"accounts": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The accounts associated with this project cloud access role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the account.",
						},
					},
				},
			},
			"apply_to_all_accounts": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this role applies to all accounts.",
			},
			"aws_iam_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The AWS IAM path for the role.",
			},
			"aws_iam_permissions_boundary": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The AWS IAM permissions boundary ID.",
			},
			"aws_iam_policies": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The AWS IAM policies associated with this role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the AWS IAM policy.",
						},
					},
				},
			},
			"aws_iam_role_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The AWS IAM role name.",
			},
			"azure_role_definitions": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The Azure role definitions associated with this role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the Azure role definition.",
						},
					},
				},
			},
			"future_accounts": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this role applies to future accounts.",
			},
			"gcp_iam_roles": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The GCP IAM roles associated with this role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the GCP IAM role.",
						},
					},
				},
			},
			"long_term_access_keys": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether long-term access keys are enabled.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the project cloud access role.",
			},
			"project_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the project this role belongs to.",
			},
			"short_term_access_keys": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether short-term access keys are enabled.",
			},
			"user_groups": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The user groups associated with this role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the user group.",
						},
					},
				},
			},
			"users": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The users associated with this role.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the user.",
						},
					},
				},
			},
			"web_access": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether web access is enabled.",
			},
		},
	}
}

// dataSourceProjectCloudAccessRoleRead retrieves data for a specific project cloud access role by ID.
func dataSourceProjectCloudAccessRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	projectCloudAccessRoleID := d.Get("id").(string)

	// Fetch the specific project cloud access role by ID using the same endpoint as the resource
	resp := new(hc.ProjectCloudAccessRoleResponse)
	err := client.GET(fmt.Sprintf("/v3/project-cloud-access-role/%s", projectCloudAccessRoleID), resp)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read project cloud access role by ID: %v", err))
	}

	item := resp.Data

	// Set the ID
	d.SetId(projectCloudAccessRoleID)

	// Set all the computed fields using the same logic as the resource read function
	diags = append(diags, hc.SafeSet(d, "apply_to_all_accounts", item.ProjectCloudAccessRole.ApplyToAllAccounts, "Failed to set apply_to_all_accounts")...)
	diags = append(diags, hc.SafeSet(d, "aws_iam_path", item.ProjectCloudAccessRole.AwsIamPath, "Failed to set aws_iam_path")...)
	
	// Handle aws_iam_role_name (nullable string)
	awsIamRoleName := ""
	if item.ProjectCloudAccessRole.AwsIamRoleName != nil {
		awsIamRoleName = *item.ProjectCloudAccessRole.AwsIamRoleName
	}
	diags = append(diags, hc.SafeSet(d, "aws_iam_role_name", awsIamRoleName, "Failed to set aws_iam_role_name")...)
	
	// Handle aws_iam_permissions_boundary (single object with ID)
	awsIamPermissionsBoundary := hc.InflateSingleObjectWithID(item.AwsIamPermissionsBoundary)
	diags = append(diags, hc.SafeSet(d, "aws_iam_permissions_boundary", awsIamPermissionsBoundary, "Failed to set aws_iam_permissions_boundary")...)
	
	diags = append(diags, hc.SafeSet(d, "long_term_access_keys", item.ProjectCloudAccessRole.LongTermAccessKeys, "Failed to set long_term_access_keys")...)
	diags = append(diags, hc.SafeSet(d, "name", item.ProjectCloudAccessRole.Name, "Failed to set name")...)
	diags = append(diags, hc.SafeSet(d, "project_id", item.ProjectCloudAccessRole.ProjectID, "Failed to set project_id")...)
	diags = append(diags, hc.SafeSet(d, "short_term_access_keys", item.ProjectCloudAccessRole.ShortTermAccessKeys, "Failed to set short_term_access_keys")...)
	diags = append(diags, hc.SafeSet(d, "web_access", item.ProjectCloudAccessRole.WebAccess, "Failed to set web_access")...)
	diags = append(diags, hc.SafeSet(d, "future_accounts", item.ProjectCloudAccessRole.FutureAccounts, "Failed to set future_accounts")...)

	// Set association fields using the same inflate functions as the resource
	diags = append(diags, hc.SafeSet(d, "accounts", hc.InflateObjectWithID(item.Accounts), "Failed to set accounts")...)
	diags = append(diags, hc.SafeSet(d, "aws_iam_policies", hc.InflateObjectWithID(item.AwsIamPolicies), "Failed to set aws_iam_policies")...)
	diags = append(diags, hc.SafeSet(d, "azure_role_definitions", hc.InflateObjectWithID(item.AzureRoleDefinitions), "Failed to set azure_role_definitions")...)
	diags = append(diags, hc.SafeSet(d, "gcp_iam_roles", hc.InflateObjectWithID(item.GCPIamRoles), "Failed to set gcp_iam_roles")...)
	diags = append(diags, hc.SafeSet(d, "user_groups", hc.InflateObjectWithID(item.UserGroups), "Failed to set user_groups")...)
	diags = append(diags, hc.SafeSet(d, "users", hc.InflateObjectWithID(item.Users), "Failed to set users")...)

	return diags
}