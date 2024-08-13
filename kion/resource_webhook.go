package kion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// resourceWebhook returns a schema.Resource for managing webhooks in Kion.
func resourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWebhookImport,
		},
		Schema: map[string]*schema.Schema{
			"callout_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL to which the webhook will send requests.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the webhook.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the webhook.",
			},
			"owner_user_group_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Set of user group IDs that own the webhook.",
			},
			"owner_user_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Set of user IDs that own the webhook.",
			},
			"request_body": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The request body to be sent with the webhook.",
			},
			"request_headers": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "HTTP headers to use when the webhook is triggered",
			},
			"request_method": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "HTTP method to be used for the webhook (GET, POST, etc.).",
			},
			"should_send_secure_info": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether the webhook should send secure information.",
			},
			"skip_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to skip SSL verification.",
			},
			"timeout_in_seconds": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The number of seconds the application will wait before considering the webhook 'timed out'",
			},
			"use_request_headers": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to use request headers in the webhook request.",
			},
		},
		// Set the CustomizeDiff function
		CustomizeDiff: validateOwnerFields,
	}
}

// populateWebhook creates a Webhook object from the provided schema.ResourceData.
func populateWebhook(d *schema.ResourceData) (hc.Webhook, error) {
	webhook := hc.Webhook{
		CalloutURL:           d.Get("callout_url").(string),
		Description:          d.Get("description").(string),
		Name:                 d.Get("name").(string),
		ShouldSendSecureInfo: d.Get("should_send_secure_info").(bool),
		SkipSSL:              d.Get("skip_ssl").(bool),
		TimeoutInSeconds:     d.Get("timeout_in_seconds").(int),
		UseRequestHeaders:    d.Get("use_request_headers").(bool),
	}

	var err error

	// Normalize request_body
	if requestBody, ok := d.Get("request_body").(string); ok && requestBody != "" {
		webhook.RequestBody, err = normalizeJSONString(requestBody)
		if err != nil {
			return hc.Webhook{}, fmt.Errorf("error normalizing request_body: %w", err)
		}
	}

	// Normalize request_headers
	if requestHeaders, ok := d.Get("request_headers").(string); ok && requestHeaders != "" {
		webhook.RequestHeaders, err = normalizeJSONString(requestHeaders)
		if err != nil {
			return hc.Webhook{}, fmt.Errorf("error normalizing request_headers: %w", err)
		}
	}

	// Convert owner user IDs and handle potential errors
	webhook.OwnerUserIDs, err = hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_ids").(*schema.Set).List())
	if err != nil {
		return hc.Webhook{}, fmt.Errorf("error converting owner_user_ids: %w", err)
	}

	// Convert owner user group IDs and handle potential errors
	webhook.OwnerUserGroupIDs, err = hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_group_ids").(*schema.Set).List())
	if err != nil {
		return hc.Webhook{}, fmt.Errorf("error converting owner_user_group_ids: %w", err)
	}

	return webhook, nil
}

// validateOwnerFields checks if at least one of owner_user_ids or owner_user_group_ids is specified
func validateOwnerFields(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	err := hc.AtLeastOneFieldPresent(map[string]interface{}{
		"owner_user_ids":       diff.Get("owner_user_ids").(*schema.Set),
		"owner_user_group_ids": diff.Get("owner_user_group_ids").(*schema.Set),
	})

	if err != nil {
		return err
	}

	return nil
}

// resourceWebhookCreate handles the creation of the webhook resource.
func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Populate webhook and handle any errors
	webhook, err := populateWebhook(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Validate that at least one of owner_user_ids or owner_user_group_ids is set
	if len(webhook.OwnerUserIDs) == 0 && len(webhook.OwnerUserGroupIDs) == 0 {
		return diag.Errorf("at least one of owner_user_ids or owner_user_group_ids must be set")
	}

	// Create webhook
	resp, err := client.POST("/v3/webhook", webhook)
	if diags := hc.HandleError(err); diags != nil {
		return diags
	}

	// Set the ID of the resource
	d.SetId(strconv.Itoa(resp.RecordID))

	return resourceWebhookRead(ctx, d, m)
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	webhookID := d.Id()

	// Prepare response struct
	resp := new(hc.WebhookWithOwnersResponse)

	// Perform GET request to fetch webhook data
	err := client.GET(fmt.Sprintf("/v3/webhook/%s", webhookID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Webhook",
			Detail:   fmt.Sprintf("Error: %v\nWebhook ID: %v", err.Error(), webhookID),
		})
		return diags
	}

	webhook := resp.Data.Webhook

	// Normalize JSON fields before setting them in state
	normalizedRequestBody, err := normalizeJSONString(webhook.RequestBody)
	if err != nil {
		return diag.Errorf("failed to normalize request_body: %v", err)
	}

	normalizedRequestHeaders, err := normalizeJSONString(webhook.RequestHeaders)
	if err != nil {
		return diag.Errorf("failed to normalize request_headers: %v", err)
	}

	// Set fields based on the API response
	diags = append(diags, hc.SafeSet(d, "callout_url", webhook.CalloutURL, "Failed to set callout_url")...)
	diags = append(diags, hc.SafeSet(d, "description", webhook.Description, "Failed to set description")...)
	diags = append(diags, hc.SafeSet(d, "name", webhook.Name, "Failed to set name")...)
	diags = append(diags, hc.SafeSet(d, "request_body", normalizedRequestBody, "Failed to set request_body")...)
	diags = append(diags, hc.SafeSet(d, "request_headers", normalizedRequestHeaders, "Failed to set request_headers")...)
	diags = append(diags, hc.SafeSet(d, "request_method", webhook.RequestMethod, "Failed to set request_method")...)
	diags = append(diags, hc.SafeSet(d, "should_send_secure_info", webhook.ShouldSendSecureInfo, "Failed to set should_send_secure_info")...)
	diags = append(diags, hc.SafeSet(d, "skip_ssl", webhook.SkipSSL, "Failed to set skip_ssl")...)
	diags = append(diags, hc.SafeSet(d, "timeout_in_seconds", webhook.TimeoutInSeconds, "Failed to set timeout_in_seconds")...)
	diags = append(diags, hc.SafeSet(d, "use_request_headers", webhook.UseRequestHeaders, "Failed to set use_request_headers")...)

	// Extract owner user group IDs
	ownerUserGroupIDs := make([]int, len(resp.Data.OwnerUserGroups))
	for i, group := range resp.Data.OwnerUserGroups {
		ownerUserGroupIDs[i] = group.ID
	}
	diags = append(diags, hc.SafeSet(d, "owner_user_group_ids", ownerUserGroupIDs, "Failed to set owner_user_group_ids")...)

	// Extract owner user IDs
	ownerUserIDs := make([]int, len(resp.Data.OwnerUsers))
	for i, user := range resp.Data.OwnerUsers {
		ownerUserIDs[i] = user.ID
	}
	diags = append(diags, hc.SafeSet(d, "owner_user_ids", ownerUserIDs, "Failed to set owner_user_ids")...)

	return diags
}

