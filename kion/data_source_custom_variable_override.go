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
						"value_string": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value_list": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"value_map": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
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

	diags = append(diags, hc.SafeSet(d, "list", arr, "Failed to set list")...)
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func getAllOUOverrides(d *schema.ResourceData, client *hc.Client) ([]map[string]interface{}, diag.Diagnostics) {
	var ous hc.OUListResponse
	err := client.GET("/v3/ou", &ous)
	if err != nil {
		return nil, hc.HandleError(fmt.Errorf("error getting OUs: %v", err))
	}

	f := hc.NewFilterable(d)
	var arr []map[string]interface{}

	for _, ou := range ous.Data {
		var overrides hc.CustomVariableOverrideListResponse
		err := client.GET(fmt.Sprintf("/v3/ou/%d/custom-variable?count=999999", ou.ID), &overrides)
		if err != nil {
			return nil, hc.HandleError(fmt.Errorf("error getting OU overrides: %v", err))
		}

		for _, override := range overrides.Data.Items {
			if override.Override == nil || override.Override.Value == nil {
				continue
			}

			cvResp := new(hc.CustomVariableResponse)
			err := client.GET(fmt.Sprintf("/v3/custom-variable/%d", override.CustomVariableID), cvResp)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("failed to get custom variable type: %v", err))
			}

			cvValueStr, err := hc.PackCvValueIntoJsonStr(override.Override.Value, cvResp.Data.Type)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("failed to process value: %v", err))
			}

			data := map[string]interface{}{
				"value_string":       cvValueStr,
				"entity_type":        "ou",
				"entity_id":          fmt.Sprintf("%d", ou.ID),
				"custom_variable_id": fmt.Sprintf("%d", override.CustomVariableID),
			}

			match, err := f.Match(data)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("unable to filter Custom Variables: %v", err))
			} else if !match {
				continue
			}

			arr = append(arr, data)
		}
	}

	return arr, nil
}

func getAllProjectOverrides(d *schema.ResourceData, client *hc.Client) ([]map[string]interface{}, diag.Diagnostics) {
	var projects hc.ProjectListResponse
	err := client.GET("/v3/project", &projects)
	if err != nil {
		return nil, hc.HandleError(fmt.Errorf("error getting projects: %v", err))
	}

	f := hc.NewFilterable(d)
	var arr []map[string]interface{}
	for _, project := range projects.Data {
		var overrides hc.CustomVariableOverrideListResponse
		err := client.GET(fmt.Sprintf("/v3/project/%d/custom-variable?count=999999", project.ID), &overrides)
		if err != nil {
			return nil, hc.HandleError(fmt.Errorf("error getting project overrides: %v", err))
		}

		for _, override := range overrides.Data.Items {
			if override.Override.Value == nil {
				continue
			}

			cvResp := new(hc.CustomVariableResponse)
			err := client.GET(fmt.Sprintf("/v3/custom-variable/%d", override.CustomVariableID), cvResp)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("failed to get custom variable type: %v", err))
			}

			cvValueStr, err := hc.PackCvValueIntoJsonStr(override.Override.Value, cvResp.Data.Type)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("failed to process value: %v", err))
			}

			data := map[string]interface{}{
				"value_string":       cvValueStr,
				"entity_type":        "project",
				"entity_id":          fmt.Sprintf("%d", project.ID),
				"custom_variable_id": fmt.Sprintf("%d", override.CustomVariableID),
			}

			match, err := f.Match(data)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("unable to filter Custom Variables: %v", err))
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
	var accounts hc.AccountListResponse
	err := client.GET("/v3/account", &accounts)
	if err != nil {
		return nil, hc.HandleError(fmt.Errorf("error getting accounts: %v", err))
	}

	f := hc.NewFilterable(d)
	var arr []map[string]interface{}
	for _, account := range accounts.Data {
		var overrides hc.CustomVariableOverrideListResponse
		err := client.GET(fmt.Sprintf("/v3/account/%d/custom-variable?count=999999", account.ID), &overrides)
		if err != nil {
			return nil, hc.HandleError(fmt.Errorf("error getting account overrides: %v", err))
		}

		for _, override := range overrides.Data.Items {
			if override.Override.Value == nil {
				continue
			}

			cvResp := new(hc.CustomVariableResponse)
			err := client.GET(fmt.Sprintf("/v3/custom-variable/%d", override.CustomVariableID), cvResp)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("failed to get custom variable type: %v", err))
			}

			cvValueStr, err := hc.PackCvValueIntoJsonStr(override.Override.Value, cvResp.Data.Type)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("failed to process value: %v", err))
			}

			data := map[string]interface{}{
				"value_string":       cvValueStr,
				"entity_type":        "account",
				"entity_id":          fmt.Sprintf("%d", account.ID),
				"custom_variable_id": fmt.Sprintf("%d", override.CustomVariableID),
			}

			match, err := f.Match(data)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("unable to filter Custom Variables: %v", err))
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
	var accountCaches hc.AccountCacheListResponse
	err := client.GET("/v3/account-cache", &accountCaches)
	if err != nil {
		return nil, hc.HandleError(fmt.Errorf("error getting account caches: %v", err))
	}

	f := hc.NewFilterable(d)
	var arr []map[string]interface{}
	for _, accountCache := range accountCaches.Data {
		var overrides hc.CustomVariableOverrideListResponse
		err := client.GET(fmt.Sprintf("/v3/account-cache/%d/custom-variable?count=999999", accountCache.ID), &overrides)
		if err != nil {
			return nil, hc.HandleError(fmt.Errorf("error getting account cache overrides: %v", err))
		}

		for _, override := range overrides.Data.Items {
			if override.Override.Value == nil {
				continue
			}

			cvResp := new(hc.CustomVariableResponse)
			err := client.GET(fmt.Sprintf("/v3/custom-variable/%d", override.CustomVariableID), cvResp)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("failed to get custom variable type: %v", err))
			}

			cvValueStr, err := hc.PackCvValueIntoJsonStr(override.Override.Value, cvResp.Data.Type)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("failed to process value: %v", err))
			}

			data := map[string]interface{}{
				"value_string":       cvValueStr,
				"entity_type":        "account-cache",
				"entity_id":          fmt.Sprintf("%d", accountCache.ID),
				"custom_variable_id": fmt.Sprintf("%d", override.CustomVariableID),
			}

			match, err := f.Match(data)
			if err != nil {
				return nil, hc.HandleError(fmt.Errorf("unable to filter Custom Variables: %v", err))
			}
			if !match {
				continue
			}

			arr = append(arr, data)
		}
	}

	return arr, nil
}
