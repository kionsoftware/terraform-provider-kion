package kion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// dataSourceWebhook returns a schema.Resource for reading a specific webhook by ID from Kion.
func dataSourceWebhook() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWebhookRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the webhook to retrieve.",
			},
			"callout_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL to which the webhook will send requests.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the webhook.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the webhook.",
			},
			"request_body": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The request body to be sent with the webhook.",
			},
			"request_headers": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Headers to be included in the webhook request.",
			},
			"request_method": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "HTTP method to be used for the webhook (GET, POST, etc.).",
			},
			"should_send_secure_info": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the webhook should send secure information.",
			},
			"skip_ssl": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether to skip SSL verification.",
			},
			"timeout_in_seconds": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Timeout for the webhook request in seconds.",
			},
			"use_request_headers": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether to use request headers in the webhook request.",
			},
			"owner_user_group_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Set of user group IDs that own the webhook.",
			},
			"owner_user_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Set of user IDs that own the webhook.",
			},
		},
	}
}

// dataSourceWebhookRead retrieves data for a specific webhook by ID.
func dataSourceWebhookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	webhookID := d.Get("id").(string)

	var diags diag.Diagnostics

	// Fetch the specific webhook by ID
	resp := new(hc.WebhookWithOwnersResponse)
	err := client.GET(fmt.Sprintf("/v3/webhook/%s", webhookID), resp)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read webhook by ID: %v", err))
	}

	webhook := resp.Data.Webhook

	d.SetId(webhookID)

	// Use SafeSet to set each field in the schema from the webhook response
	diags = append(diags, hc.SafeSet(d, "callout_url", webhook.CalloutURL, "Failed to set callout_url")...)
	diags = append(diags, hc.SafeSet(d, "description", webhook.Description, "Failed to set description")...)
	diags = append(diags, hc.SafeSet(d, "name", webhook.Name, "Failed to set name")...)
	diags = append(diags, hc.SafeSet(d, "request_body", webhook.RequestBody, "Failed to set request_body")...)
	diags = append(diags, hc.SafeSet(d, "request_headers", webhook.RequestHeaders, "Failed to set request_headers")...)
	diags = append(diags, hc.SafeSet(d, "request_method", webhook.RequestMethod, "Failed to set request_method")...)
	diags = append(diags, hc.SafeSet(d, "should_send_secure_info", webhook.ShouldSendSecureInfo, "Failed to set should_send_secure_info")...)
	diags = append(diags, hc.SafeSet(d, "skip_ssl", webhook.SkipSSL, "Failed to set skip_ssl")...)
	diags = append(diags, hc.SafeSet(d, "timeout_in_seconds", webhook.TimeoutInSeconds, "Failed to set timeout_in_seconds")...)
	diags = append(diags, hc.SafeSet(d, "use_request_headers", webhook.UseRequestHeaders, "Failed to set use_request_headers")...)

	// Set owner user group IDs
	ownerUserGroupIDs := extractOwnerGroupIDs(resp.Data.OwnerUserGroups)
	diags = append(diags, hc.SafeSet(d, "owner_user_group_ids", ownerUserGroupIDs, "Failed to set owner_user_group_ids")...)

	// Set owner user IDs
	ownerUserIDs := extractOwnerUserIDs(resp.Data.OwnerUsers)
	diags = append(diags, hc.SafeSet(d, "owner_user_ids", ownerUserIDs, "Failed to set owner_user_ids")...)

	return diags
}

// Helper functions to extract IDs from owner groups and users
func extractOwnerGroupIDs(groups []hc.OwnerUserGroup) []int {
	ids := make([]int, len(groups))
	for i, group := range groups {
		ids[i] = group.ID
	}
	return ids
}

func extractOwnerUserIDs(users []hc.OwnerUser) []int {
	ids := make([]int, len(users))
	for i, user := range users {
		ids[i] = user.ID
	}
	return ids
}
