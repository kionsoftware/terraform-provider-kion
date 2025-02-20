package kion

import (
	"context"
	"encoding/json"
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_value_string": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"default_value_list", "default_value_map"},
			},
			"default_value_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"default_value_string", "default_value_map"},
			},
			"default_value_map": {
				Type:          schema.TypeMap,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"default_value_string", "default_value_list"},
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
	client := m.(*hc.Client)

	ownerUserIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_ids").(*schema.Set).List())
	if err != nil {
		return hc.HandleError(fmt.Errorf("failed to convert owner_user_ids: %v", err))
	}
	ownerUserGroupIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_group_ids").(*schema.Set).List())
	if err != nil {
		return hc.HandleError(fmt.Errorf("failed to convert owner_user_group_ids: %v", err))
	}

	cvType := d.Get("type").(string)
	var defaultValue interface{}

	switch cvType {
	case hc.TypeString:
		defaultValue = d.Get("default_value_string")
	case hc.TypeList:
		defaultValue = d.Get("default_value_list")
	case hc.TypeMap:
		defaultValue = d.Get("default_value_map")
	default:
		return hc.HandleError(fmt.Errorf("unsupported type: %s", cvType))
	}

	if defaultValue == nil {
		return hc.HandleError(fmt.Errorf("default_value_%s must be set when type is %s", cvType, cvType))
	}

	cvValue, err := hc.UnpackCvValueJsonStr(defaultValue, cvType)
	if err != nil {
		return hc.HandleError(fmt.Errorf("failed to process default_value: %v", err))
	}

	post := hc.CustomVariableCreate{
		Name:                   d.Get("name").(string),
		Description:            d.Get("description").(string),
		Type:                   cvType,
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
		return hc.HandleError(fmt.Errorf("unable to create CustomVariable: %v", err))
	} else if resp.RecordID == 0 {
		return hc.HandleError(fmt.Errorf("received item ID of 0"))
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	return resourceCustomVariableRead(ctx, d, m)
}

func resourceCustomVariableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.CustomVariableResponse)
	err := client.GET(fmt.Sprintf("/v3/custom-variable/%s", ID), resp)
	if err != nil {
		return hc.HandleError(fmt.Errorf("unable to read CustomVariable: %v", err))
	}
	item := resp.Data

	cvType := item.Type
	cvValueStr, err := hc.PackCvValueIntoJsonStr(item.DefaultValue, cvType)
	if err != nil {
		return hc.HandleError(fmt.Errorf("failed to process default_value: %v", err))
	}

	switch cvType {
	case "string":
		diags = append(diags, hc.SafeSet(d, "default_value_string", cvValueStr, "Failed to set default_value_string")...)
	case "list":
		var list []interface{}
		if err := json.Unmarshal([]byte(cvValueStr), &list); err != nil {
			return hc.HandleError(err)
		}
		diags = append(diags, hc.SafeSet(d, "default_value_list", list, "Failed to set default_value_list")...)
	case "map":
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(cvValueStr), &m); err != nil {
			return hc.HandleError(err)
		}
		diags = append(diags, hc.SafeSet(d, "default_value_map", m, "Failed to set default_value_map")...)
	}

	fields := map[string]interface{}{
		"name":                     item.Name,
		"description":              item.Description,
		"type":                     item.Type,
		"value_validation_regex":   item.ValueValidationRegex,
		"value_validation_message": item.ValueValidationMessage,
		"key_validation_regex":     item.KeyValidationRegex,
		"key_validation_message":   item.KeyValidationMessage,
		"owner_user_ids":           item.OwnerUserIDs,
		"owner_user_group_ids":     item.OwnerUserGroupIDs,
	}

	for k, v := range fields {
		diags = append(diags, hc.SafeSet(d, k, v, fmt.Sprintf("Failed to set %s", k))...)
	}

	return diags
}

func resourceCustomVariableUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	if d.HasChanges("description", "default_value_string", "default_value_list", "default_value_map",
		"value_validation_regex", "value_validation_message", "key_validation_regex",
		"key_validation_message", "owner_user_ids", "owner_user_group_ids") {

		ownerUserIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_ids").(*schema.Set).List())
		if err != nil {
			return hc.HandleError(fmt.Errorf("failed to convert owner_user_ids: %v", err))
		}
		ownerUserGroupIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_group_ids").(*schema.Set).List())
		if err != nil {
			return hc.HandleError(fmt.Errorf("failed to convert owner_user_group_ids: %v", err))
		}

		cvType := d.Get("type").(string)
		var defaultValue interface{}

		switch cvType {
		case hc.TypeString:
			defaultValue = d.Get("default_value_string")
		case hc.TypeList:
			defaultValue = d.Get("default_value_list")
		case hc.TypeMap:
			defaultValue = d.Get("default_value_map")
		default:
			return hc.HandleError(fmt.Errorf("unsupported type: %s", cvType))
		}

		if defaultValue == nil {
			return hc.HandleError(fmt.Errorf("default_value_%s must be set when type is %s", cvType, cvType))
		}

		cvValue, err := hc.UnpackCvValueJsonStr(defaultValue, cvType)
		if err != nil {
			return hc.HandleError(fmt.Errorf("failed to process default_value: %v", err))
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
			return hc.HandleError(fmt.Errorf("unable to update CustomVariable: %v", err))
		}

		diags := hc.SafeSet(d, "last_updated", time.Now().Format(time.RFC850), "Failed to set last_updated")
		if diags.HasError() {
			return diags
		}
	}

	return resourceCustomVariableRead(ctx, d, m)
}

func resourceCustomVariableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/custom-variable/%s", ID), nil)
	if err != nil {
		return hc.HandleError(fmt.Errorf("unable to delete CustomVariable: %v", err))
	}

	d.SetId("")

	return nil
}