// resourceWebhookUpdate handles updating the webhook resource.
func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	webhookID := d.Id()

	// Populate webhook and handle any errors
	webhook, err := populateWebhook(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Check if fields have changed and update accordingly
	if d.HasChange("name") || d.HasChange("callout_url") || d.HasChange("description") ||
		d.HasChange("request_body") || d.HasChange("request_headers") || d.HasChange("request_method") ||
		d.HasChange("should_send_secure_info") || d.HasChange("skip_ssl") || d.HasChange("timeout_in_seconds") ||
		d.HasChange("use_request_headers") {

		// Normalize JSON fields before sending the update
		webhook.RequestBody, err = normalizeJSONString(webhook.RequestBody)
		if err != nil {
			return diag.FromErr(err)
		}

		webhook.RequestHeaders, err = normalizeJSONString(webhook.RequestHeaders)
		if err != nil {
			return diag.FromErr(err)
		}

		// Send PATCH request to update the webhook
		err := client.PATCH(fmt.Sprintf("/v3/webhook/%s", webhookID), webhook)
		if diags := hc.HandleError(err); diags != nil {
			return diags
		}
	}

	if d.HasChange("owner_user_group_ids") || d.HasChange("owner_user_ids") {
		// Get previous and new owner IDs
		prevUserIDs, prevGroupIDs, err := hc.GetPreviousUserAndGroupIds(d)
		if err != nil {
			return diag.FromErr(err)
		}

		newUserIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_ids").(*schema.Set).List())
		if err != nil {
			return diag.FromErr(err)
		}
		newGroupIDs, err := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_group_ids").(*schema.Set).List())
		if err != nil {
			return diag.FromErr(err)
		}

		// Find differences
		removedUserIDs := hc.FindDifferences(prevUserIDs, newUserIDs)
		removedGroupIDs := hc.FindDifferences(prevGroupIDs, newGroupIDs)
		addedUserIDs := hc.FindDifferences(newUserIDs, prevUserIDs)
		addedGroupIDs := hc.FindDifferences(newGroupIDs, prevGroupIDs)

		// Remove old owners
		if len(removedUserIDs) > 0 || len(removedGroupIDs) > 0 {
			removalData := map[string]interface{}{
				"owner_user_ids":       removedUserIDs,
				"owner_user_group_ids": removedGroupIDs,
			}
			err := client.DELETE(fmt.Sprintf("/v3/webhook/%s/owner", webhookID), removalData)
			if diags := hc.HandleError(err); diags != nil {
				return diags
			}
		}

		// Add new owners
		if len(addedUserIDs) > 0 || len(addedGroupIDs) > 0 {
			additionData := map[string]interface{}{
				"owner_user_ids":       addedUserIDs,
				"owner_user_group_ids": addedGroupIDs,
			}
			_, err := client.POST(fmt.Sprintf("/v3/webhook/%s/owner", webhookID), additionData)
			if diags := hc.HandleError(err); diags != nil {
				return diags
			}
		}
	}

	return resourceWebhookRead(ctx, d, m)
}

// resourceWebhookDelete handles the deletion of the webhook resource.
func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	webhookID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/webhook/%s", webhookID), nil)
	if diags := hc.HandleError(err); diags != nil {
		return diags
	}

	d.SetId("")

	return nil
}

// resourceWebhookImport handles the import of existing webhook resources into Terraform.
func resourceWebhookImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	webhookID := d.Id()

	if err := d.Set("id", webhookID); err != nil {
		return nil, fmt.Errorf("failed to set webhook ID: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

// normalizeJSONString minifies JSON string by removing whitespace and preventing terraform changes due to whitespace changes.
func normalizeJSONString(input string) (string, error) {
	var compacted bytes.Buffer
	err := json.Compact(&compacted, []byte(input))
	if err != nil {
		return "", err
	}
	return compacted.String(), nil
}
