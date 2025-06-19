package kion

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceBillingSourceOci() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages an OCI (Oracle Cloud Infrastructure) billing source in Kion.\n\n" +
			"OCI billing sources are used to import billing data from Oracle Cloud Infrastructure tenancies " +
			"into Kion for cost management and reporting purposes.",
		CreateContext: resourceBillingSourceOciCreate,
		ReadContext:   resourceBillingSourceOciRead,
		UpdateContext: resourceBillingSourceOciUpdate,
		DeleteContext: resourceBillingSourceOciDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the OCI billing source.",
			},
			"billing_start_date": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The start date for billing data collection in YYYY-MM format.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`),
					"must be in YYYY-MM format",
				),
			},

			// Optional fields
			"account_type_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      26,
				Description:  "The account type ID for the OCI billing source. Valid values are: 26 (OCI Commercial), 27 (OCI Government), 28 (OCI Federal). Defaults to 26.",
				ValidateFunc: validation.IntInSlice([]int{26, 27, 28}),
			},
			"tenancy_ocid": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The OCID of the OCI tenancy.",
			},
			"user_ocid": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The OCID of the OCI user for API access.",
			},
			"fingerprint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The fingerprint of the API key for authentication.",
			},
			"private_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The private key for API authentication.",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The default OCI region for API access.",
			},
			"is_parent_tenancy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicates whether this billing source is a parent OCI tenancy.",
			},
			"use_focus_reports": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If true, Kion will use FOCUS reports for this billing source.",
			},
			"use_proprietary_reports": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If true, Kion will use proprietary Oracle Cost Reports for this billing source.",
			},
			"skip_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, will skip validating the connection to the billing source during creation or update.",
			},
		},
	}
}

func resourceBillingSourceOciCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	// Validate conditional requirements
	// If any API authentication field is provided, all should be provided
	apiAuthFields := []string{"user_ocid", "fingerprint", "private_key"}
	providedAuthFields := 0
	for _, field := range apiAuthFields {
		if _, ok := d.GetOk(field); ok {
			providedAuthFields++
		}
	}

	if providedAuthFields > 0 && providedAuthFields < len(apiAuthFields) {
		missingFields := []string{}
		for _, field := range apiAuthFields {
			if _, ok := d.GetOk(field); !ok {
				missingFields = append(missingFields, field)
			}
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Incomplete API authentication configuration",
			Detail:   fmt.Sprintf("When providing OCI API authentication, all of the following fields must be provided: user_ocid, fingerprint, private_key. Missing fields: %v", missingFields),
		})
		return diags
	}

	// Build the request payload
	payload := hc.OCIBillingSourceCreate{
		Name:                  d.Get("name").(string),
		BillingStartDate:      d.Get("billing_start_date").(string),
		AccountTypeID:         uint(d.Get("account_type_id").(int)),
		IsParentTenancy:       d.Get("is_parent_tenancy").(bool),
		UseFOCUSReports:       d.Get("use_focus_reports").(bool),
		UseProprietaryReports: d.Get("use_proprietary_reports").(bool),
		SkipValidation:        d.Get("skip_validation").(bool),
	}

	// Add optional fields if provided
	if v, ok := d.GetOk("tenancy_ocid"); ok {
		payload.TenancyOCID = v.(string)
	}
	if v, ok := d.GetOk("user_ocid"); ok {
		payload.UserOCID = v.(string)
	}
	if v, ok := d.GetOk("fingerprint"); ok {
		payload.Fingerprint = v.(string)
	}
	if v, ok := d.GetOk("private_key"); ok {
		payload.PrivateKey = v.(string)
	}
	if v, ok := d.GetOk("region"); ok {
		payload.Region = v.(string)
	}

	// Create the billing source
	resp, err := c.POST("/v3/billing-source/oci", payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// The POST returns a Creation object with the ID
	d.SetId(strconv.Itoa(resp.RecordID))

	// Wait for the billing source to be available
	err = retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		_, err := readOCIBillingSource(c, d.Id())
		if err != nil {
			return retry.RetryableError(fmt.Errorf("billing source not yet available: %v", err))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	// Read the created resource
	return resourceBillingSourceOciRead(ctx, d, m)
}

func resourceBillingSourceOciRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	var diags diag.Diagnostics

	billingSource, err := readOCIBillingSource(c, d.Id())
	if err != nil {
		if resErr, ok := err.(*hc.RequestError); ok && resErr.StatusCode == http.StatusNotFound {
			tflog.Info(ctx, "OCI billing source not found, removing from state", map[string]interface{}{
				"id": d.Id(),
			})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Update the resource data
	if err := d.Set("name", billingSource.Name); err != nil {
		return diag.FromErr(err)
	}
	// AccountTypeID is not returned in GET, so we preserve it from state
	// Only set it if it's not already set (e.g., during import)
	if billingSource.AccountTypeID != 0 {
		if err := d.Set("account_type_id", billingSource.AccountTypeID); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("billing_start_date", billingSource.BillingStartDate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tenancy_ocid", billingSource.TenancyOCID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user_ocid", billingSource.UserOCID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("fingerprint", billingSource.Fingerprint); err != nil {
		return diag.FromErr(err)
	}
	// Note: We don't set private_key back as it's sensitive and not returned by the API
	if err := d.Set("region", billingSource.Region); err != nil {
		return diag.FromErr(err)
	}
	// IsParentTenancy is not returned in GET, so we preserve it from state
	if billingSource.IsParentTenancy {
		if err := d.Set("is_parent_tenancy", billingSource.IsParentTenancy); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("use_focus_reports", billingSource.UseFOCUSReports); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("use_proprietary_reports", billingSource.UseProprietaryReports); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceBillingSourceOciUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	if d.HasChanges("name", "account_type_id", "billing_start_date", "tenancy_ocid", "user_ocid",
		"fingerprint", "private_key", "region", "is_parent_tenancy", "use_focus_reports", "use_proprietary_reports") {

		// Build the update payload
		payload := hc.OCIBillingSourceUpdate{
			ID:                    uint(mustAtoi(d.Id())),
			Name:                  d.Get("name").(string),
			AccountTypeID:         uint(d.Get("account_type_id").(int)),
			BillingStartDate:      d.Get("billing_start_date").(string),
			IsParentTenancy:       d.Get("is_parent_tenancy").(bool),
			UseFOCUSReports:       d.Get("use_focus_reports").(bool),
			UseProprietaryReports: d.Get("use_proprietary_reports").(bool),
			SkipValidation:        d.Get("skip_validation").(bool),
		}

		// Add optional fields if provided
		if v, ok := d.GetOk("tenancy_ocid"); ok {
			payload.TenancyOCID = v.(string)
		}
		if v, ok := d.GetOk("user_ocid"); ok {
			payload.UserOCID = v.(string)
		}
		if v, ok := d.GetOk("fingerprint"); ok {
			payload.Fingerprint = v.(string)
		}
		if v, ok := d.GetOk("private_key"); ok {
			payload.PrivateKey = v.(string)
		}
		if v, ok := d.GetOk("region"); ok {
			payload.Region = v.(string)
		}

		// Update the billing source
		err := c.PATCH(fmt.Sprintf("/v3/billing-source/oci/%s", d.Id()), payload)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Read the updated resource
	return resourceBillingSourceOciRead(ctx, d, m)
}

func resourceBillingSourceOciDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	var diags diag.Diagnostics

	// Delete the billing source
	err := c.DELETE(fmt.Sprintf("/v3/billing-source/%s", d.Id()), nil)
	if err != nil {
		if resErr, ok := err.(*hc.RequestError); ok && resErr.StatusCode == http.StatusNotFound {
			// If already deleted, we can consider this successful
			tflog.Info(ctx, "OCI billing source already deleted", map[string]interface{}{
				"id": d.Id(),
			})
		} else {
			return diag.FromErr(err)
		}
	}

	// Wait for deletion to complete
	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{"200"},
		Target:  []string{"404"},
		Refresh: func() (interface{}, string, error) {
			resp, err := readOCIBillingSource(c, d.Id())
			if err != nil {
				if resErr, ok := err.(*hc.RequestError); ok && resErr.StatusCode == http.StatusNotFound {
					return resp, "404", nil
				}
				return nil, "", err
			}
			return resp, "200", nil
		},
		Timeout:                   2 * time.Minute,
		Delay:                     10 * time.Second,
		MinTimeout:                3 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	_, err = deleteStateConf.WaitForStateContext(ctx)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return diags
}

// Helper function to read an OCI billing source
func readOCIBillingSource(c *hc.Client, id string) (*hc.OCIBillingSource, error) {
	var resp hc.BillingSource
	err := c.GET(fmt.Sprintf("/v4/billing-source/%s", id), &resp)
	if err != nil {
		return nil, err
	}

	// Check if this is an OCI billing source
	if resp.OCIPayer == nil {
		return nil, fmt.Errorf("billing source %s is not an OCI billing source", id)
	}

	// Convert OCIPayer to OCIBillingSource
	// Note: Some fields like AccountTypeID and IsParentTenancy are not returned in the GET response
	// These fields are preserved from the Terraform state
	return &hc.OCIBillingSource{
		ID:                    resp.OCIPayer.ID,
		Name:                  resp.OCIPayer.Name,
		BillingStartDate:      resp.OCIPayer.BillingStartDate,
		TenancyOCID:           resp.OCIPayer.TenancyOCID,
		UserOCID:              resp.OCIPayer.UserOCID,
		Fingerprint:           resp.OCIPayer.Fingerprint,
		Region:                resp.OCIPayer.Region,
		UseFOCUSReports:       resp.UseFocusReports,
		UseProprietaryReports: resp.UseProprietaryReports,
	}, nil
}

// Helper function to convert string to int
func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}
