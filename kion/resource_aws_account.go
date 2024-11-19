package kion

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceAwsAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Creates or imports an AWS Account and adds it to a Kion project or the Kion account cache.\n\n" +
			"If `account_number` is provided, an existing account will be imported into Kion, otherwise " +
			"a new AWS account will be created.  If `project_id` is provided the account will be added " +
			"to the corresponding project, otherwise the account will be added to the account cache.\n\n" +
			"Once added, an account can be moved between projects or in and out of the account cache by " +
			"changing the `project_id`.  When moving accounts between projects, use `move_project_settings` " +
			"to control how financials will be treated between the old and new project.\n\n" +
			"When importing existing Kion accounts into terraform state, you can use one of these methods:\n\n" +
			"1. Default import (tries project first, then cache):\n" +
			"    terraform import kion_aws_account.example 123\n\n" +
			"2. Explicit project account import:\n" +
			"    terraform import kion_aws_account.example account_id=123\n\n" +
			"3. Explicit cache account import:\n" +
			"    terraform import kion_aws_account.example account_cache_id=123\n\n" +
			"**NOTE:** This resource requires Kion v3.8.4 or greater.",
		CreateContext: resourceAwsAccountCreate,
		ReadContext:   resourceAwsAccountRead,
		UpdateContext: resourceAwsAccountUpdate,
		DeleteContext: resourceAwsAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceAwsAccountRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			// Notice there is no 'id' field specified because it will be created.
			"account_alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Account alias is an optional short unique name that helps identify the account within Kion.",
			},
			"account_number": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The account number of the AWS account.  If account_number is provided, the existing account will be imported into Kion.  If account_number is omitted, a new account will be created.",
				ForceNew:    true,
			},
			"account_type_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "An ID representing the account type within Kion.",
			},
			"aws_organizational_unit": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "Where to place this account within AWS Organization when creating an account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the organizational unit in AWS.",
						},
						"org_unit_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "OUID of the AWS Organization unit.",
						},
					},
				},
			},
			"car_external_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external ID used when assuming cloud access roles.",
			},
			"commercial_account_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name used when creating new commercial account.",
			},
			"create_govcloud": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "True to create an AWS GovCloud account.",
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The root email address to associate with a new account.  Required when creating a new account unless an account placeholder email has been set.",
			},
			"gov_account_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name used when creating new GovCloud account.",
			},
			"include_linked_account_spend": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "True to associate spend from a linked GovCloud account with this account.",
			},
			"labels": {
				Type:         schema.TypeMap,
				Optional:     true,
				RequiredWith: []string{"project_id"},
				Elem:         &schema.Schema{Type: schema.TypeString},
				Description:  "A map of labels to assign to the account. The labels must already exist in Kion.",
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"linked_account_number": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "For AWS GovCloud accounts, this is the linked commercial account.  Otherwise this is empty.",
			},
			"linked_role": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "OrganizationAccountAccessRole",
				Description: "The AWS organization service role.",
			},
			"location": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Where the account is attached. Either \"project\" or \"cache\".",
			},
			"move_project_settings": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "Parameters used when moving an account between Kion projects.  These settings are ignored unless moving an account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"financials": {
							Type:         schema.TypeString,
							Default:      "move",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"preserve", "move"}, false),
							Description:  "One of \"move\" or \"preserve\".  If \"move\", financial history will be moved to the new project beginning on the date specified by the move_datecode parameter.  If \"preserve\", financial history will be preserved on the current project.",
						},
						"move_datecode": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The start date to use when moving financial data in YYYYMM format.  This only applies when financials is set to move.  If provided, only financial data from this date to the current month will be moved to the new project.  If omitted or 0, all financial data will be moved to the new project.",
						},
					},
				},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the AWS account within Kion.",
			},
			"payer_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The ID of the billing source containing billing data for this account.",
			},
			"project_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The ID of the Kion project to place this account within. If empty, the account will be placed within the account cache.",
			},
			"service_external_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external ID used for automated internal actions using the service role for this account.",
			},
			"skip_access_checking": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "True to skip periodic access checking on the account.",
			},
			"start_datecode": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Date when the AWS account will starting submitting payments against a funding source (YYYY-MM).  Required if placing an account within a project.",
			},
			"use_org_account_info": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "True to keep the account name and email address in Kion in sync with the account name and email address as set in AWS Organization.",
			},
		},
		CustomizeDiff: customdiff.All(
			// schema validators don't support multi-attribute validations, so we use CustomizeDiff instead
			validateAwsAccountStartDatecode,
			customDiffComputedAccountLocation,
		),
	}
}

func resourceAwsAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	// Set initial location - this should match what customDiffComputedAccountLocation does
	var accountLocation string
	if projectId := d.Get("project_id").(int); projectId != 0 {
		accountLocation = ProjectLocation
	} else {
		accountLocation = CacheLocation
	}

	// Set location before any API calls
	diags = append(diags, hc.SafeSet(d, "location", accountLocation, "Failed to set initial location")...)
	if diags.HasError() {
		return diags
	}

	if _, ok := d.GetOk("account_number"); ok {
		// Import an existing AWS account
		accountTypeId := int(hc.AWSStandard)
		if v, ok := d.GetOk("account_type_id"); ok {
			accountTypeId = v.(int)
		}

		var postAccountData interface{}
		var accountUrl string
		switch accountLocation {
		case CacheLocation:
			accountUrl = "/v3/account-cache?account-type=aws"
			postAccountData = hc.AccountCacheNewAWSImport{
				AccountAlias:              hc.OptionalValue[string](d, "account_alias"),
				AccountEmail:              d.Get("email").(string),
				AccountNumber:             d.Get("account_number").(string),
				AccountTypeID:             &accountTypeId,
				IncludeLinkedAccountSpend: hc.OptionalValue[bool](d, "include_linked_account_spend"),
				LinkedAccountNumber:       d.Get("linked_account_number").(string),
				LinkedRole:                d.Get("linked_role").(string),
				Name:                      d.Get("name").(string),
				PayerID:                   d.Get("payer_id").(int),
				SkipAccessChecking:        hc.OptionalValue[bool](d, "skip_access_checking"),
			}

		case ProjectLocation:
			fallthrough
		default:
			accountUrl = "/v3/account?account-type=aws"
			postAccountData = hc.AccountNewAWSImport{
				AccountAlias:              hc.OptionalValue[string](d, "account_alias"),
				AccountEmail:              d.Get("email").(string),
				AccountNumber:             d.Get("account_number").(string),
				AccountTypeID:             &accountTypeId,
				IncludeLinkedAccountSpend: hc.OptionalValue[bool](d, "include_linked_account_spend"),
				LinkedAccountNumber:       d.Get("linked_account_number").(string),
				LinkedRole:                d.Get("linked_role").(string),
				Name:                      d.Get("name").(string),
				PayerID:                   d.Get("payer_id").(int),
				ProjectID:                 d.Get("project_id").(int),
				SkipAccessChecking:        hc.OptionalValue[bool](d, "skip_access_checking"),
				StartDatecode:             d.Get("start_datecode").(string),
				UseOrgAccountInfo:         hc.OptionalValue[bool](d, "use_org_account_info"),
			}
		}

		// Log the request data
		if rb, err := json.Marshal(postAccountData); err == nil {
			tflog.Debug(ctx, fmt.Sprintf("Importing existing AWS account via POST %s", accountUrl), map[string]interface{}{
				"postData": string(rb),
				"url":      accountUrl,
			})
		}

		resp, err := client.POST(accountUrl, postAccountData)
		if err != nil {
			diags = append(diags, hc.HandleError(fmt.Errorf("unable to import AWS Account: %v", err))...)
			return diags
		}

		d.SetId(fmt.Sprintf("%d", resp.RecordID))

	} else {
		// Create new AWS account
		diags, accountCacheId := createAwsAccount(ctx, client, d)
		if diags.HasError() {
			return diags
		}

		switch accountLocation {
		case ProjectLocation:
			// Move cached account to the requested project
			projectId := d.Get("project_id").(int)
			startDatecode := time.Now().Format("200601")
			retries := 3
			delay := 30 * time.Second

			newId, err := retryConvertCacheAccountToProjectAccountForAWS(client, accountCacheId, projectId, startDatecode, retries, delay)
			if err != nil {
				diags = append(diags, hc.HandleError(fmt.Errorf("unable to convert AWS cached account to project account: %v", err))...)
				return diags
			}

			d.SetId(fmt.Sprintf("%d", newId))

		case CacheLocation:
			// Track the cached account
			d.SetId(fmt.Sprintf("%d", accountCacheId))
		}
	}

	// Labels are only supported on project accounts, not cached accounts
	if accountLocation == ProjectLocation {
		if _, ok := d.GetOk("labels"); ok {
			ID := d.Id()
			if err := hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "account", ID); err != nil {
				return append(diags, hc.HandleError(fmt.Errorf("unable to update AWS account labels (ID: %s): %v", ID, err))...)
			}
		}
	}

	return append(diags, resourceAwsAccountRead(ctx, d, m)...)
}

