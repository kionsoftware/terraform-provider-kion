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

func resourceFundingSource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFundingSourceCreate,
		ReadContext:   resourceFundingSourceRead,
		UpdateContext: resourceFundingSourceUpdate,
		DeleteContext: resourceFundingSourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceFundingSourceRead(ctx, d, m)
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
			"amount": {
				Type:     schema.TypeFloat,
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
			"start_datecode": {
				Type:     schema.TypeString,
				Required: true,
			},
			"end_datecode": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ou_id": {
				Type:     schema.TypeInt,
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
				Type:         schema.TypeSet,
				Optional:     true,
				Description:  "Must provide at least the owner_user_groups field or the owner_users field.",
				AtLeastOneOf: []string{"owner_user_groups", "owner_users"},
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
			"permission_scheme_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A map of labels to assign to the funding source. The labels must already exist in Kion.",
			},
		},
	}
}

func resourceFundingSourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	post := hc.FundingSourceCreate{
		Amount:             d.Get("amount").(float64),
		Description:        d.Get("description").(string),
		Name:               d.Get("name").(string),
		StartDatecode:      d.Get("start_datecode").(string),
		EndDatecode:        d.Get("end_datecode").(string),
		PermissionSchemeID: d.Get("permission_scheme_id").(int),
		OUID:               d.Get("ou_id").(int),
		OwnerUserIds:       hc.FlattenGenericIDPointer(d, "owner_users"),
		OwnerUserGroupIds:  hc.FlattenGenericIDPointer(d, "owner_user_groups"),
	}

	resp, err := client.POST("/v3/funding-source", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Funding Source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Funding Source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	if labels, ok := d.GetOk("labels"); ok && labels != nil {
		ID := d.Id()
		err = hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "funding-source", ID)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update Funding Source labels",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	resourceFundingSourceRead(ctx, d, m)

	return diags
}

func resourceFundingSourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.FundingSourceResponse)
	err := client.GET(fmt.Sprintf("/v3/funding-source/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Funding Source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["amount"] = item.Amount
	data["description"] = item.Description
	data["name"] = item.Name
	data["ou_id"] = item.OUID
	data["start_datecode"] = item.StartDatecode
	data["end_datecode"] = item.EndDatecode

	permissionResp := new(hc.FSUserMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/funding-source/%s/permission-mapping", ID), permissionResp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Funding Source permissions",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	for _, permissionItem := range permissionResp.Data {
		if permissionItem.AppRoleId == 1 {
			if permissionItem.UserGroupIds != nil {
				data["owner_user_groups"] = hc.InflateArrayOfIDs(*permissionItem.UserGroupIds)
			}
			if permissionItem.UserIds != nil {
				data["owner_users"] = hc.InflateArrayOfIDs(*permissionItem.UserIds)
			}
		}
	}

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set Funding Source",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Fetch labels
	labelData, err := hc.ReadResourceLabels(client, "funding-source", ID)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read funding source labels",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// Set labels
	err = d.Set("labels", labelData)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set labels for funding source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
	}

	return diags
}

func resourceFundingSourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	// Determine if the attributes that are updatable are changed.
	if d.HasChanges("amount", "description", "end_datecode", "name", "ou_id", "start_datecode") {
		req := hc.FundingSourceUpdate{
			Amount:        d.Get("amount").(float64),
			Description:   d.Get("description").(string),
			Name:          d.Get("name").(string),
			EndDatecode:   d.Get("end_datecode").(string),
			StartDatecode: d.Get("start_datecode").(string),
		}

		err := client.PATCH(fmt.Sprintf("/v3/funding-source/%s", ID), req)
		if err != nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Unable to update Funding Source",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				},
			}
		}
	}

	// Determine if the owners have changed.
	if d.HasChanges("owner_users", "owner_user_groups") {
		arrAddOwnerUserGroupIds, arrRemoveOwnerUserGroupIds, _, _ := hc.AssociationChanged(d, "owner_user_groups")
		arrAddOwnerUserIds, arrRemoveOwnerUserIds, _, _ := hc.AssociationChanged(d, "owner_users")

		if len(arrAddOwnerUserGroupIds) > 0 || len(arrAddOwnerUserIds) > 0 || len(arrRemoveOwnerUserGroupIds) > 0 || len(arrRemoveOwnerUserIds) > 0 {
			patch := []hc.FundingSourcePermissionMapping{
				{
					AppRoleID:    1,
					UserGroupIds: hc.FlattenGenericIDPointer(d, "owner_user_groups"),
					UserIds:      hc.FlattenGenericIDPointer(d, "owner_users"),
				},
			}

			err := client.PATCH(fmt.Sprintf("/v3/funding-source/%s/permission-mapping", ID), patch)
			if err != nil {
				return diag.Diagnostics{
					{
						Severity: diag.Error,
						Summary:  "Unable to change permission mapping on Funding Source",
						Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
					},
				}
			}
		}
	}

	// Check for label changes and update accordingly
	if d.HasChanges("labels") {
		err := hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "funding-source", ID)
		if err != nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Unable to update funding source labels",
					Detail:   fmt.Sprintf("Error: %v\nFunding source ID: %v", err.Error(), ID),
				},
			}
		}
	}

	if d.HasChanges("amount", "description", "end_datecode", "name", "ou_id", "start_datecode", "owner_users", "owner_user_groups", "labels") {
		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to set last_updated",
				Detail:   err.Error(),
			})
			return diags
		}
		return resourceFundingSourceRead(ctx, d, m)
	}

	return diags
}

func resourceFundingSourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/funding-source/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete Funding Source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
