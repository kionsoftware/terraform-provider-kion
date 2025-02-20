package kion

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
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
			"regions": {
				Elem:     &schema.Schema{Type: schema.TypeString},
				Type:     schema.TypeSet,
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
	client := m.(*hc.Client)

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
		Regions:               hc.FlattenStringArray(d.Get("regions").(*schema.Set).List()),
		SeverityTypeID:        hc.FlattenIntPointer(d, "severity_type_id"),
	}

	resp, err := client.POST("/v3/compliance/check", post)
	if err != nil {
		return diag.FromErr(err)
	} else if resp.RecordID == 0 {
		return diag.FromErr(fmt.Errorf("received item ID of 0"))
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	return resourceComplianceCheckRead(ctx, d, m)
}

func resourceComplianceCheckRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.ComplianceCheckWithOwnersResponse)
	if err := client.GET(fmt.Sprintf("/v3/compliance/check/%s", ID), resp); err != nil {
		return diag.FromErr(err)
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
			return diag.FromErr(fmt.Errorf("error setting %s: %w", k, err))
		}
	}

	return nil
}

func resourceComplianceCheckUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	if d.HasChanges("azure_policy_id", "body", "cloud_provider_id", "compliance_check_type_id",
		"description", "frequency_minutes", "frequency_type_id", "is_all_regions",
		"is_auto_archived", "name", "regions", "severity_type_id") {
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
			Regions:               hc.FlattenStringArray(d.Get("regions").(*schema.Set).List()),
			SeverityTypeID:        hc.FlattenIntPointer(d, "severity_type_id"),
		}

		if err := client.PATCH(fmt.Sprintf("/v3/compliance/check/%s", ID), req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle owner changes using AssociationChanged helper
	if d.HasChanges("owner_user_groups", "owner_users") {
		hasChanged++
		arrAddOwnerUserGroupIds, arrRemoveOwnerUserGroupIds, _, err := hc.AssociationChanged(d, "owner_user_groups")
		if err != nil {
			return diag.FromErr(fmt.Errorf("error determining owner user group changes: %w", err))
		}

		arrAddOwnerUserIds, arrRemoveOwnerUserIds, _, err := hc.AssociationChanged(d, "owner_users")
		if err != nil {
			return diag.FromErr(fmt.Errorf("error determining owner user changes: %w", err))
		}

		if len(arrAddOwnerUserGroupIds) > 0 || len(arrAddOwnerUserIds) > 0 {
			_, err := client.POST(fmt.Sprintf("/v3/compliance/check/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 || len(arrRemoveOwnerUserIds) > 0 {
			err := client.DELETE(fmt.Sprintf("/v3/compliance/check/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrRemoveOwnerUserGroupIds,
				OwnerUserIds:      &arrRemoveOwnerUserIds,
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if hasChanged > 0 {
		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceComplianceCheckRead(ctx, d, m)
}

func resourceComplianceCheckDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	if err := client.DELETE(fmt.Sprintf("/v3/compliance/check/%s", ID), nil); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