func createAwsAccount(ctx context.Context, client *hc.Client, d *schema.ResourceData) (diag.Diagnostics, int) {
	var diags diag.Diagnostics

	// Lock to ensure one account creation process at a time
	awsAccountCreationMux.Lock()
	defer awsAccountCreationMux.Unlock()

	postCacheData := hc.AccountCacheNewAWSCreate{
		AccountEmail:              d.Get("email").(string),
		AccountAlias:              hc.OptionalValue[string](d, "account_alias"),
		CommercialAccountName:     d.Get("commercial_account_name").(string),
		CreateGovcloud:            hc.OptionalValue[bool](d, "create_govcloud"),
		GovAccountName:            d.Get("gov_account_name").(string),
		IncludeLinkedAccountSpend: hc.OptionalValue[bool](d, "include_linked_account_spend"),
		LinkedRole:                d.Get("linked_role").(string),
		Name:                      d.Get("name").(string),
		PayerID:                   d.Get("payer_id").(int),
	}

	// Populate organizational unit details from Terraform resource data
	if err := populateOrgUnitFromResourceData(d, &postCacheData); err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("failed to populate organizational unit data: %v", err))...), 0
	}

	// Log the request data
	if rb, err := json.Marshal(postCacheData); err == nil {
		tflog.Debug(ctx, "Creating new AWS account via POST /v3/account-cache/create?account-type=aws", map[string]interface{}{
			"postData": string(rb),
		})
	}

	// Send the POST request to create the AWS account
	respCache, err := client.POST("/v3/account-cache/create?account-type=aws", postCacheData)
	if err != nil || respCache.RecordID == 0 {
		if err == nil {
			err = fmt.Errorf("received item ID of 0")
		}
		return append(diags, hc.HandleError(fmt.Errorf("unable to create AWS Account: %v", err))...), 0
	}

	// Wait for the account to be fully created
	if err := waitForAccountCreation(client, ctx, respCache.RecordID, d); err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("failed waiting for account creation: %v", err))...), 0
	}

	return diags, respCache.RecordID
}

// populateOrgUnitFromResourceData parses OU details from Terraform data, updating AccountCacheNewAWSCreate for account creation.
func populateOrgUnitFromResourceData(d *schema.ResourceData, postCacheData *hc.AccountCacheNewAWSCreate) error {
	if v, exists := d.GetOk("aws_organizational_unit"); exists {
		orgUnitSet := v.(*schema.Set)
		for _, item := range orgUnitSet.List() {
			orgUnitMap, ok := item.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid format for aws_organizational_unit")
			}
			postCacheData.OrganizationalUnit = &hc.PayerOrganizationalUnit{
				Name:      orgUnitMap["name"].(string),
				OrgUnitId: orgUnitMap["org_unit_id"].(string),
			}
		}
	}
	return nil
}

