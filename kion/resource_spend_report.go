package kion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceSpendReport() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSpendReportCreate,
		ReadContext:   resourceSpendReportRead,
		UpdateContext: resourceSpendReportUpdate,
		DeleteContext: resourceSpendReportDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceSpendReportRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"report_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the spend report.",
			},
			"global_visibility": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the report is globally visible.",
			},
			"date_range": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The date range for the report (e.g., 'year', 'last_six_months', 'last_three_months', 'month', 'last_month', 'custom').",
				ValidateFunc: validation.StringInSlice([]string{
					"year", "last_six_months", "last_three_months", "month", "last_month", "custom",
				}, false),
			},
			"start_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Start date for custom date range (YYYY-MM-DD format). Required when date_range is 'custom'.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
					"start_date must be in YYYY-MM-DD format",
				),
			},
			"end_date": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "End date for custom date range (YYYY-MM-DD format). Required when date_range is 'custom'.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
					"end_date must be in YYYY-MM-DD format",
				),
			},
			"scope": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "The scope of the report (e.g., 'ou', 'project', 'account', 'billingSource').",
			},
			"scope_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "The ID of the scope. Required when scope is specified.",
			},
			"spend_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The type of spend data (e.g., 'billed', 'attributed', 'unattributed').",
				ValidateFunc: validation.StringInSlice([]string{
					"billed", "attributed", "unattributed", "list", "net_attributed", "net_unattributed",
				}, false),
			},
			"dimension": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The dimension for grouping spend data (e.g., 'none', 'account', 'project', 'service', 'ou', 'billingSource', 'cloudProvider', 'cloudProviderTag', 'fundingSource', 'label', 'region', 'resource', 'usageType').",
				ValidateFunc: validation.StringInSlice([]string{
					"none", "account", "project", "service", "ou", "billingSource", "cloudProvider",
					"cloudProviderTag", "fundingSource", "label", "region", "resource", "usageType",
				}, false),
			},
			"time_granularity_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The time granularity ID (1 for monthly, 2 for daily, 3 for hourly).",
				ValidateFunc: validation.IntInSlice([]int{1, 2, 3}),
			},
			"ou_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of OU IDs to filter by.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"ou_exclusive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether OU filtering is exclusive.",
			},
			"include_descendants": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to include OU descendants.",
			},
			"project_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of project IDs to filter by.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"project_exclusive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether project filtering is exclusive.",
			},
			"billing_source_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of billing source IDs to filter by.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"billing_source_exclusive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether billing source filtering is exclusive.",
			},
			"funding_source_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of funding source IDs to filter by.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"funding_source_exclusive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether funding source filtering is exclusive.",
			},
			"cloud_provider_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of cloud provider IDs to filter by.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"cloud_provider_exclusive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether cloud provider filtering is exclusive.",
			},
			"account_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of account IDs to filter by.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"account_exclusive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether account filtering is exclusive.",
			},
			"region_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of region IDs to filter by.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"region_exclusive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether region filtering is exclusive.",
			},
			"service_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of service IDs to filter by.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"service_exclusive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether service filtering is exclusive.",
			},
			"deduct_credits": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to deduct credits from spend calculations.",
			},
			"deduct_refunds": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to deduct refunds from spend calculations.",
			},
			"scheduled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the report is scheduled.",
			},
			"scheduled_frequency": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Schedule configuration for the report.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Schedule type (0 for daily, 1 for weekly, 2 for monthly, 3 for quarterly).",
							ValidateFunc: validation.IntInSlice([]int{0, 1, 2, 3}),
						},
						"days_of_week": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Days of week for scheduling (0=Sunday, 1=Monday, etc.). Used for daily/weekly schedules.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"days_of_month": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Days of month for scheduling (1-31). Used for monthly schedules.",
							Elem:        &schema.Schema{Type: schema.TypeInt},
						},
						"hour": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Hour of day for scheduled reports (0-23).",
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"minute": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Minute of hour for scheduled reports (0-59).",
							ValidateFunc: validation.IntBetween(0, 59),
						},
						"time_zone_identifier": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Time zone identifier (e.g., 'America/Denver').",
						},
						"start_date": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Start date for the schedule (ISO 8601 format).",
						},
						"end_date": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "End date for the schedule (ISO 8601 format).",
						},
					},
				},
			},
			"scheduled_email_subject": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The email subject for scheduled reports.",
			},
			"scheduled_email_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The email message for scheduled reports.",
			},
			"scheduled_file_types": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "File types for scheduled reports (0 for CSV, 1 for Excel).",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"scheduled_file_orientation": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "File orientation for scheduled reports.",
				ValidateFunc: validation.IntInSlice([]int{0, 1, 2}),
			},
			"owner_user_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of user IDs who can view the report.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"owner_user_group_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of user group IDs who can view the report.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"external_emails": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of external email addresses to send scheduled reports to.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email_address": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Email address to send reports to.",
						},
					},
				},
			},
		},
	}
}

func resourceSpendReportCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	// Validate spend report requirements using helper function
	if validateDiags := hc.ValidateSpendReportRequirements(d); len(validateDiags) > 0 {
		return validateDiags
	}

	// Get the current user ID (we'll use 1 as default)
	createdBy := 1

	// Build the spend report object
	spendReport := hc.SpendReportCreate{
		CreatedBy:              createdBy,
		ReportName:             d.Get("report_name").(string),
		GlobalVisibility:       d.Get("global_visibility").(bool),
		DateRange:              d.Get("date_range").(string),
		Scope:                  d.Get("scope").(string),
		ScopeId:                d.Get("scope_id").(int),
		SpendType:              d.Get("spend_type").(string),
		Dimension:              d.Get("dimension").(string),
		TimeGranularityId:      d.Get("time_granularity_id").(int),
		DeductCredits:          d.Get("deduct_credits").(bool),
		DeductRefunds:          d.Get("deduct_refunds").(bool),
		Scheduled:              d.Get("scheduled").(bool),
	}

	// Set dates for custom date range
	if d.Get("date_range").(string) == "custom" {
		spendReport.StartDate = d.Get("start_date").(string)
		spendReport.EndDate = d.Get("end_date").(string)
	}

	// Set filter arrays
	spendReport.OUIds = expandIntList(d.Get("ou_ids").([]interface{}))
	spendReport.OUExclusive = d.Get("ou_exclusive").(bool)
	spendReport.IncludeDescendants = d.Get("include_descendants").(bool)
	spendReport.ProjectIds = expandIntList(d.Get("project_ids").([]interface{}))
	spendReport.ProjectExclusive = d.Get("project_exclusive").(bool)
	spendReport.BillingSourceIds = expandIntList(d.Get("billing_source_ids").([]interface{}))
	spendReport.BillingSourceExclusive = d.Get("billing_source_exclusive").(bool)
	spendReport.FundingSourceIds = expandIntList(d.Get("funding_source_ids").([]interface{}))
	spendReport.FundingSourceExclusive = d.Get("funding_source_exclusive").(bool)
	spendReport.CloudProviderIds = expandIntList(d.Get("cloud_provider_ids").([]interface{}))
	spendReport.CloudProviderExclusive = d.Get("cloud_provider_exclusive").(bool)
	spendReport.AccountIds = expandIntList(d.Get("account_ids").([]interface{}))
	spendReport.AccountExclusive = d.Get("account_exclusive").(bool)
	spendReport.ServiceIds = expandIntList(d.Get("service_ids").([]interface{}))
	spendReport.ServiceExclusive = d.Get("service_exclusive").(bool)
	spendReport.RegionIds = expandIntList(d.Get("region_ids").([]interface{}))
	spendReport.RegionExclusive = d.Get("region_exclusive").(bool)

	// Initialize empty maps for label and tag IDs
	spendReport.IncludeAppLabelIds = make(map[string]interface{})
	spendReport.ExcludeAppLabelIds = make(map[string]interface{})
	spendReport.IncludeCloudProviderTagIds = make(map[string]interface{})
	spendReport.ExcludeCloudProviderTagIds = make(map[string]interface{})

	// Set scheduled fields
	if d.Get("scheduled").(bool) {
		spendReport.ScheduledEmailSubject = d.Get("scheduled_email_subject").(string)
		spendReport.ScheduledEmailMessage = d.Get("scheduled_email_message").(string)
		spendReport.ScheduledFileTypes = expandIntList(d.Get("scheduled_file_types").([]interface{}))
		spendReport.ScheduledFileOrientation = d.Get("scheduled_file_orientation").(int)

		// Handle scheduled frequency
		if v, ok := d.GetOk("scheduled_frequency"); ok && len(v.([]interface{})) > 0 {
			freq := v.([]interface{})[0].(map[string]interface{})
			scheduledFreq := &hc.SpendReportScheduledFrequency{
				Type:               freq["type"].(int),
				Hour:               freq["hour"].(int),
				Minute:             freq["minute"].(int),
				TimeZoneIdentifier: freq["time_zone_identifier"].(string),
			}

			if days, ok := freq["days_of_week"].([]interface{}); ok {
				scheduledFreq.DaysOfWeek = expandIntList(days)
			}
			if days, ok := freq["days_of_month"].([]interface{}); ok {
				scheduledFreq.DaysOfMonth = expandIntList(days)
			}

			// Set start date
			if startDate, ok := freq["start_date"].(string); ok && startDate != "" {
				scheduledFreq.StartDate = &hc.TimeField{
					Time:  startDate,
					Valid: true,
				}
			} else {
				// Default to now
				scheduledFreq.StartDate = &hc.TimeField{
					Time:  time.Now().Format(time.RFC3339),
					Valid: true,
				}
			}

			// Set end date
			if endDate, ok := freq["end_date"].(string); ok && endDate != "" {
				scheduledFreq.EndDate = &hc.TimeField{
					Time:  endDate,
					Valid: true,
				}
			}

			spendReport.SavedReportScheduledFrequency = scheduledFreq
		}
	}

	// Build the full request
	createReq := hc.SpendReportCreateRequest{
		SavedReport: spendReport,
	}

	// Set owner IDs
	if v, ok := d.GetOk("owner_user_ids"); ok {
		createReq.UserIds = expandIntList(v.([]interface{}))
	}
	if v, ok := d.GetOk("owner_user_group_ids"); ok {
		createReq.UserGroupIds = expandIntList(v.([]interface{}))
	}

	// Set external emails
	if v, ok := d.GetOk("external_emails"); ok {
		emails := v.([]interface{})
		externalEmails := make([]hc.SpendReportExternalEmail, len(emails))
		for i, email := range emails {
			emailMap := email.(map[string]interface{})
			externalEmails[i] = hc.SpendReportExternalEmail{
				EmailAddress: emailMap["email_address"].(string),
			}
		}
		createReq.ExternalEmails = externalEmails
	}

	// Marshal request body
	requestBody, err := json.Marshal(createReq)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create request body",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	// Make the POST request to create the spend report
	url := client.HostURL + "/v1/saved-report"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create request",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	req.Header.Set("Authorization", "Bearer "+client.Token)
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := client.HTTPClient.Do(req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create spend report",
			Detail:   fmt.Sprintf("Error: %v\nNOTE: This resource uses the v1 API endpoint which is not yet stable. The public API endpoint will be available soon.", err.Error()),
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

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API request failed",
			Detail:   fmt.Sprintf("Status: %d, Body: %s\nNOTE: This resource uses the v1 API endpoint which is not yet stable. The public API endpoint will be available soon.", httpResp.StatusCode, string(body)),
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

	// Set the resource ID
	d.SetId(strconv.Itoa(resp.Data.SavedReport.ID))

	// Set last_updated
	d.Set("last_updated", time.Now().Format(time.RFC850))

	resourceSpendReportRead(ctx, d, m)

	return diags
}

func resourceSpendReportRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	spendReportID := d.Id()

	// Use the direct GET endpoint for a specific spend report
	url := client.HostURL + "/v1/saved-report/" + spendReportID
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
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
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

	// If report not found, remove from state
	if httpResp.StatusCode == 404 {
		d.SetId("")
		return diags
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API request failed",
			Detail:   fmt.Sprintf("Status: %d, Body: %s", httpResp.StatusCode, string(body)),
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

	// Set the resource data from the response
	report := resp.Data.SavedReport
	d.Set("report_name", report.ReportName)
	d.Set("global_visibility", report.GlobalVisibility)
	d.Set("date_range", report.DateRange)
	d.Set("scope", report.Scope)
	d.Set("scope_id", report.ScopeId)
	d.Set("spend_type", report.SpendType)
	d.Set("dimension", report.Dimension)
	d.Set("time_granularity_id", report.TimeGranularityId)
	d.Set("ou_ids", report.OUIds)
	d.Set("ou_exclusive", report.OUExclusive)
	d.Set("include_descendants", report.IncludeDescendants)
	d.Set("project_ids", report.ProjectIds)
	d.Set("project_exclusive", report.ProjectExclusive)
	d.Set("billing_source_ids", report.BillingSourceIds)
	d.Set("billing_source_exclusive", report.BillingSourceExclusive)
	d.Set("funding_source_ids", report.FundingSourceIds)
	d.Set("funding_source_exclusive", report.FundingSourceExclusive)
	d.Set("cloud_provider_ids", report.CloudProviderIds)
	d.Set("cloud_provider_exclusive", report.CloudProviderExclusive)
	d.Set("account_ids", report.AccountIds)
	d.Set("account_exclusive", report.AccountExclusive)
	d.Set("region_ids", report.RegionIds)
	d.Set("region_exclusive", report.RegionExclusive)
	d.Set("service_ids", report.ServiceIds)
	d.Set("service_exclusive", report.ServiceExclusive)
	d.Set("deduct_credits", report.DeductCredits)
	d.Set("deduct_refunds", report.DeductRefunds)
	d.Set("scheduled", report.Scheduled)
	d.Set("scheduled_email_subject", report.ScheduledEmailSubject)
	d.Set("scheduled_email_message", report.ScheduledEmailMessage)
	d.Set("scheduled_file_types", report.ScheduledFileTypes)
	d.Set("scheduled_file_orientation", report.ScheduledFileOrientation)
	d.Set("owner_user_ids", resp.Data.UserIds)
	d.Set("owner_user_group_ids", resp.Data.UserGroupIds)

	// Handle scheduled frequency
	if report.SavedReportScheduledFrequency != nil {
		freq := report.SavedReportScheduledFrequency
		schedFreq := []map[string]interface{}{{
			"type":                 freq.Type,
			"days_of_week":         freq.DaysOfWeek,
			"days_of_month":        freq.DaysOfMonth,
			"hour":                 freq.Hour,
			"minute":               freq.Minute,
			"time_zone_identifier": freq.TimeZoneIdentifier,
		}}

		if freq.StartDate != nil && freq.StartDate.Valid {
			schedFreq[0]["start_date"] = freq.StartDate.Time
		}
		if freq.EndDate != nil && freq.EndDate.Valid {
			schedFreq[0]["end_date"] = freq.EndDate.Time
		}

		d.Set("scheduled_frequency", schedFreq)
	}

	// Handle external emails
	if len(resp.Data.ExternalEmails) > 0 {
		externalEmails := make([]map[string]interface{}, len(resp.Data.ExternalEmails))
		for i, email := range resp.Data.ExternalEmails {
			externalEmails[i] = map[string]interface{}{
				"email_address": email.EmailAddress,
			}
		}
		d.Set("external_emails", externalEmails)
	}

	return diags
}

func resourceSpendReportUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	// Validate spend report requirements using helper function
	if validateDiags := hc.ValidateSpendReportRequirements(d); len(validateDiags) > 0 {
		return validateDiags
	}

	// First, read the current state to get all fields
	currentResp, err := getSpendReport(client, d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read current spend report",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	// Copy all values from current state
	report := currentResp.Data.SavedReport
	
	// Build the full saved report for the update (API requires all fields)
	fullReport := hc.SpendReport{
		ID:                report.ID,
		CreatedBy:         report.CreatedBy,
		ReportName:        d.Get("report_name").(string),
		GlobalVisibility:  d.Get("global_visibility").(bool),
		DateRange:         d.Get("date_range").(string),
		Scope:             d.Get("scope").(string),
		ScopeId:           d.Get("scope_id").(int),
		SpendType:         d.Get("spend_type").(string),
		Dimension:         d.Get("dimension").(string),
		TimeGranularityId: d.Get("time_granularity_id").(int),
		DeductCredits:     d.Get("deduct_credits").(bool),
		DeductRefunds:     d.Get("deduct_refunds").(bool),
		Scheduled:         d.Get("scheduled").(bool),
		CreatedAt:         report.CreatedAt,
		UpdatedAt:         report.UpdatedAt,
	}

	// Set dates for custom date range
	if d.Get("date_range").(string) == "custom" {
		fullReport.StartDate = d.Get("start_date").(string)
		fullReport.EndDate = d.Get("end_date").(string)
	}

	// Set filter arrays
	fullReport.OUIds = expandIntList(d.Get("ou_ids").([]interface{}))
	fullReport.OUExclusive = d.Get("ou_exclusive").(bool)
	fullReport.IncludeDescendants = d.Get("include_descendants").(bool)
	fullReport.ProjectIds = expandIntList(d.Get("project_ids").([]interface{}))
	fullReport.ProjectExclusive = d.Get("project_exclusive").(bool)
	fullReport.BillingSourceIds = expandIntList(d.Get("billing_source_ids").([]interface{}))
	fullReport.BillingSourceExclusive = d.Get("billing_source_exclusive").(bool)
	fullReport.FundingSourceIds = expandIntList(d.Get("funding_source_ids").([]interface{}))
	fullReport.FundingSourceExclusive = d.Get("funding_source_exclusive").(bool)
	fullReport.CloudProviderIds = expandIntList(d.Get("cloud_provider_ids").([]interface{}))
	fullReport.CloudProviderExclusive = d.Get("cloud_provider_exclusive").(bool)
	fullReport.AccountIds = expandIntList(d.Get("account_ids").([]interface{}))
	fullReport.AccountExclusive = d.Get("account_exclusive").(bool)
	fullReport.ServiceIds = expandIntList(d.Get("service_ids").([]interface{}))
	fullReport.ServiceExclusive = d.Get("service_exclusive").(bool)
	fullReport.RegionIds = expandIntList(d.Get("region_ids").([]interface{}))
	fullReport.RegionExclusive = d.Get("region_exclusive").(bool)

	// Initialize empty maps for label and tag IDs
	fullReport.IncludeAppLabelIds = make(map[string]interface{})
	fullReport.ExcludeAppLabelIds = make(map[string]interface{})
	fullReport.IncludeCloudProviderTagIds = make(map[string]interface{})
	fullReport.ExcludeCloudProviderTagIds = make(map[string]interface{})

	// Set scheduled fields
	if d.Get("scheduled").(bool) {
		fullReport.ScheduledEmailSubject = d.Get("scheduled_email_subject").(string)
		fullReport.ScheduledEmailMessage = d.Get("scheduled_email_message").(string)
		fullReport.ScheduledFileTypes = expandIntList(d.Get("scheduled_file_types").([]interface{}))
		fullReport.ScheduledFileOrientation = d.Get("scheduled_file_orientation").(int)

		// Handle scheduled frequency
		if v, ok := d.GetOk("scheduled_frequency"); ok && len(v.([]interface{})) > 0 {
			freq := v.([]interface{})[0].(map[string]interface{})
			scheduledFreq := &hc.SpendReportScheduledFrequency{
				Type:               freq["type"].(int),
				Hour:               freq["hour"].(int),
				Minute:             freq["minute"].(int),
				TimeZoneIdentifier: freq["time_zone_identifier"].(string),
			}

			if days, ok := freq["days_of_week"].([]interface{}); ok {
				scheduledFreq.DaysOfWeek = expandIntList(days)
			}
			if days, ok := freq["days_of_month"].([]interface{}); ok {
				scheduledFreq.DaysOfMonth = expandIntList(days)
			}

			// Set start date
			if startDate, ok := freq["start_date"].(string); ok && startDate != "" {
				scheduledFreq.StartDate = &hc.TimeField{
					Time:  startDate,
					Valid: true,
				}
			} else if report.SavedReportScheduledFrequency != nil && report.SavedReportScheduledFrequency.StartDate != nil {
				// Keep existing start date
				scheduledFreq.StartDate = report.SavedReportScheduledFrequency.StartDate
			} else {
				// Default to now
				scheduledFreq.StartDate = &hc.TimeField{
					Time:  time.Now().Format(time.RFC3339),
					Valid: true,
				}
			}

			// Set end date
			if endDate, ok := freq["end_date"].(string); ok && endDate != "" {
				scheduledFreq.EndDate = &hc.TimeField{
					Time:  endDate,
					Valid: true,
				}
			}

			fullReport.SavedReportScheduledFrequency = scheduledFreq
		}
	}

	// Build the request with the full report structure
	requestBody := map[string]interface{}{
		"saved_report": fullReport,
		"user_ids":     nil,
		"ugroup_ids":   nil,
		"hidden":       false,
	}

	// Set owner IDs
	if v, ok := d.GetOk("owner_user_ids"); ok {
		requestBody["user_ids"] = expandIntList(v.([]interface{}))
	}
	if v, ok := d.GetOk("owner_user_group_ids"); ok {
		requestBody["ugroup_ids"] = expandIntList(v.([]interface{}))
	}

	// Set external emails
	if v, ok := d.GetOk("external_emails"); ok {
		emails := v.([]interface{})
		externalEmails := make([]hc.SpendReportExternalEmail, len(emails))
		for i, email := range emails {
			emailMap := email.(map[string]interface{})
			externalEmails[i] = hc.SpendReportExternalEmail{
				EmailAddress: emailMap["email_address"].(string),
			}
		}
		requestBody["external_emails"] = externalEmails
	} else if len(currentResp.Data.ExternalEmails) > 0 {
		// Keep existing external emails if not specified
		requestBody["external_emails"] = currentResp.Data.ExternalEmails
	}

	// Marshal request body
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create request body",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	// Make the PUT request to update the spend report
	url := client.HostURL + "/v1/saved-report/" + d.Id()
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create request",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	req.Header.Set("Authorization", "Bearer "+client.Token)
	req.Header.Set("Content-Type", "application/json")

	httpResp, err := client.HTTPClient.Do(req)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to update spend report",
			Detail:   fmt.Sprintf("Error: %v\nNOTE: This resource uses the v1 API endpoint which is not yet stable. The public API endpoint will be available soon.", err.Error()),
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

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API request failed",
			Detail:   fmt.Sprintf("Status: %d, Body: %s\nNOTE: This resource uses the v1 API endpoint which is not yet stable. The public API endpoint will be available soon.", httpResp.StatusCode, string(body)),
		})
		return diags
	}

	// Set last_updated
	d.Set("last_updated", time.Now().Format(time.RFC850))

	resourceSpendReportRead(ctx, d, m)

	return diags
}

