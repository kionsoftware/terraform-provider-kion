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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/ctclient"
)

// Shared methods used by kion_*_account resources.
// See one of:
//   kion/resource_aws_account.go
//   kion/resource_gcp_account.go
//   kion/resource_azure_subscription_account.go

func resourceAccountRead(resource string, ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	// HACK: Special case when importing existing accounts
	// When importing accounts we only have an ID and we don't know whether the
	// ID is an account ID or account_cache ID. To work around this, we allow the
	// user to import using an `account_id=` or `account_cache_id=` prefix.
	// For example:
	//   terraform import kion_aws_account.test-account account_id=123
	//   terraform import kion_aws_account.test-account account_cache_id=321
	//
	// TODO: Find a better way to determine if the imported ID is an account
	// or account cache by reading the resource value
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
		accountLocation = getKionAccountLocation(d)
	}

	// Fetch data from account or account-cache URL
	var resp hc.MappableResponse
	var accountUrl string
	switch accountLocation {
	case CacheLocation:
		accountUrl = fmt.Sprintf("/v3/account-cache/%s", ID)
		resp = new(hc.AccountCacheResponse)
	case ProjectLocation:
		fallthrough
	default:
		accountUrl = fmt.Sprintf("/v3/account/%s", ID)
		resp = new(hc.AccountResponse)
	}
	err := c.GET(accountUrl, resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read account",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	if locationChanged {
		d.SetId(ID)
		d.Set("location", accountLocation)
	}

	data := resp.ToMap(resource)
	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Fetch labels
	if accountLocation == ProjectLocation {
		labelData, err := hc.ReadResourceLabels(c, "account", ID)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read account labels",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}

		// Set labels
		err = d.Set("labels", labelData)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to set labels for account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
		}
	}

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	var accountLocation string
	var oldProjectId, newProjectId int
	{
		oldId, newId := d.GetChange("project_id")
		oldProjectId = oldId.(int)
		newProjectId = newId.(int)
	}

	if oldProjectId == 0 && newProjectId != 0 {
		// Handle conversion from cache account to project account
		accountCacheId, err := strconv.Atoi(ID)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to convert cached account to project account, invalid cached account id",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}

		tflog.Debug(ctx, "Converting from cached account to project account", map[string]interface{}{"oldProjectId": oldProjectId, "newProjectId": newProjectId})
		newId, err := convertCacheAccountToProjectAccount(c, accountCacheId, newProjectId, d.Get("start_datecode").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to convert cached account to project account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}

		accountLocation = ProjectLocation
		ID = strconv.Itoa(newId)
		d.Set("location", accountLocation)
		d.SetId(ID)

	} else if oldProjectId != 0 && newProjectId == 0 {
		// Handle conversion from project account to cache account
		accountId, err := strconv.Atoi(ID)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to convert project account to cache account, invalid account id",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}

		tflog.Debug(ctx, "Converting from project account to cached account", map[string]interface{}{"oldProjectId": oldProjectId, "newProjectId": newProjectId})
		newId, err := convertProjectAccountToCacheAccount(c, accountId)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to convert project account to cache account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}

		accountLocation = CacheLocation
		ID = strconv.Itoa(newId)
		d.Set("location", accountLocation)
		d.SetId(ID)

	} else {
		accountLocation = getKionAccountLocation(d)

		if accountLocation == ProjectLocation && oldProjectId != newProjectId {
			// Handle moving to a different project

			req := hc.AccountMove{
				ProjectID:        d.Get("project_id").(int),
				FinancialSetting: "move",
				MoveDate:         0,
			}
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

			if rb, err := json.Marshal(req); err == nil {
				tflog.Debug(ctx, "Moving account to different project", map[string]interface{}{"oldProjectId": oldProjectId, "newProjectId": newProjectId, "postData": string(rb)})
			}

			resp, err := c.POST(fmt.Sprintf("/v3/account/%s/move", ID), req)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to move account to a different project",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}

			ID = strconv.Itoa(resp.RecordID)
			d.SetId(ID)
		}
	}

	// Determine if the attributes that are updatable are changed.
	if d.HasChanges("email",
		"name",
		"include_linked_account_spend",
		"linked_role",
		"skip_access_checking",
		"start_datecode",
		"use_org_account_info") {
		hasChanged++

		var req interface{}
		var accountUrl string
		switch accountLocation {
		case CacheLocation:
			accountUrl = fmt.Sprintf("/v3/account-cache/%s", ID)
			cacheReq := hc.AccountCacheUpdatable{}
			if v, ok := d.GetOk("name"); ok {
				cacheReq.Name = v.(string)
			}
			if v, ok := d.GetOk("email"); ok {
				cacheReq.AccountEmail = v.(string)
			}
			if v, ok := d.GetOk("linked_role"); ok {
				cacheReq.LinkedRole = v.(string)
			}
			cacheReq.IncludeLinkedAccountSpend = hc.OptionalBool(d, "include_linked_account_spend")
			cacheReq.SkipAccessChecking = hc.OptionalBool(d, "skip_access_checking")
			req = cacheReq
		case ProjectLocation:
			fallthrough
		default:
			accountUrl = fmt.Sprintf("/v3/account/%s", ID)
			accountReq := hc.AccountUpdatable{}
			if v, ok := d.GetOk("name"); ok {
				accountReq.Name = v.(string)
			}
			if v, ok := d.GetOk("email"); ok {
				accountReq.AccountEmail = v.(string)
			}
			if v, ok := d.GetOk("linked_role"); ok {
				accountReq.LinkedRole = v.(string)
			}
			if v, ok := d.GetOk("start_datecode"); ok {
				accountReq.StartDatecode = v.(string)
			}
			accountReq.IncludeLinkedAccountSpend = hc.OptionalBool(d, "include_linked_account_spend")
			accountReq.SkipAccessChecking = hc.OptionalBool(d, "skip_access_checking")
			accountReq.UseOrgAccountInfo = hc.OptionalBool(d, "use_org_account_info")
			req = accountReq
		}

		if rb, err := json.Marshal(req); err == nil {
			tflog.Debug(ctx, fmt.Sprintf("Updating account via PATCH %s", accountUrl), map[string]interface{}{"postData": string(rb)})
		}

		err := c.PATCH(accountUrl, req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update account",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	if accountLocation == ProjectLocation && d.HasChanges("labels") {
		hasChanged++

		err := hc.PutAppLabelIDs(c, hc.FlattenAssociateLabels(d, "labels"), "account", ID)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update account labels",
				Detail:   fmt.Sprintf("Error: %v\nAccount ID: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	accountLocation := getKionAccountLocation(d)

	var accountUrl string
	switch accountLocation {
	case CacheLocation:
		accountUrl = fmt.Sprintf("/v3/account-cache/%s", ID)
	case ProjectLocation:
		fallthrough
	default:
		accountUrl = fmt.Sprintf("/v3/account/%s", ID)
	}

	err := c.DELETE(accountUrl, nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete account",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func convertCacheAccountToProjectAccount(c *hc.Client, accountCacheId, newProjectId int, startDatecode string) (int, error) {

	// The API is inconsistent and convert expects YYYYMM while other methods expect YYYY-MM
	startDatecode = strings.ReplaceAll(startDatecode, "-", "")

	resp, err := c.POST(fmt.Sprintf("/v3/account-cache/%d/convert/%d?start_datecode=%s",
		accountCacheId, newProjectId, startDatecode), nil)

	if err != nil {
		return 0, err
	}

	return resp.RecordID, nil
}

func retryConvertCacheAccountToProjectAccountForAWS(c *hc.Client, accountCacheId, projectId int, startDatecode string, retries int, delay time.Duration) (int, error) {
	var lastErr error
	for i := 0; i < retries; i++ {
		id, err := convertCacheAccountToProjectAccount(c, accountCacheId, projectId, startDatecode)
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

func convertProjectAccountToCacheAccount(c *hc.Client, accountId int) (int, error) {
	respRevert := new(hc.AccountRevertResponse)
	err := c.DeleteWithResponse(fmt.Sprintf("/v3/account/revert/%d", accountId), nil, respRevert)

	if err != nil {
		return 0, err
	}

	return respRevert.RecordID, nil
}

//
// Methods for determining whether we are placing the acount in a project or the account cache
//

const (
	CacheLocation   = "cache"
	ProjectLocation = "project"
)

func getKionAccountLocation(d *schema.ResourceData) string {
	if v, exists := d.GetOk("location"); exists {
		return v.(string)
	}

	if _, exists := d.GetOk("project_id"); exists {
		return ProjectLocation
	}
	return CacheLocation
}

// Show the account location computed attribute in the diff
func customDiffComputedAccountLocation(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	if _, exists := d.GetOk("project_id"); exists {
		d.SetNew("location", ProjectLocation)
	} else {
		d.SetNew("location", CacheLocation)
	}
	return nil
}
