package kion

import (
	"context"
	"encoding/json"
	"errors"
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
			"When importing an existing Kion account into terraform state (different from using terraform to " +
			"import an existing AWS account into Kion), you must use the `account_id=` or `account_cache_id=` " +
			"ID prefix to indicate whether the ID is an account ID or a cached account ID.\n\n" +
			"For example:\n\n" +
			"    terraform import kion_aws_account.test-account account_id=123\n" +
			"    terraform import kion_aws_account.test-cached-account account_cache_id=321\n\n" +
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
				Description: "Where the account is attached.  Either \"project\" or \"cache\".",
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
				Description: "The ID of the Kion project to place this account within.  If empty, the account will be placed within the account cache.",
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

	accountLocation := getKionAccountLocation(d)

	if _, ok := d.GetOk("account_number"); ok {
		// Import an existing AWS account

		// Default to AWS commercial if not otherwise set
		// TODO: Why is this required for cache import, but not project import??
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

		if rb, err := json.Marshal(postAccountData); err == nil {
			tflog.Debug(ctx, fmt.Sprintf("Importing exiting AWS account via POST %s", accountUrl), map[string]interface{}{"postData": string(rb)})
		}
		resp, err := client.POST(accountUrl, postAccountData)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to import AWS Account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), postAccountData),
			})
			return diags
		} else if resp.RecordID == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to import AWS Account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), postAccountData),
			})
			return diags
		}

		diags = append(diags, hc.SafeSet(d, "location", accountLocation, "Failed to set location")...)
		if diags.HasError() {
			return diags
		}

		d.SetId(strconv.Itoa(resp.RecordID))

	} else {
		// Call the createAwsAccount function
		diags, accountCacheId := createAwsAccount(ctx, client, d)
		if diags.HasError() {
			return diags
		}

		switch accountLocation {
		case ProjectLocation:
			// Move cached account to the requested project
			projectId := d.Get("project_id").(int)
			startDatecode := time.Now().Format("200601")
			retries := 3              // Number of retries
			delay := 30 * time.Second // Delay between retries

			newId, err := retryConvertCacheAccountToProjectAccountForAWS(client, accountCacheId, projectId, startDatecode, retries, delay)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to convert AWS cached account to project account",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), accountCacheId),
				})
				diags = append(diags, resourceAwsAccountRead(ctx, d, m)...)
				return diags
			}

			diags = append(diags, hc.SafeSet(d, "location", accountLocation, "Failed to set location")...)
			if diags.HasError() {
				return diags
			}

			d.SetId(strconv.Itoa(newId))

		case CacheLocation:
			// Track the cached account
			diags = append(diags, hc.SafeSet(d, "location", accountLocation, "Failed to set location")...)
			if diags.HasError() {
				return diags
			}

			d.SetId(strconv.Itoa(accountCacheId))
		}
	}

	// Labels are only supported on project accounts, not cached accounts
	if accountLocation == ProjectLocation {
		if _, ok := d.GetOk("labels"); ok {
			ID := d.Id()
			err := hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "account", ID)

			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to update AWS account labels",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				diags = append(diags, resourceAwsAccountRead(ctx, d, m)...)
				return diags
			}
		}
	}

	return append(diags, resourceAwsAccountRead(ctx, d, m)...)
}

func createAwsAccount(ctx context.Context, client *hc.Client, d *schema.ResourceData) (diag.Diagnostics, int) {
	var diags diag.Diagnostics

	// Lock to ensure one account creation process at a time as AWS Orgs cannot handle more than one account creation at a time.
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

	// Populate organizational unit details from Terraform resource data, if provided by user.
	if err := populateOrgUnitFromResourceData(client, &postCacheData, d); err != nil {
		return diag.FromErr(err), 0
	}

	// Log the account creation POST data.
	if err := logPostData(ctx, client, postCacheData); err != nil {
		// If logging fails, only warn instead of failing the operation.
		tflog.Warn(ctx, "Failed to log post data for AWS account creation", map[string]interface{}{"error": err.Error()})
	}

	// Send the POST request to create the AWS account.
	respCache, err := client.POST("/v3/account-cache/create?account-type=aws", postCacheData)
	if err != nil || respCache.RecordID == 0 {
		if err == nil {
			err = fmt.Errorf("received item ID of 0")
		}
		return diag.Errorf("Unable to create AWS Account: %v", err), 0
	}

	// Wait for the account to be fully created.
	if err := waitForAccountCreation(client, ctx, respCache.RecordID, d); err != nil {
		return diag.FromErr(err), 0
	}

	// Return any diagnostics and the ID of the created account cache.
	return diags, respCache.RecordID
}

// logPostData logs the data being posted for account creation. It returns an error if marshaling fails.
func logPostData(ctx context.Context, client *hc.Client, postData interface{}) error {
	// This line makes the linter recognize that client is used
	_ = client
	rb, err := json.Marshal(postData)
	if err != nil {
		return err
	}
	tflog.Debug(ctx, "Creating new AWS account via POST /v3/account-cache/create?account-type=aws", map[string]interface{}{"postData": string(rb)})
	return nil
}

// populateOrgUnitFromResourceData parses OU details from Terraform data, updating AccountCacheNewAWSCreate for account creation.
func populateOrgUnitFromResourceData(client *hc.Client, postCacheData *hc.AccountCacheNewAWSCreate, d *schema.ResourceData) error {
	// This line makes the linter recognize that client is used
	_ = client
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

func resourceAwsAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAccountRead("kion_aws_account", ctx, d, m)
}

func resourceAwsAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := resourceAccountUpdate(ctx, d, m)
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
