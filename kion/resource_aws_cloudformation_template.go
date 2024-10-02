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

func resourceAwsCloudformationTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsCloudformationTemplateCreate,
		ReadContext:   resourceAwsCloudformationTemplateRead,
		UpdateContext: resourceAwsCloudformationTemplateUpdate,
		DeleteContext: resourceAwsCloudformationTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceAwsCloudformationTemplateRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			// Notice there is no 'id' field specified because it will be created.
			"last_updated": {
				Type:     schema.TypeString,
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
			"policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"regions": {
				Elem:     &schema.Schema{Type: schema.TypeString},
				Type:     schema.TypeSet,
				Required: true,
			},
			"sns_arns": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"template_parameters": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"termination_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tags": {
				Elem:        &schema.Schema{Type: schema.TypeString},
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Stack-level tags will apply to all supported resources in a CloudFormation stack.  Requires Kion >= 3.7.1.",
			},
		},
	}
}

func resourceAwsCloudformationTemplateCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	post := hc.CFTCreate{
		Description:           d.Get("description").(string),
		Name:                  d.Get("name").(string),
		OwnerUserGroupIds:     hc.FlattenGenericIDPointer(d, "owner_user_groups"),
		OwnerUserIds:          hc.FlattenGenericIDPointer(d, "owner_users"),
		Policy:                d.Get("policy").(string),
		Region:                d.Get("region").(string),
		Regions:               hc.FlattenStringArray(d.Get("regions").(*schema.Set).List()),
		SnsArns:               d.Get("sns_arns").(string),
		Tags:                  hc.FlattenTags(d, "tags"),
		TemplateParameters:    d.Get("template_parameters").(string),
		TerminationProtection: d.Get("termination_protection").(bool),
	}

	resp, err := client.POST("/v3/cft", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create AwsCloudformationTemplate",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create AwsCloudformationTemplate",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceAwsCloudformationTemplateRead(ctx, d, m)

	return diags
}

func resourceAwsCloudformationTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.CFTResponseWithOwnersAndTags)
	err := client.GET(fmt.Sprintf("/v3/cft/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read AwsCloudformationTemplate",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["description"] = item.Cft.Description
	data["name"] = item.Cft.Name
	if hc.InflateObjectWithID(item.OwnerUserGroups) != nil {
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
	}
	if hc.InflateObjectWithID(item.OwnerUsers) != nil {
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
	}
	data["policy"] = item.Cft.Policy
	data["region"] = item.Cft.Region
	data["regions"] = hc.FilterStringArray(item.Cft.Regions)
	data["sns_arns"] = item.Cft.SnsArns
	data["tags"] = hc.InflateTags(item.Tags)
	data["template_parameters"] = item.Cft.TemplateParameters
	data["termination_protection"] = item.Cft.TerminationProtection

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set AwsCloudformationTemplate",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceAwsCloudformationTemplateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		"policy",
		"region",
		"regions",
		"sns_arns",
		"tags",
		"template_parameters",
		"termination_protection") {
		hasChanged++
		req := hc.CFTUpdate{
			Description:           d.Get("description").(string),
			Name:                  d.Get("name").(string),
			Policy:                d.Get("policy").(string),
			Region:                d.Get("region").(string),
			Regions:               hc.FlattenStringArray(d.Get("regions").(*schema.Set).List()),
			SnsArns:               d.Get("sns_arns").(string),
			Tags:                  hc.FlattenTags(d, "tags"),
			TemplateParameters:    d.Get("template_parameters").(string),
			TerminationProtection: d.Get("termination_protection").(bool),
		}

		err := client.PATCH(fmt.Sprintf("/v3/cft/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update AwsCloudformationTemplate",
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
			_, err := client.POST(fmt.Sprintf("/v3/cft/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on AwsCloudformationTemplate",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 ||
			len(arrRemoveOwnerUserIds) > 0 {
			err := client.DELETE(fmt.Sprintf("/v3/cft/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrRemoveOwnerUserGroupIds,
				OwnerUserIds:      &arrRemoveOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on AwsCloudformationTemplate",
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

	return resourceAwsCloudformationTemplateRead(ctx, d, m)
}

func resourceAwsCloudformationTemplateDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/cft/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete AwsCloudformationTemplate",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
