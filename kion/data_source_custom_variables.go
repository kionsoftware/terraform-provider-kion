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

func dataSourceCustomVariable() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceCustomVariablesRead,
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
						// Notice there is no 'id' field specified because it will be created.
						"name": {
							Type:     schema.TypeString,
							Computed: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"default_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value_validation_regex": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value_validation_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"key_validation_regex": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"key_validation_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner_user_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Computed: true,
						},
						"owner_user_group_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceCustomVariablesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	resp := new(hc.CustomVariableListResponse)
	err := client.GET("/v3/custom-variable?count=999999", resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Custom Variables",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	f := hc.NewFilterable(d)

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data.Items {
		cvValueStr, err := hc.PackCvValueIntoJsonStr(item.DefaultValue)
		if err != nil {
			return diag.Errorf(err.Error())
		}

		data := make(map[string]interface{})
		data["name"] = item.Name
		data["description"] = item.Description
		data["type"] = item.Type
		data["default_value"] = cvValueStr
		data["value_validation_regex"] = item.ValueValidationRegex
		data["value_validation_message"] = item.ValueValidationMessage
		data["key_validation_regex"] = item.KeyValidationRegex
		data["key_validation_message"] = item.KeyValidationMessage
		data["owner_user_ids"] = item.OwnerUserIDs
		data["owner_user_group_ids"] = item.OwnerUserGroupIDs

		match, err := f.Match(data)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to filter Custom Variables",
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
			Summary:  "Unable to read Custom Variables",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	// Always run.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
