package kion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/ctclient"
)

func resourceGcpAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGcpAccountCreate,
		ReadContext:   resourceGcpAccountRead,
		UpdateContext: resourceGcpAccountUpdate,
		DeleteContext: resourceGcpAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceGcpAccountRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"create_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"create", "import"}, false),
				Description:  "One of \"create\" or \"import\".  If \"create\", Kion will attempt to create a new Google Cloud Project.  If \"import\", Kion will import the existing Google Cloud Project as specified by google_cloud_project_id.",
			},
			"google_cloud_project_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"google_cloud_parent_name": {
				Type:     schema.TypeString,
				Optional: true,
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
			validateGcpAccountStartDatecode,
			customDiffComputedAccountLocation,
		),
	}
}

func resourceGcpAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	accountLocation := getKionAccountLocation(d)

	if strings.ToLower(d.Get("create_mode").(string)) == "import" {
		// Import an existing GCP project

		var postAccountData interface{}
		var accountUrl string
		switch accountLocation {
		case CacheLocation:
			accountUrl = "/v3/account-cache?account-type=google-cloud"
			postAccountData = hc.AccountCacheNewGCPImport{
				Name:                 d.Get("name").(string),
				PayerID:              d.Get("payer_id").(int),
				AccountTypeID:        hc.OptionalInt(d, "account_type_id"),
				GoogleCloudProjectID: d.Get("google_cloud_project_id").(string),
				SkipAccessChecking:   hc.OptionalBool(d, "skip_access_checking"),
			}

		case ProjectLocation:
			fallthrough
		default:
			accountUrl = "/v3/account?account-type=google-cloud"
			postAccountData = hc.AccountNewGCPImport{
				Name:                 d.Get("name").(string),
				PayerID:              d.Get("payer_id").(int),
				AccountTypeID:        hc.OptionalInt(d, "account_type_id"),
				GoogleCloudProjectID: d.Get("google_cloud_project_id").(string),
				SkipAccessChecking:   hc.OptionalBool(d, "skip_access_checking"),
				ProjectID:            d.Get("project_id").(int),
				StartDatecode:        d.Get("start_datecode").(string),
			}
		}

		if rb, err := json.Marshal(postAccountData); err == nil {
			tflog.Debug(ctx, fmt.Sprintf("Importing exiting GCP Project via POST %s", accountUrl), map[string]interface{}{"postData": string(rb)})
		}
		resp, err := c.POST(accountUrl, postAccountData)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to import GCP Project",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), postAccountData),
			})
			return diags
		} else if resp.RecordID == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to import GCP Project",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), postAccountData),
			})
			return diags
		}

		d.Set("location", accountLocation)
		d.SetId(strconv.Itoa(resp.RecordID))

	} else {
		// Create a new GCP project

		postCacheData := hc.AccountCacheNewGCPCreate{
			DisplayName:           d.Get("name").(string),
			PayerID:               d.Get("payer_id").(int),
			GoogleCloudProjectID:  d.Get("google_cloud_project_id").(string),
			GoogleCloudParentName: d.Get("google_cloud_parent_name").(string),
		}

		if rb, err := json.Marshal(postCacheData); err == nil {
			tflog.Debug(ctx, "Creating new GCP account via POST /v3/account-cache/create?account-type=google-cloud", map[string]interface{}{"data": string(rb)})
		}
		respCache, err := c.POST("/v3/account-cache/create?account-type=google-cloud", postCacheData)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create GCP Project",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), postCacheData),
			})
			return diags
		} else if respCache.RecordID == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create GCP Project",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), postCacheData),
			})
			return diags
		}

		accountCacheId := respCache.RecordID

		// The API doesn't give any indication of when the GCP project has been created.
		// Instead we'll poll a few times to see if the cached account gets deleted.
		// TODO: Find a better way to confirm GCP Project was created.
		createStateConf := &resource.StateChangeConf{
			Refresh: func() (interface{}, string, error) {
				resp := new(hc.AccountResponse)
				err := c.GET(fmt.Sprintf("/v3/account-cache/%d", accountCacheId), resp)
				if err != nil {
					if resErr, ok := err.(*hc.RequestError); ok {
						if resErr.StatusCode == http.StatusNotFound {
							// StateChangeConf handles 404s differently than errors, so return nil instead of err
							tflog.Trace(ctx, fmt.Sprintf("Checking new GCP account status: /v3/account-cache/%d not found", accountCacheId))
							return nil, "NotFound", nil
						}
					}
					tflog.Trace(ctx, fmt.Sprintf("Checking new GCP account status: /v3/account-cache/%d error", accountCacheId), map[string]interface{}{"error": err})
					return nil, "Error", err
				}
				return resp, "AccountExists", nil
			},
			Target: []string{
				"AccountExists",
			},
			Timeout:                   d.Timeout(schema.TimeoutCreate),
			ContinuousTargetOccurence: 10,
		}
		_, err = createStateConf.WaitForState()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create GCP Project",
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
					Summary:  "Unable to convert GCP cached account to project account",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), accountCacheId),
				})
				diags = append(diags, resourceGcpAccountRead(ctx, d, m)...)
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
					Summary:  "Unable to update GCP account labels",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				diags = append(diags, resourceGcpAccountRead(ctx, d, m)...)
				return diags
			}
		}
	}

	return append(diags, resourceGcpAccountRead(ctx, d, m)...)
}

func resourceGcpAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAccountRead("kion_gcp_account", ctx, d, m)
}

func resourceGcpAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	diags := resourceAccountUpdate(ctx, d, m)
	return append(diags, resourceGcpAccountRead(ctx, d, m)...)
}

func resourceGcpAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceAccountDelete(ctx, d, m)
}

// Require startDatecode if adding to a new project, unless we are creating the account.
func validateGcpAccountStartDatecode(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {

	// if start date is already set, nothing to do
	if _, ok := d.GetOk("start_datecode"); ok {
		return nil
	}

	// if not adding to project, we don't care about start date
	if _, ok := d.GetOk("project_id"); !ok {
		return nil
	}

	oldCreateMode, newCreateMode := d.GetChange("create_mode")
	isNewResource := oldCreateMode.(string) == ""

	// if we are creating a new GCP project, start date isn't required
	// since it will be set to the current month
	if newCreateMode.(string) == "create" && isNewResource {
		return nil
	}

	// otherwise, start_datecode is required
	return fmt.Errorf("start_datecode is required when adding an existing GCP project to a project")
}
