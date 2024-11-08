package kion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceCustomVariable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCustomVariableCreate,
		ReadContext:   resourceCustomVariableRead,
		UpdateContext: resourceCustomVariableUpdate,
		DeleteContext: resourceCustomVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceCustomVariableRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			// Notice there is no 'id' field specified because it will be created.
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"default_value": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"default_value_list", "default_value_map"},
			},
			"default_value_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"default_value", "default_value_map"},
			},
			"default_value_map": {
				Type:          schema.TypeMap,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"default_value", "default_value_list"},
			},
			"value_validation_regex": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value_validation_message": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_validation_regex": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_validation_message": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner_user_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Optional: true,
			},
			"owner_user_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Optional: true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceCustomVariableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	ownerUserIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_ids").(*schema.Set).List())
	if err != nil {
		return diag.Errorf("failed to convert owner_user_ids: %v", err)
	}
	ownerUserGroupIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_group_ids").(*schema.Set).List())
	if err != nil {
		return diag.Errorf("failed to convert owner_user_group_ids: %v", err)
	}

	cvType := d.Get("type").(string)
	var defaultValue interface{}

	switch cvType {
	case "string":
		defaultValue = d.Get("default_value")
	case "list":
		defaultValue = d.Get("default_value_list")
	case "map":
		defaultValue = d.Get("default_value_map")
	default:
		return diag.Errorf("unsupported type: %s", cvType)
	}

	if defaultValue == nil {
		return diag.Errorf("default_value_%s must be set when type is %s", cvType, cvType)
	}

	cvValue, err := hc.UnpackCvValueJsonStr(defaultValue, cvType)
	if err != nil {
		return diag.Errorf("failed to process default_value: %v", err)
	}

	post := hc.CustomVariableCreate{
		Name:                   d.Get("name").(string),
		Description:            d.Get("description").(string),
		Type:                   d.Get("type").(string),
		DefaultValue:           cvValue,
		ValueValidationRegex:   d.Get("value_validation_regex").(string),
		ValueValidationMessage: d.Get("value_validation_message").(string),
		KeyValidationRegex:     d.Get("key_validation_regex").(string),
		KeyValidationMessage:   d.Get("key_validation_message").(string),
		OwnerUserIDs:           ownerUserIDs,
		OwnerUserGroupIDs:      ownerUserGroupIDs,
	}

	resp, err := client.POST("/v3/custom-variable", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create CustomVariable",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create CustomVariable",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceComplianceStandardRead(ctx, d, m)

	return diags
}

func resourceCustomVariableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.CustomVariableResponse)
	err := client.GET(fmt.Sprintf("/v3/custom-variable/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read CustomVariable",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	cvType := item.Type
	cvValueStr, err := hc.PackCvValueIntoJsonStr(item.DefaultValue, cvType)
	if err != nil {
		return diag.Errorf("failed to process default_value: %v", err)
	}

	switch cvType {
	case "string":
		if err := d.Set("default_value", cvValueStr); err != nil {
			return diag.FromErr(err)
		}
	case "list":
		var list []interface{}
		if err := json.Unmarshal([]byte(cvValueStr), &list); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("default_value_list", list); err != nil {
			return diag.FromErr(err)
		}
	case "map":
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(cvValueStr), &m); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("default_value_map", m); err != nil {
			return diag.FromErr(err)
		}
	}

	data := make(map[string]interface{})
	data["name"] = item.Name
	data["description"] = item.Description
	data["type"] = item.Type
	data["value_validation_regex"] = item.ValueValidationRegex
	data["value_validation_message"] = item.ValueValidationMessage
	data["key_validation_regex"] = item.KeyValidationRegex
	data["key_validation_message"] = item.KeyValidationMessage
	data["owner_user_ids"] = item.OwnerUserIDs
	data["owner_user_group_ids"] = item.OwnerUserGroupIDs

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set CustomVariable",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceCustomVariableUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := false

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("description", "default_value", "value_validation_regex", "value_validation_message", "key_validation_regex", "key_validation_message", "owner_user_ids", "owner_user_group_ids") {
		ownerUserIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_ids").(*schema.Set).List())
		if err != nil {
			return diag.Errorf("failed to convert owner_user_ids: %v", err)
		}
		ownerUserGroupIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_group_ids").(*schema.Set).List())
		if err != nil {
			return diag.Errorf("failed to convert owner_user_group_ids: %v", err)
		}

		cvType := d.Get("type").(string)
		var defaultValue interface{}

		switch cvType {
		case "string":
			defaultValue = d.Get("default_value")
		case "list":
			defaultValue = d.Get("default_value_list")
		case "map":
			defaultValue = d.Get("default_value_map")
		default:
			return diag.Errorf("unsupported type: %s", cvType)
		}

		if defaultValue == nil {
			return diag.Errorf("default_value_%s must be set when type is %s", cvType, cvType)
		}

		cvValue, err := hc.UnpackCvValueJsonStr(defaultValue, cvType)
		if err != nil {
			return diag.Errorf("failed to process default_value: %v", err)
		}

		req := hc.CustomVariableUpdate{
			Description:            d.Get("description").(string),
			DefaultValue:           cvValue,
			ValueValidationRegex:   d.Get("value_validation_regex").(string),
			ValueValidationMessage: d.Get("value_validation_message").(string),
			KeyValidationRegex:     d.Get("key_validation_regex").(string),
			KeyValidationMessage:   d.Get("key_validation_message").(string),
			OwnerUserIDs:           ownerUserIDs,
			OwnerUserGroupIDs:      ownerUserGroupIDs,
		}

		err = client.PUT(fmt.Sprintf("/v3/custom-variable/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update CustomVariable",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
		hasChanged = true
	}

	if hasChanged {
		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to set last_updated",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	return resourceCustomVariableRead(ctx, d, m)
}

func resourceCustomVariableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/custom-variable/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete CustomVariable",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
