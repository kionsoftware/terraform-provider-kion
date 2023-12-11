package kion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/ctclient"
)

func resourceAzureAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAzureAccountCreate,
		ReadContext:   resourceAzureAccountRead,
		UpdateContext: resourceAzureAccountUpdate,
		DeleteContext: resourceAzureAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceAzureAccountRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subscription_uuid": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				AtLeastOneOf: []string{"ea", "csp", "mca"},
			},
			"subscription_name": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"subscription_uuid"},
			},
			"parent_management_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"csp": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"offer_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"billing_cycle": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"ea": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeString,
							Required: true,
						},
						"billing_account": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"mca": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"billing_account": {
							Type:     schema.TypeString,
							Required: true,
						},
						"billing_profile": {
							Type:     schema.TypeString,
							Required: true,
						},
						"billing_profile_invoice": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"payer_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"project_id": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"start_datecode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"skip_access_checking": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"account_type_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"move_project_settings": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"financials": {
							Type:         schema.TypeString,
							Default:      "move",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"preserve", "move"}, false),
						},
						"move_datecode": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"location": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Where the account is attached.  Either \"project\" or \"cache\".",
			},
			"labels": {
				Type:         schema.TypeMap,
				Optional:     true,
				RequiredWith: []string{"project_id"},
				Elem:         &schema.Schema{Type: schema.TypeString},
				Description:  "A map of labels to assign to the account. The labels must already exist in Kion.",
			},
		},
		CustomizeDiff: customdiff.All(
			// schema validators don't support multi-attribute validations, so we use CustomizeDiff instead
			validateAzureAccountStartDatecode,
			customDiffComputedAccountLocation,
		),
	}
}

func resourceAzureAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	accountLocation := getKionAccountLocation(d)

	if _, ok := d.GetOk("subscription_uuid"); ok {
		// Import an existing Azure subscription

		var postAccountData interface{}
		var accountUrl string
		switch accountLocation {
		case CacheLocation:
			accountUrl = "/v3/account-cache?account-type=azure"
			postAccountData = hc.AccountCacheNewAzureImport{
				SubscriptionUUID:   d.Get("subscription_uuid").(string),
				Name:               d.Get("name").(string),
				AccountTypeID:      hc.OptionalInt(d, "account_type_id"),
				PayerID:            d.Get("payer_id").(int),
				SkipAccessChecking: hc.OptionalBool(d, "skip_access_checking"),
			}

		case ProjectLocation:
			fallthrough
		default:
			accountUrl = "/v3/account?account-type=azure"
			postAccountData = hc.AccountNewAzureImport{
				SubscriptionUUID:   d.Get("subscription_uuid").(string),
				Name:               d.Get("name").(string),
				AccountTypeID:      hc.OptionalInt(d, "account_type_id"),
				PayerID:            d.Get("payer_id").(int),
				ProjectID:          d.Get("project_id").(int),
				SkipAccessChecking: hc.OptionalBool(d, "skip_access_checking"),
				StartDatecode:      d.Get("start_datecode").(string),
			}
		}

		if rb, err := json.Marshal(postAccountData); err == nil {
			tflog.Debug(ctx, fmt.Sprintf("Importing exiting Azure account via POST %s", accountUrl), map[string]interface{}{"postData": string(rb)})
		}
		resp, err := c.POST(accountUrl, postAccountData)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to import Azure Account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), postAccountData),
			})
			return diags
		} else if resp.RecordID == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to import Azure Account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), postAccountData),
			})
			return diags
		}

		d.Set("location", accountLocation)
		d.SetId(strconv.Itoa(resp.RecordID))

	} else {
		// Create a new Azure subscription

		postCacheData := hc.AccountCacheNewAzureCreate{
			Name:                    d.Get("name").(string),
			SubscriptionName:        d.Get("subscription_name").(string),
			ParentManagementGroupID: d.Get("parent_management_group_id").(string),
			PayerID:                 d.Get("payer_id").(int),
		}

		if v, exists := d.GetOk("csp"); exists {
			cspSet := v.(*schema.Set)
			for _, item := range cspSet.List() {
				if cspMap, ok := item.(map[string]interface{}); ok {
					postCacheData.SubscriptionCSPBillingInfo = &hc.SubscriptionCSPBillingInfo{
						BillingCycle: cspMap["billing_cycle"].(string),
						OfferID:      cspMap["offer_id"].(string),
					}
				}
			}
		}

		if v, exists := d.GetOk("ea"); exists {
			eaSet := v.(*schema.Set)
			for _, item := range eaSet.List() {
				if eaMap, ok := item.(map[string]interface{}); ok {
					postCacheData.SubscriptionEABillingInfo = &hc.SubscriptionEABillingInfo{
						BillingAccountNumber: eaMap["billing_account"].(string),
						EAAccountNumber:      eaMap["account"].(string),
					}
				}
			}
		}

		if v, exists := d.GetOk("mca"); exists {
			mcaSet := v.(*schema.Set)
			for _, item := range mcaSet.List() {
				if mcaMap, ok := item.(map[string]interface{}); ok {
					postCacheData.SubscriptionMCABillingInfo = &hc.SubscriptionMCABillingInfo{
						BillingAccountNumber: mcaMap["billing_account"].(string),
						BillingProfileNumber: mcaMap["billing_profile"].(string),
						InvoiceSectionNumber: mcaMap["billing_profile_invoice"].(string),
					}
				}
			}
		}

		if rb, err := json.Marshal(postCacheData); err == nil {
			tflog.Debug(ctx, "Creating new Azure account via POST /v3/account-cache/create?account-type=azure", map[string]interface{}{"postData": string(rb)})
		}
		respCache, err := c.POST("/v3/account-cache/create?account-type=azure", postCacheData)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Azure Account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), postCacheData),
			})
			return diags
		} else if respCache.RecordID == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Azure Account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), postCacheData),
			})
			return diags
		}

		accountCacheId := respCache.RecordID

		// Wait for account to be created
		createStateConf := &resource.StateChangeConf{
			Refresh: func() (interface{}, string, error) {
				resp := new(hc.AccountResponse)
				err := c.GET(fmt.Sprintf("/v3/account-cache/%d", accountCacheId), resp)
				if err != nil {
					if resErr, ok := err.(*hc.RequestError); ok {
						if resErr.StatusCode == http.StatusNotFound {
							// StateChangeConf handles 404s differently than errors, so return nil instead of err
							tflog.Trace(ctx, fmt.Sprintf("Checking new Azure account status: /v3/account-cache/%d not found", accountCacheId))
							return nil, "NotFound", nil
						}
					}
					tflog.Trace(ctx, fmt.Sprintf("Checking new Azure account status: /v3/account-cache/%d error", accountCacheId), map[string]interface{}{"error": err})
					return nil, "Error", err
				}
				if resp.Data.AccountNumber == "" {
					tflog.Trace(ctx, fmt.Sprintf("Checking new Azure account status: /v3/account-cache/%d missing account number", accountCacheId))
					return resp, "MissingSubscriptionId", nil
				}
				return resp, "AccountCreated", nil
			},
			Pending: []string{
				"MissingSubscriptionId",
			},
			Target: []string{
				"AccountCreated",
			},
			Timeout: d.Timeout(schema.TimeoutCreate),
		}
		_, err = createStateConf.WaitForState()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Azure Account",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
			return diags
		}

		switch accountLocation {
		case ProjectLocation:
			// Move cached account to the requested project
			projectId := d.Get("project_id").(int)
			startDatecode := time.Now().Format("200601")

			newId, err := convertCacheAccountToProjectAccount(c, accountCacheId, projectId, startDatecode)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to convert Azure cached account to project account",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), accountCacheId),
				})
				diags = append(diags, resourceAzureAccountRead(ctx, d, m)...)
				return diags
			}

			d.Set("location", accountLocation)
			d.SetId(strconv.Itoa(newId))

		case CacheLocation:
			// Track the cached account
			d.Set("location", accountLocation)
			d.SetId(strconv.Itoa(accountCacheId))
		}
	}

	// Labels are only supported on project accounts, not cached accounts
	if accountLocation == ProjectLocation {
		if _, ok := d.GetOk("labels"); ok {
			ID := d.Id()
			err := hc.PutAppLabelIDs(c, hc.FlattenAssociateLabels(d, "labels"), "account", ID)

			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to update Azure account labels",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				diags = append(diags, resourceAzureAccountRead(ctx, d, m)...)
				return diags
			}
		}
	}

	return append(diags, resourceAzureAccountRead(ctx, d, m)...)
}

func resourceAzureAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAccountRead("kion_azure_account", ctx, d, m)
}

func resourceAzureAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := resourceAccountUpdate(ctx, d, m)
	return append(diags, resourceAzureAccountRead(ctx, d, m)...)
}

func resourceAzureAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAccountDelete(ctx, d, m)
}

// Require startDatecode if adding to a new project, unless we are creating the account.
func validateAzureAccountStartDatecode(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	// if start date is already set, nothing to do
	if _, ok := d.GetOk("start_datecode"); ok {
		return nil
	}

	// if not adding to project, we don't care about start date
	if _, ok := d.GetOk("project_id"); !ok {
		return nil
	}

	// if there is no subscription_uuid, then we are are creating a new Subscription and
	// start date isn't required since it will be set to the current month
	if _, ok := d.GetOk("subscription_uuid"); !ok {
		return nil
	}

	// otherwise, start_datecode is required
	return fmt.Errorf("start_datecode is required when adding an existing Azure subscription to a project")
}
