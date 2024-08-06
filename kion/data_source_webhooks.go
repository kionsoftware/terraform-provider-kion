package kion

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// dataSourceWebhooks returns a schema.Resource for reading webhooks from Kion.
func dataSourceWebhooks() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWebhooksRead,
		Schema: map[string]*schema.Schema{
			"webhooks": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of webhooks.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"callout_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner_user_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
						"owner_user_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
						"request_body": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"request_headers": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"request_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"should_send_secure_info": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"skip_ssl": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"timeout_in_seconds": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"use_request_headers": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// dataSourceWebhooksRead reads the list of webhooks from the API and sets the data in Terraform state.
func dataSourceWebhooksRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	resp := new(hc.WebhookListResponse)
	err := client.GET("/v3/webhook", resp)
	if err != nil {
		return diag.FromErr(err)
	}

	webhooks := make([]map[string]interface{}, 0, len(resp.Data))
	for _, webhook := range resp.Data {
		data := map[string]interface{}{
			"id":                      webhook.ID,
			"callout_url":             webhook.CalloutURL,
			"description":             webhook.Description,
			"name":                    webhook.Name,
			"owner_user_group_ids":    webhook.OwnerUserGroupIDs,
			"owner_user_ids":          webhook.OwnerUserIDs,
			"request_body":            webhook.RequestBody,
			"request_headers":         webhook.RequestHeaders,
			"request_method":          webhook.RequestMethod,
			"should_send_secure_info": webhook.ShouldSendSecureInfo,
			"skip_ssl":                webhook.SkipSSL,
			"timeout_in_seconds":      webhook.TimeoutInSeconds,
			"use_request_headers":     webhook.UseRequestHeaders,
		}
		webhooks = append(webhooks, data)
	}

	if err := d.Set("webhooks", webhooks); err != nil {
		return diag.FromErr(err)
	}

	// Set the ID of the data source to a unique value
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
