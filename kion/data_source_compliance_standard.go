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

func dataSourceComplianceStandard() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceComplianceStandardRead,
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
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceComplianceStandardRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	resp := new(hc.ComplianceStandardListResponse)
	if err := client.GET("/v3/compliance/standard", resp); err != nil {
		return diag.FromErr(err)
	}

	f := hc.NewFilterable(d)

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data {
		data := make(map[string]interface{})
		data["created_at"] = item.CreatedAt
		data["created_by_user_id"] = item.CreatedByUserID
		data["ct_managed"] = item.CtManaged
		data["description"] = item.Description
		data["id"] = item.ID
		data["name"] = item.Name

		match, err := f.Match(data)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error matching compliance standard: %w", err))
		} else if !match {
			continue
		}

		arr = append(arr, data)
	}

	if err := d.Set("list", arr); err != nil {
		return diag.FromErr(fmt.Errorf("error setting list: %w", err))
	}

	// Always run.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
