package kion

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceProjectEnforcement() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProjectEnforcementRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The field name whose values you wish to filter by.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"values": {
							Description: "The values of the field name you specified.",
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"regex": {
							Description: "Dictates if the values provided should be treated as regular expressions.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"enforcements": {
				Description: "List of project enforcement policies configured in the system.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"timeframe": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"spend_option": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"amount_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"threshold_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cloud_rule_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"notification_frequency": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"project_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ou_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"user_group_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
						"user_ids": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
						"triggered": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceProjectEnforcementRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	resp := new(hc.ProjectEnforcementResponse)
	err := client.GET("/v3/project/{id}/enforcement", resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Project Enforcement",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	f := hc.NewFilterable(d)

	enforcements := make([]map[string]interface{}, 0)
	for _, item := range resp.Data {
		data := make(map[string]interface{})
		data["id"] = item.ID
		data["description"] = item.Description
		data["timeframe"] = item.Timeframe
		data["spend_option"] = item.SpendOption
		data["amount_type"] = item.AmountType
		data["service_id"] = item.Service.ID
		data["threshold_type"] = item.ThresholdType
		data["threshold"] = item.Threshold
		data["cloud_rule_id"] = item.CloudRule.ID
		data["notification_frequency"] = item.NotificationFrequency
		data["project_id"] = item.ProjectID
		data["ou_id"] = item.OUID
		data["enabled"] = item.Enabled
		data["user_group_ids"] = item.UserGroupIds
		data["user_ids"] = item.UserIds
		data["triggered"] = item.Triggered

		match, err := f.Match(data)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to filter Project Enforcement",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
			return diags
		} else if !match {
			continue
		}

		enforcements = append(enforcements, data)
	}

	if err := d.Set("enforcements", enforcements); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set Project Enforcement data",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	// Set the ID of the datasource to a unique value, which is the current timestamp
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
