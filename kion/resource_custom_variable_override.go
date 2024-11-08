package kion

import (
	"context"
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
				Type:     schema.TypeString,
				Required: true,
			},
			// All following fields are in place of an ID since overrides do not have an ID.
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

	cvValue, err := hc.UnpackCvValueJsonStr(d.Get("value").(string))
	if err != nil {
		return diag.Errorf(err.Error())
	}

	data := hc.CustomVariableOverrideSet{
		Value: cvValue,
	}

	entityType := d.Get("entity_type").(string)
	entityID := d.Get("entity_id").(string)
	cvID := d.Get("custom_variable_id").(string)

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

	resourceCustomVariableOverrideRead(ctx, d, m)

	return diags
}

func resourceCustomVariableOverrideRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	entityType := d.Get("entity_type").(string)
	entityID := d.Get("entity_id").(string)
	cvID := d.Get("custom_variable_id").(string)

	resp := new(hc.CustomVariableOverrideResponse)
	err := client.GET(fmt.Sprintf("/v3/%s/%s/custom-variable/%s", entityType, entityID, cvID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read CustomVariable Override",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	cvValueStr, err := hc.PackCvValueIntoJsonStr(item.Value)
	if err != nil {
		return diag.Errorf(err.Error())
	}

	data := make(map[string]interface{})
	data["value"] = cvValueStr
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

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("value") {

		cvValue, err := hc.UnpackCvValueJsonStr(d.Get("value").(string))
		if err != nil {
			return diag.Errorf(err.Error())
		}

		entityType := d.Get("entity_type").(string)
		entityID := d.Get("entity_id").(string)
		cvID := d.Get("custom_variable_id").(string)

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
