package kion

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceLabel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLabelCreate,
		ReadContext:   resourceLabelRead,
		UpdateContext: resourceLabelUpdate,
		DeleteContext: resourceLabelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceLabelRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			// Notice there is no 'id' field specified because it will be created.
			"color": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The color of the label in hex format, (#123abc)",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^#[0-9a-fA-F]{6}$"),
					"must be a valid hex color code with leading #",
				),
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceLabelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	k := m.(*hc.Client)

	post := hc.LabelCreate{
		Color: d.Get("color").(string),
		Key:   d.Get("key").(string),
		Value: d.Get("value").(string),
	}

	resp, err := k.POST("/v3/label", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Label",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Label",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceLabelRead(ctx, d, m)

	return diags
}

func resourceLabelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	k := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.LabelResponse)
	err := k.GET(fmt.Sprintf("/v3/label/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Label",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	label := resp.Data

	d.Set("key", label.Key)
	d.Set("value", label.Value)
	d.Set("color", label.Color)

	return diags
}

func resourceLabelUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	k := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	if d.HasChanges("color",
		"key",
		"value") {
		hasChanged++
		req := hc.LabelUpdatable{
			Color: d.Get("color").(string),
			Key:   d.Get("key").(string),
			Value: d.Get("value").(string),
		}

		err := k.PATCH(fmt.Sprintf("/v3/label/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update Label",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return resourceLabelRead(ctx, d, m)
}

func resourceLabelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	k := m.(*hc.Client)
	ID := d.Id()

	err := k.DELETE(fmt.Sprintf("/v3/label/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete label",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
