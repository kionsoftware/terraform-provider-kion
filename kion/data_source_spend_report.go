package kion

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceSpendReport() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSpendReportRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The ID of the spend report to retrieve.",
			},
			"spend_reports": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The unique ID of the spend report.",
						},
						"created_by": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the user who created the report.",
						},
						"report_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the spend report.",
						},
						"global_visibility": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the report is globally visible.",
						},
						"date_range": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The date range for the report (e.g., 'year', 'last_six_months').",
						},
						"scope": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The scope of the report.",
						},
						"scope_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the scope.",
						},
						"spend_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of spend data (e.g., 'billed', 'attributed').",
						},
						"dimension": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The dimension for grouping spend data (e.g., 'account', 'project', 'service').",
						},
						"time_granularity_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The time granularity ID (1 for monthly, 2 for daily, 3 for hourly).",
						},
						"ou_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of OU IDs to filter by.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"ou_exclusive": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether OU filtering is exclusive.",
						},
						"include_descendants": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to include OU descendants.",
						},
						"project_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of project IDs to filter by.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"project_exclusive": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether project filtering is exclusive.",
						},
						"billing_source_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of billing source IDs to filter by.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"billing_source_exclusive": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether billing source filtering is exclusive.",
						},
						"funding_source_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of funding source IDs to filter by.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"funding_source_exclusive": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether funding source filtering is exclusive.",
						},
						"cloud_provider_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of cloud provider IDs to filter by.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"cloud_provider_exclusive": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether cloud provider filtering is exclusive.",
						},
						"account_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of account IDs to filter by.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"account_exclusive": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether account filtering is exclusive.",
						},
						"region_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of region IDs to filter by.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"region_exclusive": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether region filtering is exclusive.",
						},
						"service_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of service IDs to filter by.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"service_exclusive": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether service filtering is exclusive.",
						},
						"deduct_credits": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to deduct credits from spend calculations.",
						},
						"deduct_refunds": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether to deduct refunds from spend calculations.",
						},
						"scheduled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the report is scheduled.",
						},
						"scheduled_email_subject": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The email subject for scheduled reports.",
						},
						"scheduled_email_message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The email message for scheduled reports.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The creation timestamp of the report.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The last update timestamp of the report.",
						},
					},
				},
			},
		},
	}
}

func dataSourceSpendReportRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	spendReportID := d.Get("id").(int)
	
	// Use the direct GET endpoint for a specific spend report
	url := client.HostURL + "/v1/saved-report/" + strconv.Itoa(spendReportID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create request",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	req.Header.Set("Authorization", "Bearer "+client.Token)

	httpResp, err := client.HTTPClient.Do(req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read spend report",
			Detail:   fmt.Sprintf("Error: %v\nNOTE: This data source uses the v1 API endpoint which is not yet stable. The public API endpoint will be available soon.", err.Error()),
		})
		return diags
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read response body",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	if httpResp.StatusCode == 404 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Spend report not found",
			Detail:   fmt.Sprintf("Spend report with ID %d not found", spendReportID),
		})
		return diags
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API request failed",
			Detail:   fmt.Sprintf("Status: %d, Body: %s\nNOTE: This data source uses the v1 API endpoint which is not yet stable. The public API endpoint will be available soon.", httpResp.StatusCode, string(body)),
		})
		return diags
	}

	resp := new(hc.SpendReportResponse)
	err = json.Unmarshal(body, resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse response",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	// Convert single response to list format
	spendReports := []map[string]interface{}{{
		"id":                       resp.Data.SavedReport.ID,
		"created_by":               resp.Data.SavedReport.CreatedBy,
		"report_name":              resp.Data.SavedReport.ReportName,
		"global_visibility":        resp.Data.SavedReport.GlobalVisibility,
		"date_range":               resp.Data.SavedReport.DateRange,
		"scope":                    resp.Data.SavedReport.Scope,
		"scope_id":                 resp.Data.SavedReport.ScopeId,
		"spend_type":               resp.Data.SavedReport.SpendType,
		"dimension":                resp.Data.SavedReport.Dimension,
		"time_granularity_id":      resp.Data.SavedReport.TimeGranularityId,
		"ou_ids":                   resp.Data.SavedReport.OUIds,
		"ou_exclusive":             resp.Data.SavedReport.OUExclusive,
		"include_descendants":      resp.Data.SavedReport.IncludeDescendants,
		"project_ids":              resp.Data.SavedReport.ProjectIds,
		"project_exclusive":        resp.Data.SavedReport.ProjectExclusive,
		"billing_source_ids":       resp.Data.SavedReport.BillingSourceIds,
		"billing_source_exclusive": resp.Data.SavedReport.BillingSourceExclusive,
		"funding_source_ids":       resp.Data.SavedReport.FundingSourceIds,
		"funding_source_exclusive": resp.Data.SavedReport.FundingSourceExclusive,
		"cloud_provider_ids":       resp.Data.SavedReport.CloudProviderIds,
		"cloud_provider_exclusive": resp.Data.SavedReport.CloudProviderExclusive,
		"account_ids":              resp.Data.SavedReport.AccountIds,
		"account_exclusive":        resp.Data.SavedReport.AccountExclusive,
		"region_ids":               resp.Data.SavedReport.RegionIds,
		"region_exclusive":         resp.Data.SavedReport.RegionExclusive,
		"service_ids":              resp.Data.SavedReport.ServiceIds,
		"service_exclusive":        resp.Data.SavedReport.ServiceExclusive,
		"deduct_credits":           resp.Data.SavedReport.DeductCredits,
		"deduct_refunds":           resp.Data.SavedReport.DeductRefunds,
		"scheduled":                resp.Data.SavedReport.Scheduled,
		"scheduled_email_subject":  resp.Data.SavedReport.ScheduledEmailSubject,
		"scheduled_email_message":  resp.Data.SavedReport.ScheduledEmailMessage,
	}}

	// Handle time fields
	if resp.Data.SavedReport.CreatedAt != nil && resp.Data.SavedReport.CreatedAt.Valid {
		spendReports[0]["created_at"] = resp.Data.SavedReport.CreatedAt.Time
	}
	if resp.Data.SavedReport.UpdatedAt != nil && resp.Data.SavedReport.UpdatedAt.Valid {
		spendReports[0]["updated_at"] = resp.Data.SavedReport.UpdatedAt.Time
	}

	if err := d.Set("spend_reports", spendReports); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set spend_reports",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	// Set ID with timestamp for consistency with other data sources
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}