// waitForAccountCreation polls the creation status until the account is created or a timeout occurs.
func waitForAccountCreation(client *hc.Client, ctx context.Context, accountCacheId int, d *schema.ResourceData) error {
	createStateConf := &retry.StateChangeConf{
		// Define the refresh function, which checks the account creation status.
		Refresh: func() (interface{}, string, error) {
			resp := new(hc.AccountResponse)
			err := client.GET(fmt.Sprintf("/v3/account-cache/%d", accountCacheId), resp)
			if err != nil {
				// Directly return errors, including NotFound, allowing the SDK to handle retries for NotFound appropriately.
				tflog.Trace(ctx, fmt.Sprintf("Checking new AWS account status: /v3/account-cache/%d error", accountCacheId), map[string]interface{}{"error": err, "accountCacheId": accountCacheId})
				return nil, "", err
			}

			// Check if the account number is still not available in the response.
			if resp.Data.AccountNumber == "" {
				tflog.Trace(ctx, fmt.Sprintf("Checking new AWS account status: /v3/account-cache/%d missing account number", accountCacheId), map[string]interface{}{"accountCacheId": accountCacheId})
				return resp, "MissingAccountNumber", nil
			}

			// Account creation is successful.
			return resp, "AccountCreated", nil
		},
		Pending: []string{"MissingAccountNumber"},
		Target:  []string{"AccountCreated"},
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	// Use WaitForStateContext to respect the given context's deadline or cancellation.
	_, err := createStateConf.WaitForStateContext(ctx)
	return err // Return the error, if any.
}

func retryConvertCacheAccountToProjectAccountForAWS(client *hc.Client, accountCacheId, projectId int, startDatecode string, retries int, delay time.Duration) (int, error) {
	var lastErr error
	for i := 0; i < retries; i++ {
		id, err := convertCacheAccountToProjectAccount(client, accountCacheId, projectId, startDatecode)
		if err == nil {
			return id, nil
		}
		if strings.Contains(err.Error(), "Rule is already in progress") && i < retries-1 {
			time.Sleep(delay)
			continue
		}
		lastErr = err
		break
	}
	return 0, lastErr
}

// resourceAwsAccountRead attempts to read an AWS account from either the project accounts
// or account cache in Kion. By default, it tries the project accounts first and falls back
// to the cache if needed. The location can be explicitly specified using the account_id=
// or account_cache_id= prefix when importing.
func resourceAwsAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	// Handle special case for importing accounts with prefixes
	var accountLocation string
	locationChanged := false
	if strings.HasPrefix(ID, "account_id=") {
		ID = strings.TrimPrefix(ID, "account_id=")
		accountLocation = ProjectLocation
		locationChanged = true
	} else if strings.HasPrefix(ID, "account_cache_id=") {
		ID = strings.TrimPrefix(ID, "account_cache_id=")
		accountLocation = CacheLocation
		locationChanged = true
	} else {
		// For direct imports without a prefix, try project location first
		accountLocation = ProjectLocation
	}

	// Log the read operation
	tflog.Debug(ctx, "Reading AWS account", map[string]interface{}{
		"id":       ID,
		"location": accountLocation,
	})

	// Try to fetch from the determined location
	var resp hc.MappableResponse
	var err error

	if accountLocation == ProjectLocation {
		// Try project account first
		resp = new(hc.AccountResponse)
		err = client.GET(fmt.Sprintf("/v3/account/%s", ID), resp)
		if err != nil && !locationChanged {
			// If project account lookup fails and location wasn't explicitly set,
			// try cache account
			resp = new(hc.AccountCacheResponse)
			err = client.GET(fmt.Sprintf("/v3/account-cache/%s", ID), resp)
			if err == nil {
				accountLocation = CacheLocation
			}
		}
	} else {
		// Try cache account directly if that's what was specified
		resp = new(hc.AccountCacheResponse)
		err = client.GET(fmt.Sprintf("/v3/account-cache/%s", ID), resp)
	}

	if err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("unable to read AWS account (ID: %s): %v", ID, err))...)
	}

	// Set location if it was determined during the read
	if !locationChanged {
		diags = append(diags, hc.SafeSet(d, "location", accountLocation, "Failed to set location")...)
	}

	// Map response data to schema
	data := resp.ToMap("kion_aws_account")
	for k, v := range data {
		diags = append(diags, hc.SafeSet(d, k, v, "Unable to set AWS account field")...)
	}

	// Handle labels for project accounts
	if accountLocation == ProjectLocation {
		labelData, err := hc.ReadResourceLabels(client, "account", ID)
		if err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("unable to read AWS account labels (ID: %s): %v", ID, err))...)
		}
		diags = append(diags, hc.SafeSet(d, "labels", labelData, "Failed to set account labels")...)
	}

	return diags
}

func resourceAwsAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	// Log the update operation
	tflog.Debug(ctx, "Updating AWS account", map[string]interface{}{
		"id":       ID,
		"location": getKionAccountLocation(d),
	})

	hasChanged := false

	// Handle account location changes
	if d.HasChange("project_id") {
		hasChanged = true
		oldId, newId := d.GetChange("project_id")
		oldProjectId := oldId.(int)
		newProjectId := newId.(int)

		if oldProjectId == 0 && newProjectId != 0 {
			// Converting from cache to project
			diags = append(diags, handleCacheToProjectConversion(ctx, d, client)...)
		} else if oldProjectId != 0 && newProjectId == 0 {
			// Converting from project to cache
			diags = append(diags, handleProjectToCacheConversion(ctx, d, client)...)
		} else if oldProjectId != newProjectId {
			// Moving between projects
			diags = append(diags, handleProjectToProjectMove(ctx, d, client)...)
		}

		if diags.HasError() {
			return diags
		}
	}

	// Handle other updatable fields
	if d.HasChanges("account_alias", "email", "include_linked_account_spend",
		"linked_role", "name", "skip_access_checking",
		"start_datecode", "use_org_account_info") {
		hasChanged = true
		diags = append(diags, handleAccountUpdate(ctx, d, client)...)
		if diags.HasError() {
			return diags
		}
	}

	// Handle label changes for project accounts
	if getKionAccountLocation(d) == ProjectLocation && d.HasChange("labels") {
		hasChanged = true
		if err := hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "account", ID); err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("unable to update AWS account labels (ID: %s): %v", ID, err))...)
		}
	}

	if hasChanged {
		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("unable to set last_updated: %v", err))...)
		}
	}

	return append(diags, resourceAwsAccountRead(ctx, d, m)...)
}

func resourceAwsAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAccountDelete(ctx, d, m)
}

// Require startDatecode if adding to a new project, unless we are creating the account.
func validateAwsAccountStartDatecode(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	// if start date is already set, nothing to do
	if _, ok := d.GetOk("start_datecode"); ok {
		return nil
	}

	// if not adding to project, we don't care about start date
	if _, ok := d.GetOk("project_id"); !ok {
		return nil
	}

	// if there is no account_number, then we are are creating a new Account and
	// start date isn't required since it will be set to the current month
	if _, ok := d.GetOk("account_number"); !ok {
		return nil
	}

	// otherwise, start_datecode is required
	return fmt.Errorf("start_datecode is required when adding an existing AWS account to a project")
}

// Helper functions for AWS account updates and conversions

func handleCacheToProjectConversion(ctx context.Context, d *schema.ResourceData, client *hc.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	ID := strings.TrimPrefix(d.Id(), "account_cache_id=")

	accountCacheId, err := strconv.Atoi(ID)
	if err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("invalid account cache ID: %v", err))...)
	}

	projectId := d.Get("project_id").(int)
	startDatecode := d.Get("start_datecode").(string)

	tflog.Debug(ctx, "Converting AWS account from cache to project", map[string]interface{}{
		"account_cache_id": accountCacheId,
		"project_id":       projectId,
		"start_datecode":   startDatecode,
	})

	newId, err := convertCacheAccountToProjectAccount(client, accountCacheId, projectId, startDatecode)
	if err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("failed to convert cache account to project: %v", err))...)
	}

	d.SetId(fmt.Sprintf("%d", newId))
	return diags
}

func handleProjectToCacheConversion(ctx context.Context, d *schema.ResourceData, client *hc.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	accountId, err := strconv.Atoi(d.Id())
	if err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("invalid account ID: %v", err))...)
	}

	tflog.Debug(ctx, "Converting AWS account from project to cache", map[string]interface{}{
		"account_id": accountId,
	})

	newId, err := convertProjectAccountToCacheAccount(client, accountId)
	if err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("failed to convert project account to cache: %v", err))...)
	}

	d.SetId(fmt.Sprintf("%d", newId))
	return diags
}

