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

func dataSourceCustomVariableOverride() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCustomVariableOverrideRead,
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
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						// All following fields are in place of an ID since overrides do not have an ID.
						"entity_type": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"entity_id": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"custom_variable_id": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCustomVariableOverrideRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	var arr []map[string]interface{}

	ouOverrides, errDiags := getAllOUOverrides(d, client)
	if errDiags != nil {
		return errDiags
	}
	arr = append(arr, ouOverrides...)

	projectOverrides, errDiags := getAllProjectOverrides(d, client)
	if errDiags != nil {
		return errDiags
	}
	arr = append(arr, projectOverrides...)

	accountOverrides, errDiags := getAllAccountOverrides(d, client)
	if errDiags != nil {
		return errDiags
	}
	arr = append(arr, accountOverrides...)

	accountCacheOverrides, errDiags := getAllAccountCacheOverrides(d, client)
	if errDiags != nil {
		return errDiags
	}
	arr = append(arr, accountCacheOverrides...)

	if err := d.Set("list", arr); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Custom Variables",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	// Always run.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func getAllOUOverrides(d *schema.ResourceData, client *hc.Client) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var ous hc.OUListResponse
	err := client.GET("/v3/ou", &ous)
	if err != nil {
		return nil, diag.Errorf("Error getting OUs: %v", err)
	}

	f := hc.NewFilterable(d)
	var arr []map[string]interface{}
	for _, ou := range ous.Data {
		var overrides hc.CustomVariableOverrideListResponse
		err := client.GET(fmt.Sprintf("/v3/ou/%d/custom-variable?count=999999", ou.ID), &overrides)
		if err != nil {
			return nil, diag.Errorf("Error getting OU overrides: %v", err)
		}

		for _, override := range overrides.Data.Items {
			if override.Override.Value == nil {
				continue
			}

			cvValueStr, err := hc.PackCvValueIntoJsonStr(override.Override.Value)
			if err != nil {
				return nil, diag.Errorf("Error packing CV value: %v", err)
			}

			data := map[string]interface{}{
				"value":              cvValueStr,
				"entity_type":        "ou",
				"entity_id":          fmt.Sprintf("%d", ou.ID),
				"custom_variable_id": fmt.Sprintf("%d", override.CustomVariableID),
			}

			match, err := f.Match(data)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to filter Custom Variables",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "filter"),
				})
				return nil, diags
			} else if !match {
				continue
			}

			arr = append(arr, data)
		}
	}

	return arr, nil
}

func getAllProjectOverrides(d *schema.ResourceData, client *hc.Client) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var projects hc.ProjectListResponse
	err := client.GET("/v3/project", &projects)
	if err != nil {
		return nil, diag.Errorf("Error getting projects: %v", err)
	}

	f := hc.NewFilterable(d)
	var arr []map[string]interface{}
	for _, project := range projects.Data {
		var overrides hc.CustomVariableOverrideListResponse
		err := client.GET(fmt.Sprintf("/v3/project/%d/custom-variable?count=999999", project.ID), &overrides)
		if err != nil {
			return nil, diag.Errorf("Error getting project overrides: %v", err)
		}

		for _, override := range overrides.Data.Items {
			if override.Override.Value == nil {
				continue
			}

			cvValueStr, err := hc.PackCvValueIntoJsonStr(override.Override.Value)
			if err != nil {
				return nil, diag.Errorf("Error packing CV value: %v", err)
			}

			data := map[string]interface{}{
				"value":              cvValueStr,
				"entity_type":        "project",
				"entity_id":          fmt.Sprintf("%d", project.ID),
				"custom_variable_id": fmt.Sprintf("%d", override.CustomVariableID),
			}

			match, err := f.Match(data)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to filter Custom Variables",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "filter"),
				})
				return nil, diags
			}
			if !match {
				continue
			}

			arr = append(arr, data)
		}
	}

	return arr, nil
}

func getAllAccountOverrides(d *schema.ResourceData, client *hc.Client) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var accounts hc.AccountListResponse
	err := client.GET("/v3/account", &accounts)
	if err != nil {
		return nil, diag.Errorf("Error getting accounts: %v", err)
	}

	f := hc.NewFilterable(d)
	var arr []map[string]interface{}
	for _, account := range accounts.Data {
		var overrides hc.CustomVariableOverrideListResponse
		err := client.GET(fmt.Sprintf("/v3/account/%d/custom-variable?count=999999", account.ID), &overrides)
		if err != nil {
			return nil, diag.Errorf("Error getting account overrides: %v", err)
		}

		for _, override := range overrides.Data.Items {
			if override.Override.Value == nil {
				continue
			}

			cvValueStr, err := hc.PackCvValueIntoJsonStr(override.Override.Value)
			if err != nil {
				return nil, diag.Errorf("Error packing CV value: %v", err)
			}

			data := map[string]interface{}{
				"value":              cvValueStr,
				"entity_type":        "account",
				"entity_id":          fmt.Sprintf("%d", account.ID),
				"custom_variable_id": fmt.Sprintf("%d", override.CustomVariableID),
			}

			match, err := f.Match(data)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to filter Custom Variables",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "filter"),
				})
				return nil, diags
			}
			if !match {
				continue
			}

			arr = append(arr, data)
		}
	}

	return arr, nil
}

func getAllAccountCacheOverrides(d *schema.ResourceData, client *hc.Client) ([]map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var accountCaches hc.AccountCacheListResponse
	err := client.GET("/v3/account-cache", &accountCaches)
	if err != nil {
		return nil, diag.Errorf("Error getting account caches: %v", err)
	}

	f := hc.NewFilterable(d)
	var arr []map[string]interface{}
	for _, accountCache := range accountCaches.Data {
		var overrides hc.CustomVariableOverrideListResponse
		err := client.GET(fmt.Sprintf("/v3/account-cache/%d/custom-variable?count=999999", accountCache.ID), &overrides)
		if err != nil {
			return nil, diag.Errorf("Error getting account cache overrides: %v", err)
		}

		for _, override := range overrides.Data.Items {
			if override.Override.Value == nil {
				continue
			}

			cvValueStr, err := hc.PackCvValueIntoJsonStr(override.Override.Value)
			if err != nil {
				return nil, diag.Errorf("Error packing CV value: %v", err)
			}

			data := map[string]interface{}{
				"value":              cvValueStr,
				"entity_type":        "account-cache",
				"entity_id":          fmt.Sprintf("%d", accountCache.ID),
				"custom_variable_id": fmt.Sprintf("%d", override.CustomVariableID),
			}

			match, err := f.Match(data)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to filter Custom Variables",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "filter"),
				})
				return nil, diags
			}
			if !match {
				continue
			}

			arr = append(arr, data)
		}
	}

	return arr, nil
}
