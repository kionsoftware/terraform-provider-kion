package kion

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// Shared methods used by kion_*_account resources.
// See one of:
//   kion/resource_aws_account.go
//   kion/resource_gcp_account.go
//   kion/resource_azure_subscription_account.go

func resourceAccountRead(resource string, ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	// Handle special case for importing accounts with prefixes
	var accountLocation string
	locationChanged := false
	originalID := ID
	if strings.HasPrefix(ID, "account_id=") {
		ID = strings.TrimPrefix(ID, "account_id=")
		accountLocation = ProjectLocation
		locationChanged = true
	} else if strings.HasPrefix(ID, "account_cache_id=") {
		ID = strings.TrimPrefix(ID, "account_cache_id=")
		accountLocation = CacheLocation
		locationChanged = true
	} else if loc, exists := d.GetOk("location"); exists {
		// If location is explicitly set in state, use it
		accountLocation = loc.(string)
	} else {
		// Default to project first if no explicit location
		accountLocation = ProjectLocation
	}

	// Always set the clean ID
	d.SetId(ID)

	tflog.Debug(ctx, "Reading account", map[string]interface{}{
		"id":          ID,
		"original_id": originalID,
		"location":    accountLocation,
		"resource":    resource,
	})

	// Fetch data from account or account-cache URL
	var resp hc.MappableResponse
	var err error

	// If location was explicitly set (either via import prefix or state), only try that location
	if locationChanged || d.Get("location") != "" {
		if accountLocation == ProjectLocation {
			resp = new(hc.AccountResponse)
			err = client.GET(fmt.Sprintf("/v3/account/%s", ID), resp)
		} else {
			resp = new(hc.AccountCacheResponse)
			err = client.GET(fmt.Sprintf("/v3/account-cache/%s", ID), resp)
		}
		// Return error if we can't find it in the specified location
		if err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("unable to read account from %s (ID: %s): %v", accountLocation, ID, err))...)
		}
	} else {
		// If no explicit location, try project first then fall back to cache
		resp = new(hc.AccountResponse)
		err = client.GET(fmt.Sprintf("/v3/account/%s", ID), resp)
		if err != nil {
			resp = new(hc.AccountCacheResponse)
			err = client.GET(fmt.Sprintf("/v3/account-cache/%s", ID), resp)
			if err == nil {
				accountLocation = CacheLocation
			}
		} else {
			accountLocation = ProjectLocation
		}
	}

	// Always set the location based on where we found the account
	diags = append(diags, hc.SafeSet(d, "location", accountLocation, "Failed to set location for account")...)

	data := resp.ToMap(resource)
	for k, v := range data {
		diags = append(diags, hc.SafeSet(d, k, v, "Unable to set account field")...)
	}

	// Handle labels for project accounts
	if accountLocation == ProjectLocation {
		labelData, err := hc.ReadResourceLabels(client, "account", ID)
		if err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("unable to read account labels (ID: %s): %v", ID, err))...)
		}
		diags = append(diags, hc.SafeSet(d, "labels", labelData, "Failed to set account labels")...)
	}

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	var hasChanged bool
	var accountLocation string

	// Handle project ID changes
	if d.HasChange("project_id") {
		hasChanged = true
		oldId, newId := d.GetChange("project_id")
		oldProjectId := oldId.(int)
		newProjectId := newId.(int)

		diags = append(diags, handleAccountConversion(ctx, d, client, oldProjectId, newProjectId)...)
		if diags.HasError() {
			return diags
		}

		accountLocation = getKionAccountLocation(d)
	} else {
		accountLocation = getKionAccountLocation(d)
	}

	// Handle labels for project accounts
	if accountLocation == ProjectLocation && d.HasChange("labels") {
		hasChanged = true
		if err := hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "account", ID); err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("unable to update account labels (ID: %s): %v", ID, err))...)
		}
	}

	if hasChanged {
		diags = append(diags, hc.SafeSet(d, "last_updated", time.Now().Format(time.RFC850), "Failed to set last_updated")...)
	}

	return diags
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()
	accountLocation := getKionAccountLocation(d)

	tflog.Debug(ctx, "Deleting account", map[string]interface{}{
		"id":       ID,
		"location": accountLocation,
	})

	var accountUrl string
	switch accountLocation {
	case CacheLocation:
		accountUrl = fmt.Sprintf("/v3/account-cache/%s", ID)
	case ProjectLocation:
		fallthrough
	default:
		accountUrl = fmt.Sprintf("/v3/account/%s", ID)
	}

	if err := client.DELETE(accountUrl, nil); err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("failed to delete account (ID: %s): %v", ID, err))...)
	}

	d.SetId("")
	return diags
}

