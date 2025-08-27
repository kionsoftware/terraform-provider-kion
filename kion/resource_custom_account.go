package kion

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceCustomAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Attaches a custom account to a project in Kion.\n\n" +
			"If `project_id` is provided, the account will be added to the corresponding project, " +
			"otherwise the account will be added to the account cache.\n\n" +
			"Once added, an account can be moved between projects or in and out of the account cache by " +
			"changing the `project_id`.\n\n" +
			"When importing existing Kion accounts into terraform state, you can use one of these methods:\n\n" +
			"1. Default import (tries project first, then cache):\n" +
			"    terraform import kion_custom_account.example 123\n\n" +
			"2. Explicit project account import:\n" +
			"    terraform import kion_custom_account.example account_id=123\n\n" +
			"3. Explicit cache account import:\n" +
			"    terraform import kion_custom_account.example account_cache_id=123",
		CreateContext: resourceCustomAccountCreate,
		ReadContext:   resourceCustomAccountRead,
		UpdateContext: resourceCustomAccountUpdate,
		DeleteContext: resourceCustomAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceCustomAccountRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"account_alias": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Account alias is an optional short unique name that helps identify the account within Kion.",
			},
			"account_number": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The account number of the custom account.",
				ForceNew:    true,
			},
			"account_type_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "An ID representing the account type within Kion.",
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
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
			"location": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Where the account is attached. Either \"project\" or \"cache\".",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the custom account within Kion.",
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
				Description: "Date when the custom account will starting submitting payments against a funding source (YYYY-MM). Required if placing an account within a project.",
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^\d{4}-\d{2}$`),
					"start_datecode must be in the format YYYY-MM (e.g., '2024-03')",
				),
			},
		},
		CustomizeDiff: customdiff.All(
			validateCustomAccountStartDatecode,
			customDiffComputedAccountLocation,
		),
	}
}

func resourceCustomAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	// Set initial location
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

	// Import the existing custom account
	var postAccountData interface{}
	var accountUrl string
	switch accountLocation {
	case CacheLocation:
		accountUrl = "/v3/account-cache?account-type=custom"
		postAccountData = hc.AccountCacheNewCustomImport{
			AccountAlias:  hc.OptionalValue[string](d, "account_alias"),
			AccountNumber: d.Get("account_number").(string),
			Name:          d.Get("name").(string),
			PayerID:       d.Get("payer_id").(int),
		}

	case ProjectLocation:
		fallthrough
	default:
		accountUrl = "/v3/account?account-type=custom"
		postAccountData = hc.AccountNewCustomImport{
			AccountAlias:  hc.OptionalValue[string](d, "account_alias"),
			AccountNumber: d.Get("account_number").(string),
			Name:          d.Get("name").(string),
			PayerID:       d.Get("payer_id").(int),
			ProjectID:     d.Get("project_id").(int),
			StartDatecode: d.Get("start_datecode").(string),
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("Importing custom account via POST %s", accountUrl), map[string]interface{}{
		"url": accountUrl,
	})

	resp, err := client.POST(accountUrl, postAccountData)
	if err != nil {
		diags = append(diags, hc.HandleError(fmt.Errorf("unable to import custom account: %v", err))...)
		return diags
	}

	d.SetId(fmt.Sprintf("%d", resp.RecordID))

	// Labels are only supported on project accounts, not cached accounts
	if accountLocation == ProjectLocation {
		if _, ok := d.GetOk("labels"); ok {
			ID := d.Id()
			if err := hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "account", ID); err != nil {
				return append(diags, hc.HandleError(fmt.Errorf("unable to update custom account labels (ID: %s): %v", ID, err))...)
			}
		}
	}

	return append(diags, resourceCustomAccountRead(ctx, d, m)...)
}

func resourceCustomAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAccountRead("kion_custom_account", ctx, d, m)
}

func resourceCustomAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAccountUpdate(ctx, d, m)
}

func resourceCustomAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAccountDelete(ctx, d, m)
}

// Require startDatecode if adding to a new project
func validateCustomAccountStartDatecode(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	// if start date is already set, nothing to do
	if _, ok := d.GetOk("start_datecode"); ok {
		return nil
	}

	// if not adding to project, we don't care about start date
	if _, ok := d.GetOk("project_id"); !ok {
		return nil
	}

	// otherwise, start_datecode is required
	return fmt.Errorf("start_datecode is required when adding a custom account to a project")
}
