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

func dataSourceAzureArmTemplate() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAzureArmTemplateRead,
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
						"ct_managed": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"deployment_mode": {
							Type:     schema.TypeInt,
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
						"resource_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_group_region_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"template": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"template_parameters": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAzureArmTemplateRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	resp := new(hc.AzureARMTemplateListResponse)
	err := client.GET("/v3/azure-arm-template", resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Azure ARM Template",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	f := hc.NewFilterable(d)

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data {
		data := make(map[string]interface{})
		data["ct_managed"] = item.AzureArmTemplate.CtManaged
		data["deployment_mode"] = item.AzureArmTemplate.DeploymentMode
		data["description"] = item.AzureArmTemplate.Description
		data["id"] = item.AzureArmTemplate.ID
		data["name"] = item.AzureArmTemplate.Name
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
		data["resource_group_name"] = item.AzureArmTemplate.ResourceGroupName
		data["resource_group_region_id"] = item.AzureArmTemplate.ResourceGroupRegionID
		data["template"] = item.AzureArmTemplate.Template
		data["template_parameters"] = item.AzureArmTemplate.TemplateParameters
		data["version"] = item.AzureArmTemplate.Version

		match, err := f.Match(data)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to filter Azure ARM Template",
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
			Summary:  "Unable to read Azure ARM Template",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	// Always run.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
