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

func resourceCloudRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudRuleCreate,
		ReadContext:   resourceCloudRuleRead,
		UpdateContext: resourceCloudRuleUpdate,
		DeleteContext: resourceCloudRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceCloudRuleRead(ctx, d, m)
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
			"aws_cloudformation_templates": {
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
			"azure_arm_template_definitions": {
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
			"azure_policy_definitions": {
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
			"built_in": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"compliance_standards": {
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"internal_aws_amis": {
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
			"internal_aws_service_catalog_portfolios": {
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ous": {
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
			"post_webhook_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"pre_webhook_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"projects": {
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
			"service_control_policies": {
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
		},
	}
}

func resourceCloudRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	post := hc.CloudRuleCreate{
		AzureArmTemplateDefinitionIds: hc.FlattenGenericIDPointer(d, "azure_arm_template_definitions"),
		AzurePolicyDefinitionIds:      hc.FlattenGenericIDPointer(d, "azure_policy_definitions"),
		AzureRoleDefinitionIds:        hc.FlattenGenericIDPointer(d, "azure_role_definitions"),
		CftIds:                        hc.FlattenGenericIDPointer(d, "aws_cloudformation_templates"),
		ComplianceStandardIds:         hc.FlattenGenericIDPointer(d, "compliance_standards"),
		Description:                   d.Get("description").(string),
		GcpIamRoleIds:                 hc.FlattenGenericIDPointer(d, "gcp_iam_roles"),
		IamPolicyIds:                  hc.FlattenGenericIDPointer(d, "aws_iam_policies"),
		InternalAmiIds:                hc.FlattenGenericIDPointer(d, "internal_aws_amis"),
		InternalPortfolioIds:          hc.FlattenGenericIDPointer(d, "internal_aws_service_catalog_portfolios"),
		Name:                          d.Get("name").(string),
		OUIds:                         hc.FlattenGenericIDPointer(d, "ous"),
		OwnerUserGroupIds:             hc.FlattenGenericIDPointer(d, "owner_user_groups"),
		OwnerUserIds:                  hc.FlattenGenericIDPointer(d, "owner_users"),
		PostWebhookID:                 hc.FlattenIntPointer(d, "post_webhook_id"),
		PreWebhookID:                  hc.FlattenIntPointer(d, "pre_webhook_id"),
		ProjectIds:                    hc.FlattenGenericIDPointer(d, "projects"),
		ServiceControlPolicyIds:       hc.FlattenGenericIDPointer(d, "service_control_policies"),
	}

	resp, err := c.POST("/v3/cloud-rule", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create CloudRule",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create CloudRule",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceCloudRuleRead(ctx, d, m)

	return diags
}

func resourceCloudRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.CloudRuleResponse)
	err := c.GET(fmt.Sprintf("/v3/cloud-rule/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read CloudRule",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	if hc.InflateObjectWithID(item.AwsCloudformationTemplates) != nil {
		data["aws_cloudformation_templates"] = hc.InflateObjectWithID(item.AwsCloudformationTemplates)
	}
	if hc.InflateObjectWithID(item.AwsIamPolicies) != nil {
		data["aws_iam_policies"] = hc.InflateObjectWithID(item.AwsIamPolicies)
	}
	if hc.InflateObjectWithID(item.AzureArmTemplateDefinitions) != nil {
		data["azure_arm_template_definitions"] = hc.InflateObjectWithID(item.AzureArmTemplateDefinitions)
	}
	if hc.InflateObjectWithID(item.AzurePolicyDefinitions) != nil {
		data["azure_policy_definitions"] = hc.InflateObjectWithID(item.AzurePolicyDefinitions)
	}
	if hc.InflateObjectWithID(item.AzureRoleDefinitions) != nil {
		data["azure_role_definitions"] = hc.InflateObjectWithID(item.AzureRoleDefinitions)
	}
	data["built_in"] = item.CloudRule.BuiltIn
	if hc.InflateObjectWithID(item.ComplianceStandards) != nil {
		data["compliance_standards"] = hc.InflateObjectWithID(item.ComplianceStandards)
	}
	data["description"] = item.CloudRule.Description
	if hc.InflateObjectWithID(item.InternalAwsAmis) != nil {
		data["internal_aws_amis"] = hc.InflateObjectWithID(item.InternalAwsAmis)
	}
	if hc.InflateObjectWithID(item.GCPIAMRoles) != nil {
		data["gcp_iam_roles"] = hc.InflateObjectWithID(item.GCPIAMRoles)
	}
	if hc.InflateObjectWithID(item.InternalAwsServiceCatalogPortfolios) != nil {
		data["internal_aws_service_catalog_portfolios"] = hc.InflateObjectWithID(item.InternalAwsServiceCatalogPortfolios)
	}
	data["name"] = item.CloudRule.Name
	if hc.InflateObjectWithID(item.OUs) != nil {
		data["ous"] = hc.InflateObjectWithID(item.OUs)
	}
	if hc.InflateObjectWithID(item.OwnerUserGroups) != nil {
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
	}
	if hc.InflateObjectWithID(item.OwnerUsers) != nil {
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
	}
	if item.CloudRule.PostWebhookID != nil {
		data["post_webhook_id"] = item.CloudRule.PostWebhookID
	}
	if item.CloudRule.PreWebhookID != nil {
		data["pre_webhook_id"] = item.CloudRule.PreWebhookID
	}
	if hc.InflateObjectWithID(item.Projects) != nil {
		data["projects"] = hc.InflateObjectWithID(item.Projects)
	}
	if hc.InflateObjectWithID(item.ServiceControlPolicies) != nil {
		data["service_control_policies"] = hc.InflateObjectWithID(item.ServiceControlPolicies)
	}

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set CloudRule",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceCloudRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("description",
		"name",
		"post_webhook_id",
		"pre_webhook_id") {
		hasChanged++
		req := hc.CloudRuleUpdate{
			Description:   d.Get("description").(string),
			Name:          d.Get("name").(string),
			PostWebhookID: hc.FlattenIntPointer(d, "post_webhook_id"),
			PreWebhookID:  hc.FlattenIntPointer(d, "pre_webhook_id"),
		}

		err := c.PATCH(fmt.Sprintf("/v3/cloud-rule/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update CloudRule",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Handle associations.
	if d.HasChanges("azure_arm_template_definitions",
		"azure_policy_definitions",
		"azure_role_definitions",
		"aws_cloudformation_templates",
		"compliance_standards",
		"aws_iam_policies",
		"internal_aws_amis",
		"internal_aws_service_catalog_portfolios",
		"ous",
		"projects",
		"service_control_policies",
		"gcp_iam_roles") {
		hasChanged++
		arrAddAzureArmTemplateDefinitionIds, arrRemoveAzureArmTemplateDefinitionIds, _, _ := hc.AssociationChanged(d, "azure_arm_template_definitions")
		arrAddAzurePolicyDefinitionIds, arrRemoveAzurePolicyDefinitionIds, _, _ := hc.AssociationChanged(d, "azure_policy_definitions")
		arrAddAzureRoleDefinitionIds, arrRemoveAzureRoleDefinitionIds, _, _ := hc.AssociationChanged(d, "azure_role_definitions")
		arrAddCftIds, arrRemoveCftIds, _, _ := hc.AssociationChanged(d, "aws_cloudformation_templates")
		arrAddComplianceStandardIds, arrRemoveComplianceStandardIds, _, _ := hc.AssociationChanged(d, "compliance_standards")
		arrAddIamPolicyIds, arrRemoveIamPolicyIds, _, _ := hc.AssociationChanged(d, "aws_iam_policies")
		arrAddInternalAmiIds, arrRemoveInternalAmiIds, _, _ := hc.AssociationChanged(d, "internal_aws_amis")
		arrAddInternalPortfolioIds, arrRemoveInternalPortfolioIds, _, _ := hc.AssociationChanged(d, "internal_aws_service_catalog_portfolios")
		arrAddOUIds, arrRemoveOUIds, _, _ := hc.AssociationChanged(d, "ous")
		arrAddProjectIds, arrRemoveProjectIds, _, _ := hc.AssociationChanged(d, "projects")
		arrAddServiceControlPolicyIds, arrRemoveServiceControlPolicyIds, _, _ := hc.AssociationChanged(d, "service_control_policies")
		arrAddGcpIamRoleIds, arrRemoveGcpIamRoleIds, _, _ := hc.AssociationChanged(d, "gcp_iam_roles")

		if len(arrAddAzureArmTemplateDefinitionIds) > 0 ||
			len(arrAddAzurePolicyDefinitionIds) > 0 ||
			len(arrAddAzureRoleDefinitionIds) > 0 ||
			len(arrAddCftIds) > 0 ||
			len(arrAddComplianceStandardIds) > 0 ||
			len(arrAddGcpIamRoleIds) > 0 ||
			len(arrAddIamPolicyIds) > 0 ||
			len(arrAddInternalAmiIds) > 0 ||
			len(arrAddInternalPortfolioIds) > 0 ||
			len(arrAddOUIds) > 0 ||
			len(arrAddProjectIds) > 0 ||
			len(arrAddServiceControlPolicyIds) > 0 {
			_, err := c.POST(fmt.Sprintf("/v3/cloud-rule/%s/association", ID), hc.CloudRuleAssociationsAdd{
				AzureArmTemplateDefinitionIds: &arrAddAzureArmTemplateDefinitionIds,
				AzurePolicyDefinitionIds:      &arrAddAzurePolicyDefinitionIds,
				AzureRoleDefinitionIds:        &arrAddAzureRoleDefinitionIds,
				CftIds:                        &arrAddCftIds,
				ComplianceStandardIds:         &arrAddComplianceStandardIds,
				GcpIamRoleIds:                 &arrAddGcpIamRoleIds,
				IamPolicyIds:                  &arrAddIamPolicyIds,
				InternalAmiIds:                &arrAddInternalAmiIds,
				InternalPortfolioIds:          &arrAddInternalPortfolioIds,
				OUIds:                         &arrAddOUIds,
				ProjectIds:                    &arrAddProjectIds,
				ServiceControlPolicyIds:       &arrAddServiceControlPolicyIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on CloudRule",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if len(arrRemoveAzureArmTemplateDefinitionIds) > 0 ||
			len(arrRemoveAzurePolicyDefinitionIds) > 0 ||
			len(arrRemoveAzureRoleDefinitionIds) > 0 ||
			len(arrRemoveCftIds) > 0 ||
			len(arrRemoveComplianceStandardIds) > 0 ||
			len(arrRemoveGcpIamRoleIds) > 0 ||
			len(arrRemoveIamPolicyIds) > 0 ||
			len(arrRemoveInternalAmiIds) > 0 ||
			len(arrRemoveInternalPortfolioIds) > 0 ||
			len(arrRemoveOUIds) > 0 ||
			len(arrRemoveProjectIds) > 0 ||
			len(arrRemoveServiceControlPolicyIds) > 0 {
			err := c.DELETE(fmt.Sprintf("/v3/cloud-rule/%s/association", ID), hc.CloudRuleAssociationsRemove{
				AzureArmTemplateDefinitionIds: &arrRemoveAzureArmTemplateDefinitionIds,
				AzurePolicyDefinitionIds:      &arrRemoveAzurePolicyDefinitionIds,
				AzureRoleDefinitionIds:        &arrRemoveAzureRoleDefinitionIds,
				CftIds:                        &arrRemoveCftIds,
				ComplianceStandardIds:         &arrRemoveComplianceStandardIds,
				GcpIamRoleIds:                 &arrRemoveGcpIamRoleIds,
				IamPolicyIds:                  &arrRemoveIamPolicyIds,
				InternalAmiIds:                &arrRemoveInternalAmiIds,
				InternalPortfolioIds:          &arrRemoveInternalPortfolioIds,
				OUIds:                         &arrRemoveOUIds,
				ProjectIds:                    &arrRemoveProjectIds,
				ServiceControlPolicyIds:       &arrRemoveServiceControlPolicyIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on CloudRule",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
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
			_, err := c.POST(fmt.Sprintf("/v3/cloud-rule/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on CloudRule",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 ||
			len(arrRemoveOwnerUserIds) > 0 {
			err := c.DELETE(fmt.Sprintf("/v3/cloud-rule/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrRemoveOwnerUserGroupIds,
				OwnerUserIds:      &arrRemoveOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on CloudRule",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged > 0 {
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceCloudRuleRead(ctx, d, m)
}

func resourceCloudRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	err := c.DELETE(fmt.Sprintf("/v3/cloud-rule/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete CloudRule",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