func convertCacheAccountToProjectAccount(client *hc.Client, accountCacheId, projectId int, startDatecode string) (int, error) {
	startDatecode = strings.ReplaceAll(startDatecode, "-", "")

	resp, err := client.POST(fmt.Sprintf("/v3/account-cache/%d/convert/%d?start_datecode=%s",
		accountCacheId, projectId, startDatecode), nil)

	if err != nil {
		return 0, fmt.Errorf("failed to convert cache account to project account: %v", err)
	}
	return resp.RecordID, nil
}

func convertProjectAccountToCacheAccount(client *hc.Client, accountId int) (int, error) {
	respRevert := new(hc.AccountRevertResponse)
	err := client.DeleteWithResponse(fmt.Sprintf("/v3/account/revert/%d", accountId), nil, respRevert)

	if err != nil {
		return 0, fmt.Errorf("failed to convert project account to cache account: %v", err)
	}
	return respRevert.RecordID, nil
}

func handleAccountConversion(ctx context.Context, d *schema.ResourceData, client *hc.Client, oldProjectId, newProjectId int) diag.Diagnostics {
	var diags diag.Diagnostics
	ID := d.Id()

	if oldProjectId == 0 && newProjectId != 0 {
		// Converting from cache to project
		accountCacheId, err := strconv.Atoi(ID)
		if err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("invalid account cache id: %v", err))...)
		}

		tflog.Debug(ctx, "Converting from cached account to project account", map[string]interface{}{
			"account_cache_id": accountCacheId,
			"project_id":       newProjectId,
		})

		newId, err := convertCacheAccountToProjectAccount(client, accountCacheId, newProjectId, d.Get("start_datecode").(string))
		if err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("failed to convert cache account to project: %v", err))...)
		}

		d.SetId(fmt.Sprintf("%d", newId))
		diags = append(diags, hc.SafeSet(d, "location", ProjectLocation, "Failed to set location")...)

	} else if oldProjectId != 0 && newProjectId == 0 {
		// Converting from project to cache
		accountId, err := strconv.Atoi(ID)
		if err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("invalid account id: %v", err))...)
		}

		tflog.Debug(ctx, "Converting from project account to cached account", map[string]interface{}{
			"account_id": accountId,
		})

		newId, err := convertProjectAccountToCacheAccount(client, accountId)
		if err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("failed to convert project account to cache: %v", err))...)
		}

		d.SetId(fmt.Sprintf("%d", newId))
		diags = append(diags, hc.SafeSet(d, "location", CacheLocation, "Failed to set location")...)
	}

	return diags
}

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
	// Get the project_id
	projectId := d.Get("project_id").(int)

	// If we're creating a new resource and project_id is set,
	// we need to wait for the project to exist
	if d.Id() == "" {
		// During creation, always mark location as computed
		// This allows the value to be determined after project creation
		return d.SetNewComputed("location")
	}

	// For existing resources, we can compute the location directly
	if projectId != 0 {
		if err := d.SetNew("location", ProjectLocation); err != nil {
			return fmt.Errorf("failed to set location to project: %v", err)
		}
		return nil
	}

	if err := d.SetNew("location", CacheLocation); err != nil {
		return fmt.Errorf("failed to set location to cache: %v", err)
	}
	return nil
}
