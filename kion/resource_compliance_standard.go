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

func resourceComplianceStandard() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceComplianceStandardCreate,
		ReadContext:   resourceComplianceStandardRead,
		UpdateContext: resourceComplianceStandardUpdate,
		DeleteContext: resourceComplianceStandardDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceComplianceStandardRead(ctx, d, m)
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
			"compliance_checks": {
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by_user_id": {
				Type:     schema.TypeInt,
				Required: true,
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
				Description:  "Must provide at least the owner_user_groups field or the owner_users field.",
				Optional:     true,
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
				Description:  "Must provide at least the owner_user_groups field or the owner_users field.",
				Optional:     true,
				AtLeastOneOf: []string{"owner_user_groups", "owner_users"},
			},
		},
	}
}

func resourceComplianceStandardCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	post := hc.ComplianceStandardCreate{
		ComplianceCheckIds: hc.FlattenGenericIDPointer(d, "compliance_checks"),
		CreatedByUserID:    d.Get("created_by_user_id").(int),
		Description:        d.Get("description").(string),
		Name:               d.Get("name").(string),
		OwnerUserGroupIds:  hc.FlattenGenericIDPointer(d, "owner_user_groups"),
		OwnerUserIds:       hc.FlattenGenericIDPointer(d, "owner_users"),
	}

	resp, err := client.POST("/v3/compliance/standard", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create ComplianceStandard",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create ComplianceStandard",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceComplianceStandardRead(ctx, d, m)

	return diags
}

func resourceComplianceStandardRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.ComplianceStandardResponse)
	err := client.GET(fmt.Sprintf("/v3/compliance/standard/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read ComplianceStandard",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	if hc.InflateObjectWithID(item.ComplianceChecks) != nil {
		data["compliance_checks"] = hc.InflateObjectWithID(item.ComplianceChecks)
	}
	data["created_at"] = item.ComplianceStandard.CreatedAt
	data["created_by_user_id"] = item.ComplianceStandard.CreatedByUserID
	data["ct_managed"] = item.ComplianceStandard.CtManaged
	data["description"] = item.ComplianceStandard.Description
	data["name"] = item.ComplianceStandard.Name
	if hc.InflateObjectWithID(item.OwnerUserGroups) != nil {
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
	}
	if hc.InflateObjectWithID(item.OwnerUsers) != nil {
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
	}

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set ComplianceStandard",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceComplianceStandardUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	if d.HasChanges("description", "name") {
		hasChanged++
		req := hc.ComplianceStandardUpdate{
			Description: d.Get("description").(string),
			Name:        d.Get("name").(string),
		}

		if err := client.PATCH(fmt.Sprintf("/v3/compliance/standard/%s", ID), req); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle associations.
	if d.HasChanges("compliance_checks") {
		hasChanged++

		// Get current state before making changes
		resp := new(hc.ComplianceStandardResponse)
		if err := client.GET(fmt.Sprintf("/v3/compliance/standard/%s", ID), resp); err != nil {
			return diag.FromErr(err)
		}

		// Create map of current compliance checks
		currentChecks := make(map[int]bool)
		for _, check := range resp.Data.ComplianceChecks {
			currentChecks[check.ID] = true
		}

		// Use AssociationChanged helper to get changes
		arrAddComplianceCheckIds, arrRemoveComplianceCheckIds, changed, err := hc.AssociationChanged(d, "compliance_checks")
		if err != nil {
			return diag.FromErr(fmt.Errorf("error determining compliance check changes: %w", err))
		}

		// Filter out checks that don't exist from removal list
		var validRemoveChecks []int
		for _, checkID := range arrRemoveComplianceCheckIds {
			if currentChecks[checkID] {
				validRemoveChecks = append(validRemoveChecks, checkID)
			}
		}

		if len(arrAddComplianceCheckIds) > 0 {
			_, err := client.POST(fmt.Sprintf("/v3/compliance/standard/%s/association", ID), hc.ComplianceStandardAssociationsAdd{
				ComplianceCheckIds: &arrAddComplianceCheckIds,
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if len(validRemoveChecks) > 0 {
			err := client.DELETE(fmt.Sprintf("/v3/compliance/standard/%s/association", ID), hc.ComplianceStandardAssociationsRemove{
				ComplianceCheckIds: validRemoveChecks,
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if changed {
			hasChanged++
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
			_, err := client.POST(fmt.Sprintf("/v3/compliance/standard/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				return diag.FromErr(err)
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 || len(arrRemoveOwnerUserIds) > 0 {
			err := client.DELETE(fmt.Sprintf("/v3/compliance/standard/%s/owner", ID), hc.ChangeOwners{
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

	return resourceComplianceStandardRead(ctx, d, m)
}

func resourceComplianceStandardDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/compliance/standard/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete ComplianceStandard",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
