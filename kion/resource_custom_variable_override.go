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
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	// Get the custom variable type first
	cvID := d.Get("custom_variable_id").(string)
	cvResp := new(hc.CustomVariableResponse)
	err := client.GET(fmt.Sprintf("/v3/custom-variable/%s", cvID), cvResp)
	if err != nil {
		return diag.Errorf("failed to get custom variable type: %v", err)
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
		return diag.Errorf("unsupported type: %s", cvResp.Data.Type)
	}

	if value == nil {
		return diag.Errorf("value_%s must be set when type is %s", cvResp.Data.Type, cvResp.Data.Type)
	}

	cvValue, err := hc.UnpackCvValueJsonStr(value, cvResp.Data.Type)
	if err != nil {
		return diag.Errorf("failed to process value: %v", err)
	}

	data := hc.CustomVariableOverrideSet{
		Value: cvValue,
	}

	entityType := d.Get("entity_type").(string)
	entityID := d.Get("entity_id").(string)

	err = client.PUT(fmt.Sprintf("/v3/%s/%s/custom-variable/%s", entityType, entityID, cvID), data)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create CustomVariable Override",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), data),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", entityType, entityID, cvID))

	return resourceCustomVariableOverrideRead(ctx, d, m)
}

func resourceCustomVariableOverrideRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	entityType := d.Get("entity_type").(string)
	entityID := d.Get("entity_id").(string)
	cvID := d.Get("custom_variable_id").(string)

	// Get the custom variable type first
	cvResp := new(hc.CustomVariableResponse)
	err := client.GET(fmt.Sprintf("/v3/custom-variable/%s", cvID), cvResp)
	if err != nil {
		return diag.Errorf("failed to get custom variable type: %v", err)
	}

	resp := new(hc.CustomVariableOverrideResponse)
	err = client.GET(fmt.Sprintf("/v3/%s/%s/custom-variable/%s", entityType, entityID, cvID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read CustomVariable Override",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	// Only process if there's an override value
	if item.Override != nil && item.Override.Value != nil {
		cvValueStr, err := hc.PackCvValueIntoJsonStr(item.Override.Value, cvResp.Data.Type)
		if err != nil {
			return diag.Errorf("failed to process value: %v", err)
		}

		// Set the appropriate value based on type
		switch cvResp.Data.Type {
		case "string":
			if err := d.Set("value", cvValueStr); err != nil {
				return diag.FromErr(err)
			}
		case "list":
			var list []interface{}
			if err := json.Unmarshal([]byte(cvValueStr), &list); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("value_list", list); err != nil {
				return diag.FromErr(err)
			}
		case "map":
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(cvValueStr), &m); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("value_map", m); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	data := make(map[string]interface{})
	data["entity_type"] = entityType
	data["entity_id"] = entityID
	data["custom_variable_id"] = cvID

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set CustomVariable Override",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceCustomVariableOverrideUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := false

	// Get the custom variable type first
	cvID := d.Get("custom_variable_id").(string)
	cvResp := new(hc.CustomVariableResponse)
	err := client.GET(fmt.Sprintf("/v3/custom-variable/%s", cvID), cvResp)
	if err != nil {
		return diag.Errorf("failed to get custom variable type: %v", err)
	}

	// Check for changes in any of the value fields
	if d.HasChanges("value", "value_list", "value_map") {
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
			return diag.Errorf("unsupported type: %s", cvResp.Data.Type)
		}

		if value == nil {
			return diag.Errorf("value_%s must be set when type is %s", cvResp.Data.Type, cvResp.Data.Type)
		}

		cvValue, err := hc.UnpackCvValueJsonStr(value, cvResp.Data.Type)
		if err != nil {
			return diag.Errorf("failed to process value: %v", err)
		}

		entityType := d.Get("entity_type").(string)
		entityID := d.Get("entity_id").(string)

		req := hc.CustomVariableOverrideSet{
			Value: cvValue,
		}

		err = client.PUT(fmt.Sprintf("/v3/%s/%s/custom-variable/%s", entityType, entityID, cvID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update CustomVariable Override",
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

	return resourceCustomVariableOverrideRead(ctx, d, m)
}

func resourceCustomVariableOverrideDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	entityType := d.Get("entity_type").(string)
	entityID := d.Get("entity_id").(string)
	cvID := d.Get("custom_variable_id").(string)

	err := client.DELETE(fmt.Sprintf("/v3/%s/%s/custom-variable/%s", entityType, entityID, cvID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete CustomVariable Override",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
