package kion

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceCustomVariableOverride() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCustomVariableOverrideCreate,
		ReadContext:   resourceCustomVariableOverrideRead,
		UpdateContext: resourceCustomVariableOverrideUpdate,
		DeleteContext: resourceCustomVariableOverrideDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceCustomVariableOverrideRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"value": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"value_list", "value_map"},
			},
			"value_list": {
				Type:          schema.TypeList,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"value", "value_map"},
			},
			"value_map": {
				Type:          schema.TypeMap,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"value", "value_list"},
			},
			"entity_type": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"entity_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"custom_variable_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceCustomVariableOverrideCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Get the custom variable type first
	cvID := d.Get("custom_variable_id").(string)
	cvResp := new(hc.CustomVariableResponse)
	err := client.GET(fmt.Sprintf("/v3/custom-variable/%s", cvID), cvResp)
	if err != nil {
		return hc.HandleError(fmt.Errorf("failed to get custom variable type: %v", err))
	}

	// Get the appropriate value based on type
	var value interface{}
	switch cvResp.Data.Type {
	case "string":
		value = d.Get("value")
	case "list":
		value = d.Get("value_list")
	case "map":
		value = d.Get("value_map")
	default:
		return hc.HandleError(fmt.Errorf("unsupported type: %s", cvResp.Data.Type))
	}

	if value == nil {
		return hc.HandleError(fmt.Errorf("value_%s must be set when type is %s", cvResp.Data.Type, cvResp.Data.Type))
	}

	cvValue, err := hc.UnpackCvValueJsonStr(value, cvResp.Data.Type)
	if err != nil {
		return hc.HandleError(fmt.Errorf("failed to process value: %v", err))
	}

	data := hc.CustomVariableOverrideSet{
		Value: cvValue,
	}

	entityType := d.Get("entity_type").(string)
	entityID := d.Get("entity_id").(string)

	err = client.PUT(fmt.Sprintf("/v3/%s/%s/custom-variable/%s", entityType, entityID, cvID), data)
	if err != nil {
		return hc.HandleError(fmt.Errorf("unable to create CustomVariable Override: %v", err))
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", entityType, entityID, cvID))

	return resourceCustomVariableOverrideRead(ctx, d, m)
}

func resourceCustomVariableOverrideRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	entityType := d.Get("entity_type").(string)
	entityID := d.Get("entity_id").(string)
	cvID := d.Get("custom_variable_id").(string)

	// Get the custom variable type first
	cvResp := new(hc.CustomVariableResponse)
	err := client.GET(fmt.Sprintf("/v3/custom-variable/%s", cvID), cvResp)
	if err != nil {
		return hc.HandleError(fmt.Errorf("failed to get custom variable type: %v", err))
	}

	resp := new(hc.CustomVariableOverrideResponse)
	err = client.GET(fmt.Sprintf("/v3/%s/%s/custom-variable/%s", entityType, entityID, cvID), resp)
	if err != nil {
		return hc.HandleError(fmt.Errorf("unable to read CustomVariable Override: %v", err))
	}
	item := resp.Data

	// Only process if there's an override value
	if item.Override != nil && item.Override.Value != nil {
		cvValueStr, err := hc.PackCvValueIntoJsonStr(item.Override.Value, cvResp.Data.Type)
		if err != nil {
			return hc.HandleError(fmt.Errorf("failed to process value: %v", err))
		}

		// Set the appropriate value based on type
		switch cvResp.Data.Type {
		case "string":
			diags = append(diags, hc.SafeSet(d, "value", cvValueStr, "Failed to set value")...)
		case "list":
			var list []interface{}
			if err := json.Unmarshal([]byte(cvValueStr), &list); err != nil {
				return hc.HandleError(err)
			}
			diags = append(diags, hc.SafeSet(d, "value_list", list, "Failed to set value_list")...)
		case "map":
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(cvValueStr), &m); err != nil {
				return hc.HandleError(err)
			}
			diags = append(diags, hc.SafeSet(d, "value_map", m, "Failed to set value_map")...)
		}
	}

	fields := map[string]interface{}{
		"entity_type":        entityType,
		"entity_id":          entityID,
		"custom_variable_id": cvID,
	}

	for k, v := range fields {
		diags = append(diags, hc.SafeSet(d, k, v, fmt.Sprintf("Failed to set %s", k))...)
	}

	return diags
}

func resourceCustomVariableOverrideUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	if d.HasChanges("value", "value_list", "value_map") {
		// Get the custom variable type first
		cvID := d.Get("custom_variable_id").(string)
		cvResp := new(hc.CustomVariableResponse)
		err := client.GET(fmt.Sprintf("/v3/custom-variable/%s", cvID), cvResp)
		if err != nil {
			return hc.HandleError(fmt.Errorf("failed to get custom variable type: %v", err))
		}

		// Get the appropriate value based on type
		var value interface{}
		switch cvResp.Data.Type {
		case "string":
			value = d.Get("value")
		case "list":
			value = d.Get("value_list")
		case "map":
			value = d.Get("value_map")
		default:
			return hc.HandleError(fmt.Errorf("unsupported type: %s", cvResp.Data.Type))
		}

		if value == nil {
			return hc.HandleError(fmt.Errorf("value_%s must be set when type is %s", cvResp.Data.Type, cvResp.Data.Type))
		}

		cvValue, err := hc.UnpackCvValueJsonStr(value, cvResp.Data.Type)
		if err != nil {
			return hc.HandleError(fmt.Errorf("failed to process value: %v", err))
		}

		entityType := d.Get("entity_type").(string)
		entityID := d.Get("entity_id").(string)

		req := hc.CustomVariableOverrideSet{
			Value: cvValue,
		}

		err = client.PUT(fmt.Sprintf("/v3/%s/%s/custom-variable/%s", entityType, entityID, cvID), req)
		if err != nil {
			return hc.HandleError(fmt.Errorf("unable to update CustomVariable Override: %v", err))
		}

		diags := hc.SafeSet(d, "last_updated", time.Now().Format(time.RFC850), "Failed to set last_updated")
		if diags.HasError() {
			return diags
		}
	}

	return resourceCustomVariableOverrideRead(ctx, d, m)
}

func resourceCustomVariableOverrideDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	entityType := d.Get("entity_type").(string)
	entityID := d.Get("entity_id").(string)
	cvID := d.Get("custom_variable_id").(string)

	err := client.DELETE(fmt.Sprintf("/v3/%s/%s/custom-variable/%s", entityType, entityID, cvID), nil)
	if err != nil {
		return hc.HandleError(fmt.Errorf("unable to delete CustomVariable Override: %v", err))
	}

	d.SetId("")

	return nil
}
