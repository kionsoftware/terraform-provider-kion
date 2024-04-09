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

func dataSourceAzurePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAzurePolicyRead,
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
						"azure_managed_policy_def_id": {
							Type:     schema.TypeString,
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
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner_user_groups": {
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
							Type:     schema.TypeList,
							Computed: true,
						},
						"owner_users": {
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
							Type:     schema.TypeList,
							Computed: true,
						},
						"parameters": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAzurePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	k := m.(*hc.Client)

	resp := new(hc.AzurePolicyListResponse)
	err := k.GET("/v3/azure-policy", resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read AzurePolicy",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	f := hc.NewFilterable(d)

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data {
		data := make(map[string]interface{})
		data["azure_managed_policy_def_id"] = item.AzurePolicy.AzureManagedPolicyDefID
		data["ct_managed"] = item.AzurePolicy.CtManaged
		data["description"] = item.AzurePolicy.Description
		data["id"] = item.AzurePolicy.ID
		data["name"] = item.AzurePolicy.Name
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
		data["parameters"] = item.AzurePolicy.Parameters
		data["policy"] = item.AzurePolicy.Policy

		match, err := f.Match(data)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to filter AzurePolicy",
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
			Summary:  "Unable to read AzurePolicy",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	// Always run.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