func handleProjectToProjectMove(ctx context.Context, d *schema.ResourceData, client *hc.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	ID := d.Id()

	req := hc.AccountMove{
		ProjectID:        d.Get("project_id").(int),
		FinancialSetting: "move",
		MoveDate:         0,
	}

	// Get move settings if provided
	if v, exists := d.GetOk("move_project_settings"); exists {
		moveSettings := v.(*schema.Set)
		for _, item := range moveSettings.List() {
			if moveSettingsMap, ok := item.(map[string]interface{}); ok {
				req.FinancialSetting = moveSettingsMap["financials"].(string)
				if val, ok := moveSettingsMap["move_datecode"]; ok {
					req.MoveDate = val.(int)
				}
			}
		}
	}

	tflog.Debug(ctx, "Moving AWS account between projects", map[string]interface{}{
		"account_id":        ID,
		"project_id":        req.ProjectID,
		"financial_setting": req.FinancialSetting,
		"move_date":         req.MoveDate,
	})

	resp, err := client.POST(fmt.Sprintf("/v3/account/%s/move", ID), req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to move account between projects: %v", err))
	}

	d.SetId(fmt.Sprintf("%d", resp.RecordID))
	return diags
}

func handleAccountUpdate(ctx context.Context, d *schema.ResourceData, client *hc.Client) diag.Diagnostics {
	var diags diag.Diagnostics
	ID := d.Id()
	accountLocation := getKionAccountLocation(d)

	var req interface{}
	var accountUrl string
	switch accountLocation {
	case CacheLocation:
		accountUrl = fmt.Sprintf("/v3/account-cache/%s", ID)
		cacheReq := hc.AccountCacheUpdatable{}
		if v, ok := d.GetOk("account_alias"); ok {
			AccountAlias := v.(string)
			cacheReq.AccountAlias = &AccountAlias
		} else if d.HasChange("account_alias") {
			emptyAlias := ""
			cacheReq.AccountAlias = &emptyAlias
		}
		if v, ok := d.GetOk("email"); ok {
			email := v.(string)
			cacheReq.AccountEmail = email
		}
		if v, ok := d.GetOk("linked_role"); ok {
			linkedRole := v.(string)
			cacheReq.LinkedRole = linkedRole
		}
		if v, ok := d.GetOk("name"); ok {
			name := v.(string)
			cacheReq.Name = name
		}
		if v, ok := d.GetOk("include_linked_account_spend"); ok {
			includeLinkedSpend := v.(bool)
			cacheReq.IncludeLinkedAccountSpend = &includeLinkedSpend
		}
		if v, ok := d.GetOk("skip_access_checking"); ok {
			skipAccess := v.(bool)
			cacheReq.SkipAccessChecking = &skipAccess
		}
		req = cacheReq

	case ProjectLocation:
		fallthrough
	default:
		accountUrl = fmt.Sprintf("/v3/account/%s", ID)
		accountReq := hc.AccountUpdatable{}
		if v, ok := d.GetOk("account_alias"); ok {
			AccountAlias := v.(string)
			accountReq.AccountAlias = &AccountAlias
		} else if d.HasChange("account_alias") {
			emptyAlias := ""
			accountReq.AccountAlias = &emptyAlias
		}
		if v, ok := d.GetOk("email"); ok {
			email := v.(string)
			accountReq.AccountEmail = email
		}
		if v, ok := d.GetOk("linked_role"); ok {
			linkedRole := v.(string)
			accountReq.LinkedRole = linkedRole
		}
		if v, ok := d.GetOk("name"); ok {
			name := v.(string)
			accountReq.Name = name
		}
		if v, ok := d.GetOk("start_datecode"); ok {
			startDatecode := v.(string)
			accountReq.StartDatecode = startDatecode
		}
		if v, ok := d.GetOk("include_linked_account_spend"); ok {
			includeLinkedSpend := v.(bool)
			accountReq.IncludeLinkedAccountSpend = &includeLinkedSpend
		}
		if v, ok := d.GetOk("skip_access_checking"); ok {
			skipAccess := v.(bool)
			accountReq.SkipAccessChecking = &skipAccess
		}
		if v, ok := d.GetOk("use_org_account_info"); ok {
			useOrgInfo := v.(bool)
			accountReq.UseOrgAccountInfo = &useOrgInfo
		}
		req = accountReq
	}

	tflog.Debug(ctx, "Updating AWS account", map[string]interface{}{
		"account_id": ID,
		"location":   accountLocation,
		"url":        accountUrl,
	})

	if err := client.PATCH(accountUrl, req); err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("failed to update account: %v", err))...)
	}

	return diags
}
