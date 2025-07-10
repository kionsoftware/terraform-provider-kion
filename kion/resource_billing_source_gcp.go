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

func resourceBillingSourceGcp() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages a GCP (Google Cloud Platform) billing source in Kion.\n\n" +
			"GCP billing sources are used to import billing data from Google Cloud Platform projects " +
			"into Kion for cost management and reporting purposes. The billing data is exported from " +
			"BigQuery where Google Cloud exports billing information.",
		CreateContext: resourceBillingSourceGcpCreate,
		ReadContext:   resourceBillingSourceGcpRead,
		UpdateContext: resourceBillingSourceGcpUpdate,
		DeleteContext: resourceBillingSourceGcpDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			return validateBillingSourceGcpFields(diff)
		},
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the GCP billing source.",
			},
			"service_account_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The ID of the GCP service account used for authentication.",
			},
			"gcp_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The GCP ID of the billing account (e.g., '012345-678901-ABCDEF').",
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
			"big_query_export": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "BigQuery export configuration for billing data.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gcp_project_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the GCP project where the BigQuery dataset lives.",
						},
						"dataset_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the BigQuery dataset where the export lives.",
						},
						"table_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the BigQuery table where the export lives.",
						},
						"table_format": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "auto",
							Description:  "The format of the BigQuery table where the export lives. One of 'auto', 'standard' or 'detailed'.",
							ValidateFunc: validation.StringInSlice([]string{"auto", "standard", "detailed"}, false),
						},
						"focus_view_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name of the FOCUS view in BigQuery.",
						},
					},
				},
			},

			// Optional fields
			"account_type_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      15,
				Description:  "The account type ID for the GCP billing source. Defaults to 15 (Google Cloud).",
				ValidateFunc: validation.IntInSlice([]int{15}),
			},
			"is_reseller": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Denotes if the billing account is that of a Parent Reseller Billing Account.",
			},
			"use_focus": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Use GCP FOCUS view for billing data.",
			},
			"use_proprietary": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Use the GCP Proprietary Billing Table.",
			},
		},
	}
}

func resourceBillingSourceGcpCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Build BigQuery export configuration
	bigQueryExport := hc.GCPBigQueryExport{}
	if v, ok := d.GetOk("big_query_export"); ok {
		bqList := v.([]interface{})
		if len(bqList) > 0 {
			bq := bqList[0].(map[string]interface{})
			bigQueryExport.GCPProjectID = bq["gcp_project_id"].(string)
			bigQueryExport.DatasetName = bq["dataset_name"].(string)
			bigQueryExport.TableName = bq["table_name"].(string)
			if tableFormat, ok := bq["table_format"].(string); ok {
				bigQueryExport.TableFormat = tableFormat
			}
			if focusViewName, ok := bq["focus_view_name"].(string); ok {
				bigQueryExport.FOCUSViewName = focusViewName
			}
		}
	}

	// Build the request payload
	payload := hc.GCPBillingSourceCreate{
		AccountTypeID: uint(d.Get("account_type_id").(int)),
		GCPBillingAccountCreate: hc.GCPBillingAccountWithStart{
			ServiceAccountID: uint(d.Get("service_account_id").(int)),
			Name:             d.Get("name").(string),
			GCPID:            d.Get("gcp_id").(string),
			BillingStartDate: d.Get("billing_start_date").(string),
			BigQueryExport:   bigQueryExport,
			IsReseller:       d.Get("is_reseller").(bool),
			UseFOCUS:         d.Get("use_focus").(bool),
			UseProprietary:   d.Get("use_proprietary").(bool),
		},
	}

	// Create the billing source
	resp, err := c.POST("/v3/billing-source/gcp", payload)
	if err != nil {
		return diag.FromErr(err)
	}

	// The POST returns a Creation object with the ID
	d.SetId(strconv.Itoa(resp.RecordID))

	// Wait for the billing source to be available
	err = retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
		_, err := readGCPBillingSource(c, d.Id())
		if err != nil {
			return retry.RetryableError(fmt.Errorf("billing source not yet available: %v", err))
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	// Read the created resource
	return resourceBillingSourceGcpRead(ctx, d, m)
}

func resourceBillingSourceGcpRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	var diags diag.Diagnostics

	billingSource, err := readGCPBillingSource(c, d.Id())
	if err != nil {
		if resErr, ok := err.(*hc.RequestError); ok && resErr.StatusCode == http.StatusNotFound {
			tflog.Info(ctx, "GCP billing source not found, removing from state", map[string]interface{}{
				"id": d.Id(),
			})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Update the resource data using SafeSet helper
	diags = append(diags, hc.SafeSet(d, "name", billingSource.Name, "Unable to set name")...)
	diags = append(diags, hc.SafeSet(d, "service_account_id", billingSource.ServiceAccountID, "Unable to set service_account_id")...)
	diags = append(diags, hc.SafeSet(d, "gcp_id", billingSource.GCPID, "Unable to set gcp_id")...)

	if len(diags) > 0 {
		return diags
	}
	// BillingStartDate is not returned in GET response, so we preserve it from state
	if billingSource.BillingStartDate == "" {
		// Keep the existing value from state
		billingSource.BillingStartDate = d.Get("billing_start_date").(string)
	}
	diags = append(diags, hc.SafeSet(d, "billing_start_date", billingSource.BillingStartDate, "Unable to set billing_start_date")...)

	// Set BigQuery export configuration
	bigQueryExport := []map[string]interface{}{
		{
			"gcp_project_id":  billingSource.BigQueryExport.GCPProjectID,
			"dataset_name":    billingSource.BigQueryExport.DatasetName,
			"table_name":      billingSource.BigQueryExport.TableName,
			"table_format":    billingSource.BigQueryExport.TableFormat,
			"focus_view_name": billingSource.BigQueryExport.FOCUSViewName,
		},
	}
	diags = append(diags, hc.SafeSet(d, "big_query_export", bigQueryExport, "Unable to set big_query_export")...)

	// AccountTypeID is not returned in GET, so we preserve it from state
	// Only set it if it's not already set (e.g., during import)
	if billingSource.AccountTypeID != 0 {
		diags = append(diags, hc.SafeSet(d, "account_type_id", billingSource.AccountTypeID, "Unable to set account_type_id")...)
	}

	diags = append(diags, hc.SafeSet(d, "is_reseller", billingSource.IsReseller, "Unable to set is_reseller")...)
	diags = append(diags, hc.SafeSet(d, "use_focus", billingSource.UseFOCUSReports, "Unable to set use_focus")...)
	diags = append(diags, hc.SafeSet(d, "use_proprietary", billingSource.UseProprietaryReports, "Unable to set use_proprietary")...)

	return diags
}

func resourceBillingSourceGcpUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// GCP billing sources do not support updates via the API
	// If any changes are detected, we need to recreate the resource
	if d.HasChanges("name", "service_account_id", "gcp_id", "billing_start_date", "big_query_export",
		"account_type_id", "is_reseller", "use_focus", "use_proprietary") {

		tflog.Info(ctx, "GCP billing source does not support updates, resource must be recreated", map[string]interface{}{
			"id": d.Id(),
		})

		// Force recreation by returning an error
		return diag.Errorf("GCP billing sources cannot be updated. The resource must be recreated.")
	}

	// If no changes, just read the current state
	return resourceBillingSourceGcpRead(ctx, d, m)
}

func resourceBillingSourceGcpDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	var diags diag.Diagnostics

	// Delete the billing source
	err := c.DELETE(fmt.Sprintf("/v3/billing-source/%s", d.Id()), nil)
	if err != nil {
		if resErr, ok := err.(*hc.RequestError); ok && resErr.StatusCode == http.StatusNotFound {
			// If already deleted, we can consider this successful
			tflog.Info(ctx, "GCP billing source already deleted", map[string]interface{}{
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
			resp, err := readGCPBillingSource(c, d.Id())
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

// Helper function to read a GCP billing source
func readGCPBillingSource(c *hc.Client, id string) (*hc.GCPBillingSource, error) {
	// Get billing source by ID using the list endpoint since direct ID endpoint is inconsistent
	billingSourceID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("unable to parse billing source ID: %v", err)
	}

	listResp := new(hc.BillingSourceListResponse)
	err = c.GET("/v4/billing-source", listResp)
	if err != nil {
		return nil, err
	}

	// Find the billing source with matching ID
	var resp hc.BillingSource
	found := false
	for _, bs := range listResp.Data.Items {
		if bs.ID == uint(billingSourceID) {
			resp = bs
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("billing source %s not found", id)
	}

	// Check if this is a GCP billing source
	if resp.GCPPayer == nil {
		return nil, fmt.Errorf("billing source %s is not a GCP billing source", id)
	}

	// Convert GCPPayer to GCPBillingSource
	// Note: Some fields like AccountTypeID and BillingStartDate are not returned in the GET response
	// These fields are preserved from the Terraform state
	return &hc.GCPBillingSource{
		ID:                    resp.ID,
		Name:                  resp.GCPPayer.GCPBillingAccount.Name,
		ServiceAccountID:      resp.GCPPayer.GCPBillingAccount.ServiceAccountID,
		GCPID:                 resp.GCPPayer.GCPBillingAccount.GCPID,
		BillingStartDate:      "", // This will be preserved from state in the Read function
		BigQueryExport:        resp.GCPPayer.GCPBillingAccount.BigQueryExport,
		IsReseller:            resp.GCPPayer.GCPBillingAccount.IsReseller,
		UseFOCUSReports:       resp.UseFocusReports,
		UseProprietaryReports: resp.UseProprietaryReports,
	}, nil
}

func validateBillingSourceGcpFields(diff *schema.ResourceDiff) error {
	// Validate that focus_view_name is provided when use_focus is true
	if diff.Get("use_focus").(bool) {
		if v, ok := diff.GetOk("big_query_export"); ok {
			bqList := v.([]interface{})
			if len(bqList) > 0 {
				bq := bqList[0].(map[string]interface{})
				if focusViewName, ok := bq["focus_view_name"].(string); !ok || focusViewName == "" {
					return fmt.Errorf("focus_view_name is required in big_query_export when use_focus is true")
				}
			}
		}
	}

	return nil
}
