package kion

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// resourceWebhooks returns a schema.Resource for managing webhooks in Kion.
func resourceWebhooks() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhooksCreate,
		ReadContext:   resourceWebhooksRead,
		UpdateContext: resourceWebhooksUpdate,
		DeleteContext: resourceWebhooksDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWebhooksImport,
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
				Description: "Headers to be included in the webhook request.",
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
				Optional:    true,
				Description: "Timeout for the webhook request in seconds.",
			},
			"use_request_headers": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to use request headers in the webhook request.",
			},
		},
	}
}

// resourceWebhooksCreate handles the creation of the webhook resource.
func resourceWebhooksCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Create a webhook object from the provided data
	webhook := hc.Webhook{
		CalloutURL:           d.Get("callout_url").(string),
		Description:          d.Get("description").(string),
		Name:                 d.Get("name").(string),
		OwnerUserGroupIDs:    *hc.FlattenIntArrayPointer(d.Get("owner_user_group_ids").(*schema.Set).List()),
		OwnerUserIDs:         *hc.FlattenIntArrayPointer(d.Get("owner_user_ids").(*schema.Set).List()),
		RequestBody:          d.Get("request_body").(string),
		RequestHeaders:       d.Get("request_headers").(string),
		RequestMethod:        d.Get("request_method").(string),
		ShouldSendSecureInfo: d.Get("should_send_secure_info").(bool),
		SkipSSL:              d.Get("skip_ssl").(bool),
		TimeoutInSeconds:     d.Get("timeout_in_seconds").(int),
		UseRequestHeaders:    d.Get("use_request_headers").(bool),
	}

	// Make a POST request to create the webhook
	resp, err := client.POST("/v3/webhook", webhook)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the resource ID using the returned webhook ID
	d.SetId(strconv.Itoa(resp.RecordID))

	// Ensure the state reflects the provided data
	return resourceWebhooksRead(ctx, d, m)
}

// resourceWebhooksRead retrieves the state of the webhook resource from the Kion API.
func resourceWebhooksRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Get the webhook ID from the resource data
	webhookID := d.Id()

	resp := new(hc.WebhookWithOwnersResponse)
	err := client.GET(fmt.Sprintf("/v3/webhook/%s", webhookID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	webhook := resp.Data.Webhook
	diags := diag.Diagnostics{}

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
	diags = append(diags, hc.SafeSet(d, "owner_user_group_ids", webhook.OwnerUserGroupIDs)...)
	diags = append(diags, hc.SafeSet(d, "owner_user_ids", webhook.OwnerUserIDs)...)

	return diags
}

// resourceWebhooksUpdate handles updating the webhook resource.
func resourceWebhooksUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Get the webhook ID from the resource data
	webhookID := d.Id()

	// Create a webhook object from the updated data
	webhook := hc.Webhook{
		CalloutURL:           d.Get("callout_url").(string),
		Description:          d.Get("description").(string),
		Name:                 d.Get("name").(string),
		RequestBody:          d.Get("request_body").(string),
		RequestHeaders:       d.Get("request_headers").(string),
		RequestMethod:        d.Get("request_method").(string),
		ShouldSendSecureInfo: d.Get("should_send_secure_info").(bool),
		SkipSSL:              d.Get("skip_ssl").(bool),
		TimeoutInSeconds:     d.Get("timeout_in_seconds").(int),
		UseRequestHeaders:    d.Get("use_request_headers").(bool),
	}

	// Make a PATCH request to update the webhook
	err := client.PATCH(fmt.Sprintf("/v3/webhook/%s", webhookID), webhook)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ensure the state reflects the updated data
	return resourceWebhooksRead(ctx, d, m)
}

// resourceWebhooksDelete handles the deletion of the webhook resource.
func resourceWebhooksDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	// Get the webhook ID from the resource data
	webhookID := d.Id()

	// Send a DELETE request to remove the webhook
	err := client.DELETE(fmt.Sprintf("/v3/webhook/%s", webhookID))
	if err != nil {
		return diag.FromErr(err)
	}

	// Remove the resource ID to indicate it has been deleted
	d.SetId("")

	return nil
}

// resourceWebhooksImport handles the import of existing webhook resources into Terraform.
func resourceWebhooksImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	webhookID := d.Id()

	// Set the webhook_id field in the resource data
	if err := d.Set("id", webhookID); err != nil {
		return nil, err
	}

	// Return the resource data for importing
	return []*schema.ResourceData{d}, nil
}
