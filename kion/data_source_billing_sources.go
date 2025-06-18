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

func dataSourceBillingSources() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceBillingSourcesRead,
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
			"list": {
				Description: "This is where Kion makes the discovered data available as a list of resources.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the billing source.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the billing source.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The type of the billing source (aws, azure, gcp, or oci).",
						},
						"account_creation": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "When true, Kion is able to create accounts on this payer.",
						},
						"use_focus_reports": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if billing source is configured to read FOCUS reports.",
						},
						"use_proprietary_reports": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if billing source is configured to read proprietary billing reports from the CSP.",
						},
					},
				},
			},
		},
	}
}

func dataSourceBillingSourcesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	resp := new(hc.BillingSourceListResponse)
	err := client.GET("/v4/billing-source", resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read billing sources",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	f := hc.NewFilterable(d)
	arr := make([]map[string]interface{}, 0)

	for _, item := range resp.Data.Items {
		// Determine the billing source type and name
		var bsType string
		var bsName string

		if item.AWSPayer != nil {
			bsType = "aws"
			// Extract name from AWS payer
			if awsPayerData, ok := (*item.AWSPayer).(map[string]interface{}); ok {
				bsName = hc.GetStringFromInterface(awsPayerData["name"])
			}
		} else if item.AzurePayer != nil {
			bsType = "azure"
			// Extract name from Azure payer
			if azurePayerData, ok := (*item.AzurePayer).(map[string]interface{}); ok {
				bsName = hc.GetStringFromInterface(azurePayerData["name"])
			}
		} else if item.GCPPayer != nil {
			bsType = "gcp"
			// Extract name from GCP payer
			if item.GCPPayer.GCPBillingAccount.Name != "" {
				bsName = item.GCPPayer.GCPBillingAccount.Name
			}
		} else if item.OCIPayer != nil {
			bsType = "oci"
			// Extract name from OCI payer
			if item.OCIPayer.Name != "" {
				bsName = item.OCIPayer.Name
			}
		}

		data := map[string]interface{}{
			"id":                      int(item.ID),
			"name":                    bsName,
			"type":                    bsType,
			"account_creation":        item.AccountCreation,
			"use_focus_reports":       item.UseFocusReports,
			"use_proprietary_reports": item.UseProprietaryReports,
		}

		match, err := f.Match(data)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to filter billing sources",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
			continue
		} else if !match {
			continue
		}

		arr = append(arr, data)
	}

	if err := d.Set("list", arr); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set list",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