// Helper function to get a spend report
func getSpendReport(client *hc.Client, id string) (*hc.SpendReportResponse, error) {
	url := client.HostURL + "/v1/saved-report/" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+client.Token)

	httpResp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed - Status: %d, Body: %s", httpResp.StatusCode, string(body))
	}

	resp := new(hc.SpendReportResponse)
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func resourceSpendReportDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	spendReportID := d.Id()

	// Make the DELETE request
	url := client.HostURL + "/v1/saved-report/" + spendReportID
	req, err := http.NewRequest("DELETE", url, nil)
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
			Summary:  "Unable to delete spend report",
			Detail:   fmt.Sprintf("Error: %v\nNOTE: This resource uses the v1 API endpoint which is not yet stable. The public API endpoint will be available soon.", err.Error()),
		})
		return diags
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API request failed",
			Detail:   fmt.Sprintf("Status: %d, Body: %s", httpResp.StatusCode, string(body)),
		})
		return diags
	}

	// Remove from state
	d.SetId("")

	return diags
}

// Helper function to expand interface list to int list
// expandIntList converts a slice of interfaces to a slice of integers
// Uses the helper function ConvertInterfaceSliceToIntSlice with error handling
func expandIntList(list []interface{}) []int {
	result, err := hc.ConvertInterfaceSliceToIntSlice(list)
	if err != nil {
		// For backwards compatibility, return empty slice on error
		// This matches the original behavior where type assertion failures would panic
		return make([]int, 0)
	}
	return result
}