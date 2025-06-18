package kion

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceBillingSourceAws() *schema.Resource {
	return &schema.Resource{
		Description: "Creates and manages an AWS commercial billing source.\n\n" +
			"AWS billing sources enable cost management and account management capabilities " +
			"by connecting Kion to AWS billing data. This resource creates commercial AWS " +
			"billing sources (account type 1). For GovCloud billing sources, use the " +
			"`kion_billing_source_aws_govcloud` resource.",
		CreateContext: resourceBillingSourceAwsCreate,
		ReadContext:   resourceBillingSourceAwsRead,
		UpdateContext: resourceBillingSourceAwsUpdate,
		DeleteContext: resourceBillingSourceAwsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the billing source.",
			},
			"aws_account_number": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The AWS account number of the master billing account.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\d{12}$`),
					"must be a 12-digit AWS account number",
				),
			},
			"billing_start_date": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The start date for billing data collection in YYYY-MM format.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\d{4}-(?:0[1-9]|1[0-2])$`),
					"must be in YYYY-MM format",
				),
			},
			
			// Optional fields
			"account_creation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, Kion is able to automatically create accounts in this billing source.",
			},
			"billing_bucket_account_number": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The AWS account number of the S3 bucket holding the billing reports. Defaults to aws_account_number if not specified.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\d{12}$`),
					"must be a 12-digit AWS account number",
				),
			},
			"billing_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region of the S3 bucket holding billing reports (both CUR and DBR reports).",
			},
			"billing_report_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "cur",
				Description: "The billing report type to use. Options: 'none' (no proprietary billing report), 'cur' (AWS Cost and Usage Report), 'dbrrt' (AWS Detailed Billing Report with Resources and Tags).",
				ValidateFunc: validation.StringInSlice([]string{"none", "cur", "dbrrt"}, false),
			},
			"bucket_access_role": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An alternate IAM role for accessing the billing buckets (optional).",
			},
			
			// CUR-specific fields
			"cur_bucket": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the S3 bucket containing the Cost and Usage Reports. Required if billing_report_type is 'cur'.",
			},
			"cur_bucket_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region of the S3 bucket containing the Cost and Usage Reports. Required if billing_report_type is 'cur'.",
			},
			"cur_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the Cost and Usage Report. Required if billing_report_type is 'cur'.",
			},
			"cur_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The report prefix for the Cost and Usage Reports. Required if billing_report_type is 'cur'.",
			},
			
			// FOCUS billing fields
			"focus_billing_bucket_account_number": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The AWS account number of the S3 bucket holding the FOCUS reports.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\d{12}$`),
					"must be a 12-digit AWS account number",
				),
			},
			"focus_billing_report_bucket": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the S3 bucket containing the FOCUS reports.",
			},
			"focus_billing_report_bucket_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region of the S3 bucket containing the FOCUS reports.",
			},
			"focus_billing_report_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the FOCUS billing report.",
			},
			"focus_billing_report_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The prefix for the FOCUS billing reports.",
			},
			"focus_bucket_access_role": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An alternate IAM role for accessing the FOCUS billing buckets (optional).",
			},
			
			// Authentication fields
			"key_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS Access Key ID used to access the billing S3 bucket.",
			},
			"key_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The AWS Secret Access Key used to access the billing S3 bucket.",
			},
			
			// Other optional fields
			"linked_role": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "OrganizationAccountAccessRole",
				Description: "The name of an existing IAM role that has full administrator permissions. This role will be prefilled as the linked role when creating or importing new accounts under this billing source.",
			},
			"mr_bucket": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the S3 bucket containing the monthly reports (detailed billing reports). Required if billing_report_type is 'dbrrt'.",
			},
			"skip_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When true, will skip validating the connection to the billing source during creation.",
			},
			
			// Computed fields
			"use_focus_reports": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if billing source is configured to read FOCUS reports.",
			},
			"use_proprietary_reports": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if billing source is configured to read proprietary billing reports from AWS (CUR, DBRRT).",
			},
		},
	}
}

func resourceBillingSourceAwsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	
	// Validate conditional requirements
	billingReportType := d.Get("billing_report_type").(string)
	if billingReportType == "cur" {
		// All CUR-specific required fields
		requiredCurFields := map[string]string{
			"cur_bucket": "The S3 bucket containing the Cost and Usage Reports",
			"cur_bucket_region": "The region of the S3 bucket containing the Cost and Usage Reports",
			"cur_name": "The name of the Cost and Usage Report",
			"cur_prefix": "The report prefix for the Cost and Usage Reports",
		}
		
		for field, description := range requiredCurFields {
			if _, ok := d.GetOk(field); !ok {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Missing required field",
					Detail:   fmt.Sprintf("%s is required when billing_report_type is 'cur'. %s.", field, description),
				})
			}
		}
		if len(diags) > 0 {
			return diags
		}
	}
	
	if billingReportType == "focus" {
		// All FOCUS-specific required fields
		requiredFocusFields := map[string]string{
			"focus_billing_report_bucket": "The S3 bucket containing the FOCUS reports",
			"focus_billing_report_name": "The name of the FOCUS billing report",
		}
		
		for field, description := range requiredFocusFields {
			if _, ok := d.GetOk(field); !ok {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Missing required field",
					Detail:   fmt.Sprintf("%s is required when billing_report_type is 'focus'. %s.", field, description),
				})
			}
		}
		if len(diags) > 0 {
			return diags
		}
	}
	
	if billingReportType == "dbrrt" {
		// DBR (Detailed Billing Report) required field
		if _, ok := d.GetOk("s3_bucket_name"); !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Missing required field",
				Detail:   "s3_bucket_name is required when billing_report_type is 'dbrrt'. The S3 bucket containing the monthly detailed billing reports.",
			})
			return diags
		}
	}
	
	// Build the create request
	post := hc.AWSBillingSourceCreate{
		Name:             d.Get("name").(string),
		AWSAccountNumber: d.Get("aws_account_number").(string),
		AccountTypeID:    1, // AWS Commercial
		AccountCreation:  d.Get("account_creation").(bool),
		BillingStartDate: d.Get("billing_start_date").(string),
		LinkedRole:       d.Get("linked_role").(string),
		SkipValidation:   d.Get("skip_validation").(bool),
	}
	
	// Set billing bucket account number (defaults to aws_account_number if not specified)
	if v, ok := d.GetOk("billing_bucket_account_number"); ok {
		post.BillingBucketAccountNumber = v.(string)
	} else {
		post.BillingBucketAccountNumber = d.Get("aws_account_number").(string)
	}
	
	// Set optional fields using helper functions
	if region := hc.FlattenStringPointer(d, "billing_region"); region != nil {
		post.BillingRegion = *region
	}
	if reportType := hc.FlattenStringPointer(d, "billing_report_type"); reportType != nil {
		post.BillingReportType = *reportType
	}
	if bucketRole := hc.FlattenStringPointer(d, "bucket_access_role"); bucketRole != nil {
		post.BucketAccessRole = *bucketRole
	}
	
	// CUR-specific fields using helper functions
	if curBucket := hc.FlattenStringPointer(d, "cur_bucket"); curBucket != nil {
		post.CURBucket = *curBucket
	}
	if curRegion := hc.FlattenStringPointer(d, "cur_bucket_region"); curRegion != nil {
		post.CURBucketRegion = *curRegion
	}
	if curName := hc.FlattenStringPointer(d, "cur_name"); curName != nil {
		post.CURName = *curName
	}
	if curPrefix := hc.FlattenStringPointer(d, "cur_prefix"); curPrefix != nil {
		post.CURPrefix = *curPrefix
	}
	
	// FOCUS billing fields
	if v, ok := d.GetOk("focus_billing_bucket_account_number"); ok {
		post.FocusBillingBucketAccountNumber = v.(string)
	}
	if v, ok := d.GetOk("focus_billing_report_bucket"); ok {
		post.FocusBillingReportBucket = v.(string)
	}
	if v, ok := d.GetOk("focus_billing_report_bucket_region"); ok {
		post.FocusBillingReportBucketRegion = v.(string)
	}
	if v, ok := d.GetOk("focus_billing_report_name"); ok {
		post.FocusBillingReportName = v.(string)
	}
	if v, ok := d.GetOk("focus_billing_report_prefix"); ok {
		post.FocusBillingReportPrefix = v.(string)
	}
	if v, ok := d.GetOk("focus_bucket_access_role"); ok {
		post.FocusBucketAccessRole = v.(string)
	}
	
	// Authentication fields
	if v, ok := d.GetOk("key_id"); ok {
		post.KeyID = v.(string)
	}
	if v, ok := d.GetOk("key_secret"); ok {
		post.KeySecret = v.(string)
	}
	
	// Other optional fields
	if v, ok := d.GetOk("mr_bucket"); ok {
		post.MRBucket = v.(string)
	}
	
	resp, err := client.POST("/v3/billing-source/aws", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create AWS billing source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	}
	
	d.SetId(fmt.Sprintf("%d", resp.RecordID))
	
	resourceBillingSourceAwsRead(ctx, d, m)
	
	return diags
}

func resourceBillingSourceAwsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	
	ID := d.Id()
	billingSourceID, err := strconv.Atoi(ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse billing source ID",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}
	
	// Get billing source by account number to find the ID
	// Since there's no direct GET by ID endpoint, we need to use the account number
	accountNumber := d.Get("aws_account_number").(string)
	if accountNumber == "" {
		// During import, we don't have the account number, so we need to list all billing sources
		// and find the one with the matching ID
		resp := new(hc.BillingSourceListResponse)
		err := client.GET("/v4/billing-source", resp)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to list billing sources",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
			return diags
		}
		
		// Find the billing source with matching ID
		found := false
		for _, bs := range resp.Data.Items {
			if bs.ID == uint(billingSourceID) && bs.AWSPayer != nil {
				// AWSPayer is a pointer to interface{}, so we need to dereference and type assert it
				if awsPayerInterface := *bs.AWSPayer; awsPayerInterface != nil {
					if awsPayerMap, ok := awsPayerInterface.(map[string]interface{}); ok {
						if accountNum, ok := awsPayerMap["account_number"].(string); ok {
							accountNumber = accountNum
							found = true
							break
						}
					}
				}
			}
		}
		
		if !found {
			d.SetId("")
			return diags
		}
	}
	
	// Get billing source by account number
	resp := new(hc.BillingSourceResponse)
	err = client.GET(fmt.Sprintf("/v4/billing-source/by-child-account-number/%s", accountNumber), resp)
	if err != nil {
		// If not found, clear the ID
		if reqErr, ok := err.(*hc.RequestError); ok && reqErr.StatusCode == 404 {
			d.SetId("")
			return diags
		}
		
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read AWS billing source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	
	billingSource := resp.Data
	
	// Verify this is an AWS billing source
	if billingSource.AWSPayer == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Billing source is not an AWS billing source",
			Detail:   fmt.Sprintf("Billing source ID %v is not configured as an AWS billing source", ID),
		})
		return diags
	}
	
	// Set the resource data from the response
	awsPayer := billingSource.AWSPayer
	
	// Set basic fields using SafeSet helper
	diags = append(diags, hc.SafeSet(d, "name", awsPayer.Name, "Unable to set name")...)
	diags = append(diags, hc.SafeSet(d, "aws_account_number", awsPayer.AccountNumber, "Unable to set aws_account_number")...)
	diags = append(diags, hc.SafeSet(d, "account_creation", billingSource.AccountCreation, "Unable to set account_creation")...)
	
	// Set other fields from AWSPayer using SafeSet helper
	if awsPayer.BillingBucketAccountNumber != "" {
		diags = append(diags, hc.SafeSet(d, "billing_bucket_account_number", awsPayer.BillingBucketAccountNumber, "Unable to set billing_bucket_account_number")...)
	}
	
	if awsPayer.BillingRegion != "" {
		diags = append(diags, hc.SafeSet(d, "billing_region", awsPayer.BillingRegion, "Unable to set billing_region")...)
	}
	
	if awsPayer.BillingReportType != "" {
		if err := d.Set("billing_report_type", awsPayer.BillingReportType); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set billing_report_type",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.BillingStartDate != "" {
		if err := d.Set("billing_start_date", awsPayer.BillingStartDate); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set billing_start_date",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.BucketAccessRole != "" {
		if err := d.Set("bucket_access_role", awsPayer.BucketAccessRole); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set bucket_access_role",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	// CUR fields
	if awsPayer.BillingReportBucket != "" {
		if err := d.Set("cur_bucket", awsPayer.BillingReportBucket); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set cur_bucket",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.BillingReportBucketRegion != "" {
		if err := d.Set("cur_bucket_region", awsPayer.BillingReportBucketRegion); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set cur_bucket_region",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.BillingReportName != "" {
		if err := d.Set("cur_name", awsPayer.BillingReportName); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set cur_name",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.BillingReportPrefix != "" {
		if err := d.Set("cur_prefix", awsPayer.BillingReportPrefix); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set cur_prefix",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	// FOCUS fields
	if awsPayer.FOCUSBillingBucketAccountNumber != "" {
		if err := d.Set("focus_billing_bucket_account_number", awsPayer.FOCUSBillingBucketAccountNumber); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set focus_billing_bucket_account_number",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.FOCUSBillingReportBucket != "" {
		if err := d.Set("focus_billing_report_bucket", awsPayer.FOCUSBillingReportBucket); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set focus_billing_report_bucket",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.FOCUSBillingReportBucketRegion != "" {
		if err := d.Set("focus_billing_report_bucket_region", awsPayer.FOCUSBillingReportBucketRegion); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set focus_billing_report_bucket_region",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.FOCUSBillingReportName != "" {
		if err := d.Set("focus_billing_report_name", awsPayer.FOCUSBillingReportName); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set focus_billing_report_name",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.FOCUSBillingReportPrefix != "" {
		if err := d.Set("focus_billing_report_prefix", awsPayer.FOCUSBillingReportPrefix); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set focus_billing_report_prefix",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	if awsPayer.FOCUSBucketAccessRole != "" {
		if err := d.Set("focus_bucket_access_role", awsPayer.FOCUSBucketAccessRole); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set focus_bucket_access_role",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	// Other fields
	if awsPayer.DetailedBillingBucket != "" {
		if err := d.Set("mr_bucket", awsPayer.DetailedBillingBucket); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set mr_bucket",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
		}
	}
	
	// Computed fields
	if err := d.Set("use_focus_reports", billingSource.UseFocusReports); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set use_focus_reports",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
	}
	
	if err := d.Set("use_proprietary_reports", billingSource.UseProprietaryReports); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set use_proprietary_reports",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
	}
	
	// Ensure the ID is set to the billing source ID
	d.SetId(fmt.Sprintf("%d", billingSource.ID))
	
	return diags
}

func resourceBillingSourceAwsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	
	// Check what has changed and build the update request
	hasChanged := false
	update := hc.AWSBillingSourceUpdate{}
	
	if d.HasChange("name") {
		hasChanged = true
		update.Name = d.Get("name").(string)
	}
	
	if d.HasChange("account_creation") {
		hasChanged = true
		accountCreation := d.Get("account_creation").(bool)
		update.AccountCreation = &accountCreation
	}
	
	if d.HasChange("billing_bucket_account_number") {
		hasChanged = true
		update.BillingBucketAccountNumber = d.Get("billing_bucket_account_number").(string)
	}
	
	if d.HasChange("billing_region") {
		hasChanged = true
		update.BillingRegion = d.Get("billing_region").(string)
	}
	
	if d.HasChange("billing_report_type") {
		hasChanged = true
		update.BillingReportType = d.Get("billing_report_type").(string)
	}
	
	if d.HasChange("billing_start_date") {
		hasChanged = true
		update.BillingStartDate = d.Get("billing_start_date").(string)
	}
	
	if d.HasChange("bucket_access_role") {
		hasChanged = true
		update.BucketAccessRole = d.Get("bucket_access_role").(string)
	}
	
	// CUR fields
	if d.HasChange("cur_bucket") {
		hasChanged = true
		update.CURBucket = d.Get("cur_bucket").(string)
	}
	
	if d.HasChange("cur_bucket_region") {
		hasChanged = true
		update.CURBucketRegion = d.Get("cur_bucket_region").(string)
	}
	
	if d.HasChange("cur_name") {
		hasChanged = true
		update.CURName = d.Get("cur_name").(string)
	}
	
	if d.HasChange("cur_prefix") {
		hasChanged = true
		update.CURPrefix = d.Get("cur_prefix").(string)
	}
	
	// FOCUS fields
	if d.HasChange("focus_billing_bucket_account_number") {
		hasChanged = true
		update.FocusBillingBucketAccountNumber = d.Get("focus_billing_bucket_account_number").(string)
	}
	
	if d.HasChange("focus_billing_report_bucket") {
		hasChanged = true
		update.FocusBillingReportBucket = d.Get("focus_billing_report_bucket").(string)
	}
	
	if d.HasChange("focus_billing_report_bucket_region") {
		hasChanged = true
		update.FocusBillingReportBucketRegion = d.Get("focus_billing_report_bucket_region").(string)
	}
	
	if d.HasChange("focus_billing_report_name") {
		hasChanged = true
		update.FocusBillingReportName = d.Get("focus_billing_report_name").(string)
	}
	
	if d.HasChange("focus_billing_report_prefix") {
		hasChanged = true
		update.FocusBillingReportPrefix = d.Get("focus_billing_report_prefix").(string)
	}
	
	if d.HasChange("focus_bucket_access_role") {
		hasChanged = true
		update.FocusBucketAccessRole = d.Get("focus_bucket_access_role").(string)
	}
	
	// Authentication fields
	if d.HasChange("key_id") {
		hasChanged = true
		update.KeyID = d.Get("key_id").(string)
	}
	
	if d.HasChange("key_secret") {
		hasChanged = true
		update.KeySecret = d.Get("key_secret").(string)
	}
	
	// Other fields
	if d.HasChange("linked_role") {
		hasChanged = true
		update.LinkedRole = d.Get("linked_role").(string)
	}
	
	if d.HasChange("mr_bucket") {
		hasChanged = true
		update.MRBucket = d.Get("mr_bucket").(string)
	}
	
	if hasChanged {
		// Since there's no direct update endpoint for billing sources,
		// we'll need to use a PATCH endpoint if available, or recreate
		// For now, we'll return an error indicating updates are not supported
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "AWS billing source updates are not supported",
			Detail:   "The Kion API does not provide an update endpoint for AWS billing sources. Please destroy and recreate the resource to make changes.",
		})
		return diags
	}
	
	return resourceBillingSourceAwsRead(ctx, d, m)
}

func resourceBillingSourceAwsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	
	ID := d.Id()
	billingSourceID, err := strconv.Atoi(ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse billing source ID",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}
	
	err = client.DELETE(fmt.Sprintf("/v3/billing-source/%d", billingSourceID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete AWS billing source",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	
	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	
	return diags
}