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
		return hc.HandleError(fmt.Errorf("unable to read Custom Variables: %v", err))
	}

	f := hc.NewFilterable(d)
	arr := make([]map[string]interface{}, 0)

	for _, item := range resp.Data.Items {
		cvValueStr, err := hc.PackCvValueIntoJsonStr(item.DefaultValue, item.Type)
		if err != nil {
			return hc.HandleError(fmt.Errorf("failed to process default_value: %v", err))
		}

		data := map[string]interface{}{
			"name":                     item.Name,
			"description":              item.Description,
			"type":                     item.Type,
			"default_value":            cvValueStr,
			"value_validation_regex":   item.ValueValidationRegex,
			"value_validation_message": item.ValueValidationMessage,
			"key_validation_regex":     item.KeyValidationRegex,
			"key_validation_message":   item.KeyValidationMessage,
			"owner_user_ids":           item.OwnerUserIDs,
			"owner_user_group_ids":     item.OwnerUserGroupIDs,
		}

		match, err := f.Match(data)
		if err != nil {
			return hc.HandleError(fmt.Errorf("unable to filter Custom Variables: %v", err))
		} else if !match {
			continue
		}

		arr = append(arr, data)
	}

	diags = append(diags, hc.SafeSet(d, "list", arr, "Failed to set list")...)
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
