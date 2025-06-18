package kion

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceBillingSourceAwsGovcloud() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages an AWS GovCloud billing source.\n\n" +
			"AWS GovCloud billing sources are attached to commercial AWS billing sources and enable " +
			"GovCloud account management capabilities. The parent commercial billing source must exist " +
			"before creating the GovCloud billing source.",
		CreateContext: resourceBillingSourceAwsGovcloudCreate,
		ReadContext:   resourceBillingSourceAwsGovcloudRead,
		UpdateContext: resourceBillingSourceAwsGovcloudUpdate,
		DeleteContext: resourceBillingSourceAwsGovcloudDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				// The import ID should be in the format: commercial_billing_source_id
				// We'll need to read the GovCloud info from that billing source
				commercialBillingSourceID := d.Id()
				d.Set("commercial_billing_source_id", commercialBillingSourceID)
				
				// Read the GovCloud billing source info
				resourceBillingSourceAwsGovcloudRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"commercial_billing_source_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the commercial AWS billing source to attach this GovCloud billing source to.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the GovCloud billing source.",
			},
			"aws_account_number": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The AWS account number for the GovCloud payer account.",
			},
			"account_creation_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether account creation is enabled in this GovCloud billing source.",
			},
			"car_external_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external ID used when assuming the cloud access role for this billing source.",
			},
			"service_external_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external ID used for automated internal actions using the service role for this billing source.",
			},
			"govcloud_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the GovCloud billing source record.",
			},
		},
	}
}

func resourceBillingSourceAwsGovcloudCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	
	commercialBillingSourceID := d.Get("commercial_billing_source_id").(int)
	
	post := hc.BillingSourceGovcloudCreate{
		AccountCreationEnabled: d.Get("account_creation_enabled").(bool),
		AWSAccountNumber:       d.Get("aws_account_number").(string),
		Name:                   d.Get("name").(string),
	}
	
	_, err := client.POST(fmt.Sprintf("/v3/billing-source/%d/govcloud", commercialBillingSourceID), post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create AWS GovCloud billing source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	}
	
	// The API returns a 201 Created response, but we need to read the resource
	// to get the full details including the GovCloud ID
	d.SetId(strconv.Itoa(commercialBillingSourceID))
	
	// Set the commercial billing source ID so the read function can use it
	d.Set("commercial_billing_source_id", commercialBillingSourceID)
	
	resourceBillingSourceAwsGovcloudRead(ctx, d, m)
	
	return diags
}

func resourceBillingSourceAwsGovcloudRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	
	commercialBillingSourceID := d.Get("commercial_billing_source_id").(int)
	if commercialBillingSourceID == 0 {
		// Try to parse from ID if not set (for imports)
		commercialBillingSourceID, _ = strconv.Atoi(d.Id())
	}
	
	resp := new(hc.BillingSourceGovcloudResponse)
	err := client.GET(fmt.Sprintf("/v4/billing-source/%d/govcloud", commercialBillingSourceID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read AWS GovCloud billing source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), commercialBillingSourceID),
		})
		return diags
	}
	
	govcloud := resp.Data
	
	// Set all the fields from the response using SafeSet helper
	diags = append(diags, hc.SafeSet(d, "name", govcloud.Name, "Unable to set name")...)
	diags = append(diags, hc.SafeSet(d, "aws_account_number", govcloud.AWSAccountNumber, "Unable to set aws_account_number")...)
	diags = append(diags, hc.SafeSet(d, "account_creation_enabled", govcloud.AccountCreationEnabled, "Unable to set account_creation_enabled")...)
	diags = append(diags, hc.SafeSet(d, "car_external_id", govcloud.CARExternalID, "Unable to set car_external_id")...)
	diags = append(diags, hc.SafeSet(d, "service_external_id", govcloud.ServiceExternalID, "Unable to set service_external_id")...)
	
	diags = append(diags, hc.SafeSet(d, "govcloud_id", govcloud.ID, "Unable to set govcloud_id")...)
	
	// Keep the commercial billing source ID as the resource ID
	d.SetId(strconv.Itoa(commercialBillingSourceID))
	
	return diags
}

func resourceBillingSourceAwsGovcloudUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	
	commercialBillingSourceID := d.Get("commercial_billing_source_id").(int)
	
	hasChanged := false
	update := hc.BillingSourceGovcloudUpdate{}
	
	if d.HasChange("name") {
		hasChanged = true
		update.Name = d.Get("name").(string)
	}
	
	if d.HasChange("account_creation_enabled") {
		hasChanged = true
		enabled := d.Get("account_creation_enabled").(bool)
		update.AccountCreationEnabled = &enabled
	}
	
	if hasChanged {
		err := client.PATCH(fmt.Sprintf("/v3/billing-source/%d/govcloud", commercialBillingSourceID), update)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update AWS GovCloud billing source",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), commercialBillingSourceID),
			})
			return diags
		}
		
		// Read the resource to ensure we have the latest state
		resourceBillingSourceAwsGovcloudRead(ctx, d, m)
	}
	
	return diags
}

func resourceBillingSourceAwsGovcloudDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	
	commercialBillingSourceID := d.Get("commercial_billing_source_id").(int)
	
	err := client.DELETE(fmt.Sprintf("/v3/billing-source/%d/govcloud", commercialBillingSourceID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete AWS GovCloud billing source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), commercialBillingSourceID),
		})
		return diags
	}
	
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	
	return diags
}