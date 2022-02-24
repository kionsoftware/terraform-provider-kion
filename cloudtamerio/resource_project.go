package kion

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

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceProjectRead(ctx, d, m)
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
			"archived": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"auto_pay": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"default_aws_region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ou_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"owner_user_ids": {
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
			"owner_user_group_ids": {
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
			"permission_scheme_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"project_funding": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amount": {
							Type:     schema.TypeFloat,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"funding_order": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"funding_source_id": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"start_datecode": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"end_datecode": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
					},
				},
				Type:     schema.TypeList,
				Required: true,
			},
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	post := hc.ProjectCreate{
		AutoPay:            d.Get("auto_pay").(bool),
		DefaultAwsRegion:   d.Get("default_aws_region").(string),
		Description:        d.Get("description").(string),
		Name:               d.Get("name").(string),
		OUID:               d.Get("ou_id").(int),
		OwnerUserIds:       hc.FlattenGenericIDPointer(d, "owner_user_ids"),
		OwnerUserGroupIds:  hc.FlattenGenericIDPointer(d, "owner_user_group_ids"),
		PermissionSchemeID: d.Get("permission_scheme_id").(int),
	}

	// Can't cast directly to []interface{}
	// Must cast each element to map[string]interface{} & assign each value from the map to the POST object.
	post.ProjectFunding = make([]hc.ProjectFundingCreate, len(d.Get("project_funding").([]interface{})))

	for i, genericValue := range d.Get("project_funding").([]interface{}) {

		// Cast each generic interface{} value to a map of key/value pairs
		projectFundingMap := genericValue.(map[string]interface{})

		// Unpack struct values & assign them to the POST object
		post.ProjectFunding[i] = hc.ProjectFundingCreate{
			Amount:          projectFundingMap["amount"].(float64),
			FundingOrder:    projectFundingMap["funding_order"].(int),
			FundingSourceID: projectFundingMap["funding_source_id"].(int),
			StartDatecode:   projectFundingMap["start_datecode"].(string),
			EndDatecode:     projectFundingMap["end_datecode"].(string),
		}
	}

	resp, err := c.POST("/v3/project", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Project",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Project",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceProjectRead(ctx, d, m)

	return diags
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.ProjectResponse)
	err := c.GET(fmt.Sprintf("/v3/project/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Project",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["archived"] = item.Archived
	data["auto_pay"] = item.AutoPay
	data["default_aws_region"] = item.DefaultAwsRegion
	data["description"] = item.Description
	data["name"] = item.Name
	data["ou_id"] = item.OUID

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set Project",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("archived",
		"auto_pay",
		"default_aws_region",
		"description",
		"name",
		"permission_scheme_id") {
		hasChanged++
		req := hc.ProjectUpdate{
			Archived:           d.Get("archived").(bool),
			AutoPay:            d.Get("auto_pay").(bool),
			DefaultAwsRegion:   d.Get("default_aws_region").(string),
			Description:        d.Get("description").(string),
			Name:               d.Get("name").(string),
			PermissionSchemeID: d.Get("permission_scheme_id").(int),
		}

		err := c.PATCH(fmt.Sprintf("/v3/project/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update Project",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Determine if the owners have changed.
	if d.HasChanges("owner_user_ids",
		"owner_user_group_ids") {
		hasChanged++
		arrAddOwnerUserGroupIds, arrRemoveOwnerUserGroupIds, _, _ := hc.AssociationChanged(d, "owner_user_group_ids")
		arrAddOwnerUserIds, arrRemoveOwnerUserIds, _, _ := hc.AssociationChanged(d, "owner_user_ids")

		if len(arrAddOwnerUserGroupIds) > 0 ||
			len(arrAddOwnerUserIds) > 0 ||
			len(arrRemoveOwnerUserGroupIds) > 0 ||
			len(arrRemoveOwnerUserIds) > 0 {
			_, err := c.POST(fmt.Sprintf("/v1/project/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to change owners on Project",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged > 0 {
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceProjectRead(ctx, d, m)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	err := c.DELETE(fmt.Sprintf("/v3/project/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete Project",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
