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

func resourceAzureArmTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAzureArmTemplateCreate,
		ReadContext:   resourceAzureArmTemplateRead,
		UpdateContext: resourceAzureArmTemplateUpdate,
		DeleteContext: resourceAzureArmTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceAzureArmTemplateRead(ctx, d, m)
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
			"ct_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"deployment_mode": {
				Type:     schema.TypeInt,
				Required: true,
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
			"resource_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"resource_group_region_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template_parameters": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // This field is defaulted to "{}" if not specified/nil on creation.
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAzureArmTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	k := m.(*hc.Client)

	post := hc.AzureARMTemplateDefinitionCreate{
		DeploymentMode:        d.Get("deployment_mode").(int),
		Description:           d.Get("description").(string),
		Name:                  d.Get("name").(string),
		OwnerUserGroupIds:     hc.FlattenGenericIDPointer(d, "owner_user_groups"),
		OwnerUserIds:          hc.FlattenGenericIDPointer(d, "owner_users"),
		ResourceGroupName:     d.Get("resource_group_name").(string),
		ResourceGroupRegionID: d.Get("resource_group_region_id").(int),
		Template:              d.Get("template").(string),
		TemplateParameters:    d.Get("template_parameters").(string),
	}

	resp, err := k.POST("/v3/azure-arm-template", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Azure ARM Template",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Azure ARM Template",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceAzureArmTemplateRead(ctx, d, m)

	return diags
}

func resourceAzureArmTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	k := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.AzureARMTemplateResponse)
	err := k.GET(fmt.Sprintf("/v3/azure-arm-template/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Azure ARM Template",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["ct_managed"] = item.AzureArmTemplate.CtManaged
	data["deployment_mode"] = item.AzureArmTemplate.DeploymentMode
	data["description"] = item.AzureArmTemplate.Description
	data["name"] = item.AzureArmTemplate.Name
	if hc.InflateObjectWithID(item.OwnerUserGroups) != nil {
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
	}
	if hc.InflateObjectWithID(item.OwnerUsers) != nil {
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
	}
	data["resource_group_name"] = item.AzureArmTemplate.ResourceGroupName
	data["resource_group_region_id"] = item.AzureArmTemplate.ResourceGroupRegionID
	data["template"] = item.AzureArmTemplate.Template
	data["template_parameters"] = item.AzureArmTemplate.TemplateParameters
	data["version"] = item.AzureArmTemplate.Version

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set Azure ARM Template",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceAzureArmTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	k := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("deployment_mode",
		"description",
		"name",
		"template",
		"template_parameters") {
		hasChanged++
		req := hc.AzureARMTemplateDefinitionUpdate{
			DeploymentMode:     d.Get("deployment_mode").(int),
			Description:        d.Get("description").(string),
			Name:               d.Get("name").(string),
			Template:           d.Get("template").(string),
			TemplateParameters: d.Get("template_parameters").(string),
		}

		err := k.PATCH(fmt.Sprintf("/v3/azure-arm-template/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update Azure ARM Template",
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
			_, err := k.POST(fmt.Sprintf("/v3/azure-arm-template/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on Azure ARM Template",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 ||
			len(arrRemoveOwnerUserIds) > 0 {
			err := k.DELETE(fmt.Sprintf("/v3/azure-arm-template/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrRemoveOwnerUserGroupIds,
				OwnerUserIds:      &arrRemoveOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on Azure ARM Template",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged > 0 {
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceAzureArmTemplateRead(ctx, d, m)
}

func resourceAzureArmTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	k := m.(*hc.Client)
	ID := d.Id()

	err := k.DELETE(fmt.Sprintf("/v3/azure-arm-template/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete Azure ARM Template",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
