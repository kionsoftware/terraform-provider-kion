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

func resourceProjectCloudAccessRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCloudAccessRoleCreate,
		ReadContext:   resourceProjectCloudAccessRoleRead,
		UpdateContext: resourceProjectCloudAccessRoleUpdate,
		DeleteContext: resourceProjectCloudAccessRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceProjectCloudAccessRoleRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			// Notice there is no 'id' field specified because it will be created.
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"accounts": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:        schema.TypeSet, // Don't let codegen remove this.
				Optional:    true,
				Description: "This field will be ignored if 'apply_to_all_accounts' is set to: true.",
				// If apply_to_all_accounts is true, then ignore the accounts.
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("apply_to_all_accounts").(bool)
				},
			},
			"apply_to_all_accounts": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false, // Don't let codegen remove this.
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
				Type:     schema.TypeSet,
				Optional: true,
			},
			"aws_iam_role_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"azure_role_definitions": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeSet,
				Optional: true,
			},
			"future_accounts": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"gcp_iam_roles": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:     schema.TypeSet,
				Optional: true,
			},
			"long_term_access_keys": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project_id": {
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
				Type:     schema.TypeSet,
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
				Type:     schema.TypeSet,
				Optional: true,
			},
			"web_access": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceProjectCloudAccessRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	// Create the request payload
	post := hc.ProjectCloudAccessRoleCreate{
		AccountIds:                hc.FlattenGenericIDPointer(d, "accounts"),
		ApplyToAllAccounts:        d.Get("apply_to_all_accounts").(bool),
		AwsIamPath:                d.Get("aws_iam_path").(string),
		AwsIamPermissionsBoundary: hc.FlattenIntPointer(d, "aws_iam_permissions_boundary"),
		AwsIamPolicies:            hc.FlattenGenericIDPointer(d, "aws_iam_policies"),
		AwsIamRoleName:            d.Get("aws_iam_role_name").(string),
		AzureRoleDefinitions:      hc.FlattenGenericIDPointer(d, "azure_role_definitions"),
		FutureAccounts:            d.Get("future_accounts").(bool),
		GCPIamRoles:               hc.FlattenGenericIDPointer(d, "gcp_iam_roles"),
		LongTermAccessKeys:        d.Get("long_term_access_keys").(bool),
		Name:                      d.Get("name").(string),
		ProjectID:                 d.Get("project_id").(int),
		ShortTermAccessKeys:       d.Get("short_term_access_keys").(bool),
		UserGroupIds:              hc.FlattenGenericIDPointer(d, "user_groups"),
		UserIds:                   hc.FlattenGenericIDPointer(d, "users"),
		WebAccess:                 d.Get("web_access").(bool),
	}

	// Send the POST request
	resp, err := client.POST("/v3/project-cloud-access-role", post)
	if err != nil {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create ProjectCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err, post),
		})
	}

	if resp.RecordID == 0 {
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create ProjectCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
	}

	// Set the ID for the created resource
	d.SetId(strconv.Itoa(resp.RecordID))

	// Read the resource to update the state
	if readDiags := resourceProjectCloudAccessRoleRead(ctx, d, m); readDiags.HasError() {
		diags = append(diags, readDiags...)
	}

	return diags
}

func resourceProjectCloudAccessRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.ProjectCloudAccessRoleResponse)
	err := client.GET(fmt.Sprintf("/v3/project-cloud-access-role/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read ProjectCloudAccessRole",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := map[string]interface{}{
		"apply_to_all_accounts":        item.ProjectCloudAccessRole.ApplyToAllAccounts,
		"aws_iam_path":                 item.ProjectCloudAccessRole.AwsIamPath,
		"aws_iam_role_name":            item.ProjectCloudAccessRole.AwsIamRoleName,
		"aws_iam_permissions_boundary": hc.InflateSingleObjectWithID(item.AwsIamPermissionsBoundary),
		"long_term_access_keys":        item.ProjectCloudAccessRole.LongTermAccessKeys,
		"name":                         item.ProjectCloudAccessRole.Name,
		"project_id":                   item.ProjectCloudAccessRole.ProjectID,
		"short_term_access_keys":       item.ProjectCloudAccessRole.ShortTermAccessKeys,
		"web_access":                   item.ProjectCloudAccessRole.WebAccess,
		"accounts":                     hc.InflateObjectWithID(item.Accounts),
		"aws_iam_policies":             hc.InflateObjectWithID(item.AwsIamPolicies),
		"azure_role_definitions":       hc.InflateObjectWithID(item.AzureRoleDefinitions),
		"gcp_iam_roles":                hc.InflateObjectWithID(item.GCPIamRoles),
		"user_groups":                  hc.InflateObjectWithID(item.UserGroups),
		"users":                        hc.InflateObjectWithID(item.Users),
		"future_accounts":              item.ProjectCloudAccessRole.FutureAccounts,
	}

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set ProjectCloudAccessRole",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceProjectCloudAccessRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	var hasChanged bool

	// Determine if the attributes that are updatable are changed.
	if d.HasChanges("apply_to_all_accounts", "future_accounts", "long_term_access_keys", "name", "short_term_access_keys", "web_access") {
		hasChanged = true
		req := hc.ProjectCloudAccessRoleUpdate{
			ApplyToAllAccounts:  d.Get("apply_to_all_accounts").(bool),
			FutureAccounts:      d.Get("future_accounts").(bool),
			LongTermAccessKeys:  d.Get("long_term_access_keys").(bool),
			Name:                d.Get("name").(string),
			ShortTermAccessKeys: d.Get("short_term_access_keys").(bool),
			WebAccess:           d.Get("web_access").(bool),
		}

		if err := client.PATCH(fmt.Sprintf("/v3/project-cloud-access-role/%s", ID), req); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update ProjectCloudAccessRole",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Handle associations.
	if d.HasChanges("accounts", "aws_iam_permissions_boundary", "aws_iam_policies", "azure_role_definitions", "gcp_iam_roles", "user_groups", "users") {
		hasChanged = true

		var addCarAssociation hc.ProjectCloudAccessRoleAssociationsAdd
		var removeCarAssociation hc.ProjectCloudAccessRoleAssociationsRemove

		if addBoundary, removeBoundary, _, _ := hc.AssociationChangedInt(d, "aws_iam_permissions_boundary"); addBoundary != nil || removeBoundary != nil {
			addCarAssociation.AwsIamPermissionsBoundary = addBoundary
			removeCarAssociation.AwsIamPermissionsBoundary = removeBoundary
		}

		if arrAdd, arrRemove, _, _ := hc.AssociationChanged(d, "accounts"); len(arrAdd) > 0 || len(arrRemove) > 0 {
			addCarAssociation.AccountIds = &arrAdd
			removeCarAssociation.AccountIds = &arrRemove
		}

		if arrAdd, arrRemove, _, _ := hc.AssociationChanged(d, "aws_iam_policies"); len(arrAdd) > 0 || len(arrRemove) > 0 {
			addCarAssociation.AwsIamPolicies = &arrAdd
			removeCarAssociation.AwsIamPolicies = &arrRemove
		}

		if arrAdd, arrRemove, _, _ := hc.AssociationChanged(d, "azure_role_definitions"); len(arrAdd) > 0 || len(arrRemove) > 0 {
			addCarAssociation.AzureRoleDefinitions = &arrAdd
			removeCarAssociation.AzureRoleDefinitions = &arrRemove
		}

		if arrAdd, arrRemove, _, _ := hc.AssociationChanged(d, "gcp_iam_roles"); len(arrAdd) > 0 || len(arrRemove) > 0 {
			addCarAssociation.GCPIamRoles = &arrAdd
			removeCarAssociation.GCPIamRoles = &arrRemove
		}

		if arrAdd, arrRemove, _, _ := hc.AssociationChanged(d, "user_groups"); len(arrAdd) > 0 || len(arrRemove) > 0 {
			addCarAssociation.UserGroupIds = &arrAdd
			removeCarAssociation.UserGroupIds = &arrRemove
		}

		if arrAdd, arrRemove, _, _ := hc.AssociationChanged(d, "users"); len(arrAdd) > 0 || len(arrRemove) > 0 {
			addCarAssociation.UserIds = &arrAdd
			removeCarAssociation.UserIds = &arrRemove
		}

		if addCarAssociation != (hc.ProjectCloudAccessRoleAssociationsAdd{}) {
			if _, err := client.POST(fmt.Sprintf("/v3/project-cloud-access-role/%s/association", ID), addCarAssociation); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add associations on ProjectCloudAccessRole",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if removeCarAssociation != (hc.ProjectCloudAccessRoleAssociationsRemove{}) {
			if err := client.DELETE(fmt.Sprintf("/v3/project-cloud-access-role/%s/association", ID), removeCarAssociation); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove associations on ProjectCloudAccessRole",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged {
		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set last_updated",
				Detail:   fmt.Sprintf("Error: %v", err),
			})
			return diags
		}
	}

	return resourceProjectCloudAccessRoleRead(ctx, d, m)
}

func resourceProjectCloudAccessRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	// Make the DELETE request using the client and context
	err := client.DELETE(fmt.Sprintf("/v3/project-cloud-access-role/%s", ID), nil)
	if err != nil {
		// Add detailed diagnostic information on error
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete ProjectCloudAccessRole",
			Detail:   fmt.Sprintf("Error deleting ProjectCloudAccessRole with ID %s: %v", ID, err),
		})
		return diags
	}

	// Explicitly clear the resource ID to indicate deletion
	d.SetId("")

	return diags
}
