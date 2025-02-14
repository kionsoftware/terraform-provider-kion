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
	client := m.(*hc.Client)

	ouID := 0
	if v := hc.OptionalInt(d, "ou_id"); v != nil {
		ouID = *v
	}

	post := hc.FundingSourceCreate{
		Amount:             d.Get("amount").(float64),
		Description:        d.Get("description").(string),
		Name:               d.Get("name").(string),
		StartDatecode:      d.Get("start_datecode").(string),
		EndDatecode:        d.Get("end_datecode").(string),
		PermissionSchemeID: d.Get("permission_scheme_id").(int),
		OUID:               ouID,
		OwnerUserIds:       hc.FlattenGenericIDPointer(d, "owner_users"),
		OwnerUserGroupIds:  hc.FlattenGenericIDPointer(d, "owner_user_groups"),
	}

	resp, err := client.POST("/v3/funding-source", post)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create Funding Source: %v", err))
	} else if resp.RecordID == 0 {
		return diag.FromErr(fmt.Errorf("received item ID of 0 when creating Funding Source"))
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	if labels, ok := d.GetOk("labels"); ok && labels != nil {
		ID := d.Id()
		err = hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "funding-source", ID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to update Funding Source labels: %v", err))
		}
	}

	return resourceFundingSourceRead(ctx, d, m)
}

func resourceFundingSourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.FundingSourceResponse)
	err := client.GET(fmt.Sprintf("/v3/funding-source/%s", ID), resp)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read Funding Source: %v", err))
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["amount"] = item.Amount
	data["description"] = item.Description
	data["name"] = item.Name
	data["ou_id"] = item.OUID
	data["start_datecode"] = item.StartDatecode
	data["end_datecode"] = item.EndDatecode

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			return diag.FromErr(fmt.Errorf("error setting %s: %v", k, err))
		}
	}

	permissionResp := new(hc.FSUserMappingListResponse)
	err = client.GET(fmt.Sprintf("/v3/funding-source/%s/permission-mapping", ID), permissionResp)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read Funding Source permissions: %v", err))
	}

	for _, permissionItem := range permissionResp.Data {
		if permissionItem.AppRoleId == 1 {
			if permissionItem.UserGroupIds != nil {
				if err := d.Set("owner_user_groups", hc.InflateArrayOfIDs(*permissionItem.UserGroupIds)); err != nil {
					return diag.FromErr(fmt.Errorf("error setting owner_user_groups: %v", err))
				}
			}
			if permissionItem.UserIds != nil {
				if err := d.Set("owner_users", hc.InflateArrayOfIDs(*permissionItem.UserIds)); err != nil {
					return diag.FromErr(fmt.Errorf("error setting owner_users: %v", err))
				}
			}
		}
	}

	// Fetch and set labels
	labelData, err := hc.ReadResourceLabels(client, "funding-source", ID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read funding source labels: %v", err))
	}

	if err := d.Set("labels", labelData); err != nil {
		return diag.FromErr(fmt.Errorf("error setting labels: %v", err))
	}

	return diag.Diagnostics{}
}

func resourceFundingSourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	// Determine if the attributes that are updatable are changed.
	if d.HasChanges("amount", "description", "end_datecode", "name", "ou_id", "start_datecode") {
		ouID := 0
		if v := hc.OptionalInt(d, "ou_id"); v != nil {
			ouID = *v
		}

		req := hc.FundingSourceUpdate{
			Amount:        d.Get("amount").(float64),
			Description:   d.Get("description").(string),
			Name:          d.Get("name").(string),
			EndDatecode:   d.Get("end_datecode").(string),
			StartDatecode: d.Get("start_datecode").(string),
			OUID:          ouID,
		}

		err := client.PATCH(fmt.Sprintf("/v3/funding-source/%s", ID), req)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to update Funding Source: %v", err))
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
				return diag.FromErr(fmt.Errorf("unable to change permission mapping on Funding Source: %v", err))
			}
		}
	}

	// Check for label changes and update accordingly
	if d.HasChanges("labels") {
		err := hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "funding-source", ID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("unable to update funding source labels: %v", err))
		}
	}

	if d.HasChanges("amount", "description", "end_datecode", "name", "ou_id", "start_datecode", "owner_users", "owner_user_groups", "labels") {
		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting last_updated: %v", err))
		}
		return resourceFundingSourceRead(ctx, d, m)
	}

	return diag.Diagnostics{}
}

func resourceFundingSourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/funding-source/%s", ID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete Funding Source: %v", err))
	}

	d.SetId("")

	return diag.Diagnostics{}
}
