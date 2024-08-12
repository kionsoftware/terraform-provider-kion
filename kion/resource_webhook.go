package kion

import (
	"context"
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
func populateWebhook(d *schema.ResourceData) hc.Webhook {
	webhook := hc.Webhook{
		CalloutURL:           d.Get("callout_url").(string),
		Description:          d.Get("description").(string),
		Name:                 d.Get("name").(string),
		RequestBody:          d.Get("request_body").(string),
		RequestHeaders:       d.Get("request_headers").(string),
		ShouldSendSecureInfo: d.Get("should_send_secure_info").(bool),
		SkipSSL:              d.Get("skip_ssl").(bool),
		TimeoutInSeconds:     d.Get("timeout_in_seconds").(int),
		UseRequestHeaders:    d.Get("use_request_headers").(bool),
		OwnerUserIDs:         hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_ids").(*schema.Set).List()),
		OwnerUserGroupIDs:    hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_group_ids").(*schema.Set).List()),
	}

	return webhook
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

	webhook := populateWebhook(d)

	// Validate that at least one of owner_user_ids or owner_user_group_ids is set
	if len(webhook.OwnerUserIDs) == 0 && len(webhook.OwnerUserGroupIDs) == 0 {
		return diag.Errorf("at least one of owner_user_ids or owner_user_group_ids must be set")
	}

	// Create webhook
	resp, err := client.POST("/v3/webhook", webhook)
	if diags := hc.HandleError(err); diags != nil {
		return diags
	}

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

	// Set fields based on the API response
	diags = append(diags, hc.SafeSet(d, "callout_url", webhook.CalloutURL)...)
	diags = append(diags, hc.SafeSet(d, "description", webhook.Description)...)
	diags = append(diags, hc.SafeSet(d, "name", webhook.Name)...)
	diags = append(diags, hc.SafeSet(d, "request_body", webhook.RequestBody)...)
	diags = append(diags, hc.SafeSet(d, "request_headers", webhook.RequestHeaders)...)
	diags = append(diags, hc.SafeSet(d, "request_method", webhook.RequestMethod)...)
	diags = append(diags, hc.SafeSet(d, "should_send_secure_info", webhook.ShouldSendSecureInfo)...)
	diags = append(diags, hc.SafeSet(d, "skip_ssl", webhook.SkipSSL)...)
	diags = append(diags, hc.SafeSet(d, "timeout_in_seconds", webhook.TimeoutInSeconds)...)
	diags = append(diags, hc.SafeSet(d, "use_request_headers", webhook.UseRequestHeaders)...)

	// Extract owner user group IDs
	ownerUserGroupIDs := make([]int, len(resp.Data.OwnerUserGroups))
	for i, group := range resp.Data.OwnerUserGroups {
		ownerUserGroupIDs[i] = group.ID
	}
	diags = append(diags, hc.SafeSet(d, "owner_user_group_ids", ownerUserGroupIDs)...)

	// Extract owner user IDs
	ownerUserIDs := make([]int, len(resp.Data.OwnerUsers))
	for i, user := range resp.Data.OwnerUsers {
		ownerUserIDs[i] = user.ID
	}
	diags = append(diags, hc.SafeSet(d, "owner_user_ids", ownerUserIDs)...)

	return diags
}

// resourceWebhookUpdate handles updating the webhook resource.
func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	webhookID := d.Id()
	webhook := populateWebhook(d)

	// Check if fields have changed and update accordingly
	if d.HasChange("name") || d.HasChange("callout_url") || d.HasChange("description") ||
		d.HasChange("request_body") || d.HasChange("request_headers") || d.HasChange("request_method") ||
		d.HasChange("should_send_secure_info") || d.HasChange("skip_ssl") || d.HasChange("timeout_in_seconds") ||
		d.HasChange("use_request_headers") {
		err := client.PATCH(fmt.Sprintf("/v3/webhook/%s", webhookID), webhook)
		if diags := hc.HandleError(err); diags != nil {
			return diags
		}
	}

	if d.HasChange("owner_user_group_ids") || d.HasChange("owner_user_ids") {
		// Get previous and new owner IDs
		prevUserIDs, prevGroupIDs := hc.GetPreviousUserAndGroupIds(d)
		newUserIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_ids").(*schema.Set).List())
		newGroupIDs := hc.ConvertInterfaceSliceToIntSlice(d.Get("owner_user_group_ids").(*schema.Set).List())

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
