package cloudtamerio

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/cloudtamerio/internal/ctclient"
)

func dataSourceComplianceCheck() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceComplianceCheckRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"regex": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"azure_policy_id": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"body": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cloud_provider_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"compliance_check_type_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_by_user_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ct_managed": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"frequency_minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"frequency_type_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"is_all_regions": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"is_auto_archived": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"last_scan_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"regions": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"severity_type_id": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceComplianceCheckRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	resp := new(hc.ComplianceCheckListResponse)
	err := c.GET("/v3/compliance/check", resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read ComplianceCheck",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	f := hc.NewFilterable(d)

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data {
		data := make(map[string]interface{})
		if item.AzurePolicyID != nil {
			data["azure_policy_id"] = item.AzurePolicyID
		}
		data["body"] = item.Body
		data["cloud_provider_id"] = item.CloudProviderID
		data["compliance_check_type_id"] = item.ComplianceCheckTypeID
		data["created_at"] = item.CreatedAt
		data["created_by_user_id"] = item.CreatedByUserID
		data["ct_managed"] = item.CtManaged
		data["description"] = item.Description
		data["frequency_minutes"] = item.FrequencyMinutes
		data["frequency_type_id"] = item.FrequencyTypeID
		data["id"] = item.ID
		data["is_all_regions"] = item.IsAllRegions
		data["is_auto_archived"] = item.IsAutoArchived
		data["last_scan_id"] = item.LastScanID
		data["name"] = item.Name
		data["regions"] = hc.FilterStringArray(item.Regions)
		if item.SeverityTypeID != nil {
			data["severity_type_id"] = item.SeverityTypeID
		}

		match, err := f.Match(data)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to filter ComplianceCheck",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "filter"),
			})
			return diags
		} else if !match {
			continue
		}

		arr = append(arr, data)
	}

	if err := d.Set("list", arr); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read ComplianceCheck",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	// Always run.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
