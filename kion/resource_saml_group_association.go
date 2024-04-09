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

func resourceSamlGroupAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSamlGroupAssociationCreate,
		ReadContext:   resourceSamlGroupAssociationRead,
		UpdateContext: resourceSamlGroupAssociationUpdate,
		DeleteContext: resourceSamlGroupAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceSamlGroupAssociationRead(ctx, d, m)
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
			"assertion_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"assertion_regex": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"idms_id": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"idms_saml_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"should_update_on_login": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"update_on_login": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"user_group_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceSamlGroupAssociationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	post := hc.CreateSAMLGroupAssociation{
		AssertionName:  d.Get("assertion_name").(string),
		AssertionRegex: d.Get("assertion_regex").(string),
		IdmsID:         d.Get("idms_id").(int),
		UpdateOnLogin:  d.Get("update_on_login").(bool),
		UserGroupID:    d.Get("user_group_id").(int),
	}

	resp, err := k.POST("/v3/idms/group-association", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create SamlGroupAssociation",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create SamlGroupAssociation",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceSamlGroupAssociationRead(ctx, d, m)

	return diags
}

func resourceSamlGroupAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.GroupAssociationResponse)
	err := k.GET(fmt.Sprintf("/v3/idms/group-association/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read SamlGroupAssociation",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["assertion_name"] = item.AssertionName
	data["assertion_regex"] = item.AssertionRegex
	data["idms_id"] = item.IdmsID
	data["idms_saml_id"] = item.IdmsSamlID
	data["should_update_on_login"] = item.ShouldUpdateOnLogin
	data["user_group_id"] = item.UserGroupID

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set SamlGroupAssociation",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceSamlGroupAssociationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("assertion_name",
		"assertion_regex",
		"update_on_login",
		"user_group_id") {
		hasChanged++
		req := hc.UpdateSAMLGroupAssociation{
			AssertionName:  d.Get("assertion_name").(string),
			AssertionRegex: d.Get("assertion_regex").(string),
			UpdateOnLogin:  d.Get("update_on_login").(bool),
			UserGroupID:    d.Get("user_group_id").(int),
		}

		err := k.PATCH(fmt.Sprintf("/v3/idms/group-association/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update SamlGroupAssociation",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	if hasChanged > 0 {
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceSamlGroupAssociationRead(ctx, d, m)
}

func resourceSamlGroupAssociationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := k.DELETE(fmt.Sprintf("/v3/idms/group-association/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete SamlGroupAssociation",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
