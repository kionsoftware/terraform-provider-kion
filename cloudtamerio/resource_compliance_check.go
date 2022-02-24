package cloudtamerio

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

func resourceComplianceCheck() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceComplianceCheckCreate,
		ReadContext:   resourceComplianceCheckRead,
		UpdateContext: resourceComplianceCheckUpdate,
		DeleteContext: resourceComplianceCheckDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceComplianceCheckRead(ctx, d, m)
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
			"azure_policy_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"body": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloud_provider_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"compliance_check_type_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Defaults to the requesting User's ID if not specified.
			"created_by_user_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"ct_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"frequency_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"frequency_type_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3,
			},
			"is_all_regions": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"is_auto_archived": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"last_scan_id": {
				Type:     schema.TypeInt,
				Computed: true,
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
				Type:     schema.TypeList,
				Optional: true,
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
				Type:     schema.TypeList,
				Optional: true,
			},
			"regions": {
				Elem:     &schema.Schema{Type: schema.TypeString},
				Type:     schema.TypeList,
				Optional: true,
			},
			"severity_type_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3,
			},
		},
	}
}

func resourceComplianceCheckCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	post := hc.ComplianceCheckCreate{
		AzurePolicyID:         hc.FlattenIntPointer(d, "azure_policy_id"),
		Body:                  d.Get("body").(string),
		CloudProviderID:       d.Get("cloud_provider_id").(int),
		ComplianceCheckTypeID: d.Get("compliance_check_type_id").(int),
		CreatedByUserID:       d.Get("created_by_user_id").(int),
		Description:           d.Get("description").(string),
		FrequencyMinutes:      d.Get("frequency_minutes").(int),
		FrequencyTypeID:       d.Get("frequency_type_id").(int),
		IsAllRegions:          d.Get("is_all_regions").(bool),
		IsAutoArchived:        d.Get("is_auto_archived").(bool),
		Name:                  d.Get("name").(string),
		OwnerUserGroupIds:     hc.FlattenGenericIDPointer(d, "owner_user_groups"),
		OwnerUserIds:          hc.FlattenGenericIDPointer(d, "owner_users"),
		Regions:               hc.FlattenStringArray(d.Get("regions").([]interface{})),
		SeverityTypeID:        hc.FlattenIntPointer(d, "severity_type_id"),
	}

	resp, err := c.POST("/v3/compliance/check", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create ComplianceCheck",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create ComplianceCheck",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceComplianceCheckRead(ctx, d, m)

	return diags
}

func resourceComplianceCheckRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.ComplianceCheckWithOwnersResponse)
	err := c.GET(fmt.Sprintf("/v3/compliance/check/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read ComplianceCheck",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	if item.ComplianceCheck.AzurePolicyID != nil {
		data["azure_policy_id"] = item.ComplianceCheck.AzurePolicyID
	}
	data["body"] = item.ComplianceCheck.Body
	data["cloud_provider_id"] = item.ComplianceCheck.CloudProviderID
	data["compliance_check_type_id"] = item.ComplianceCheck.ComplianceCheckTypeID
	data["created_at"] = item.ComplianceCheck.CreatedAt
	data["created_by_user_id"] = item.ComplianceCheck.CreatedByUserID
	data["ct_managed"] = item.ComplianceCheck.CtManaged
	data["description"] = item.ComplianceCheck.Description
	data["frequency_minutes"] = item.ComplianceCheck.FrequencyMinutes
	data["frequency_type_id"] = item.ComplianceCheck.FrequencyTypeID
	data["is_all_regions"] = item.ComplianceCheck.IsAllRegions
	data["is_auto_archived"] = item.ComplianceCheck.IsAutoArchived
	data["last_scan_id"] = item.ComplianceCheck.LastScanID
	data["name"] = item.ComplianceCheck.Name
	if hc.InflateObjectWithID(item.OwnerUserGroups) != nil {
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
	}
	if hc.InflateObjectWithID(item.OwnerUsers) != nil {
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
	}
	data["regions"] = hc.FilterStringArray(item.ComplianceCheck.Regions)
	if item.ComplianceCheck.SeverityTypeID != nil {
		data["severity_type_id"] = item.ComplianceCheck.SeverityTypeID
	}

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set ComplianceCheck",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceComplianceCheckUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("azure_policy_id",
		"body",
		"cloud_provider_id",
		"compliance_check_type_id",
		"description",
		"frequency_minutes",
		"frequency_type_id",
		"is_all_regions",
		"is_auto_archived",
		"name",
		"regions",
		"severity_type_id") {
		hasChanged++
		req := hc.ComplianceCheckUpdate{
			AzurePolicyID:         hc.FlattenIntPointer(d, "azure_policy_id"),
			Body:                  d.Get("body").(string),
			CloudProviderID:       d.Get("cloud_provider_id").(int),
			ComplianceCheckTypeID: d.Get("compliance_check_type_id").(int),
			Description:           d.Get("description").(string),
			FrequencyMinutes:      d.Get("frequency_minutes").(int),
			FrequencyTypeID:       d.Get("frequency_type_id").(int),
			IsAllRegions:          d.Get("is_all_regions").(bool),
			IsAutoArchived:        d.Get("is_auto_archived").(bool),
			Name:                  d.Get("name").(string),
			Regions:               hc.FlattenStringArray(d.Get("regions").([]interface{})),
			SeverityTypeID:        hc.FlattenIntPointer(d, "severity_type_id"),
		}

		err := c.PATCH(fmt.Sprintf("/v3/compliance/check/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update ComplianceCheck",
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
			_, err := c.POST(fmt.Sprintf("/v3/compliance/check/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on ComplianceCheck",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 ||
			len(arrRemoveOwnerUserIds) > 0 {
			err := c.DELETE(fmt.Sprintf("/v3/compliance/check/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrRemoveOwnerUserGroupIds,
				OwnerUserIds:      &arrRemoveOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on ComplianceCheck",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged > 0 {
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceComplianceCheckRead(ctx, d, m)
}

func resourceComplianceCheckDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	err := c.DELETE(fmt.Sprintf("/v3/compliance/check/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete ComplianceCheck",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
