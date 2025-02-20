package kion

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAccountRead,
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
						"regex": {
							Description: "Dictates if the values provided should be treated as regular expressions.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"values": {
							Description: "The values of the field name you specified.",
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
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
						"account_alias": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"account_number": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"account_type_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"car_external_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"include_linked_account_spend": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"linked_account_number": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"linked_role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"payer_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"project_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"service_external_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"skip_access_checking": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"start_datecode": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"use_org_account_info": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	tflog.Debug(ctx, "Reading accounts list")

	resp := new(hc.AccountListResponse)
	if err := client.GET("/v3/account", resp); err != nil {
		return append(diags, hc.HandleError(fmt.Errorf("failed to read accounts: %v", err))...)
	}

	f := hc.NewFilterable(d)

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data {
		data := map[string]interface{}{
			"account_alias":                item.AccountAlias,
			"account_number":               item.AccountNumber,
			"account_type_id":              item.AccountTypeID,
			"car_external_id":              item.CARExternalID,
			"created_at":                   item.CreatedAt,
			"email":                        item.Email,
			"id":                           item.ID,
			"include_linked_account_spend": item.IncludeLinkedAccountSpend,
			"linked_account_number":        item.LinkedAccountNumber,
			"linked_role":                  item.LinkedRole,
			"name":                         item.Name,
			"payer_id":                     item.PayerID,
			"project_id":                   item.ProjectID,
			"service_external_id":          item.ServiceExternalID,
			"skip_access_checking":         item.SkipAccessChecking,
			"start_datecode":               item.StartDatecode,
			"use_org_account_info":         item.UseOrgAccountInfo,
		}

		match, err := f.Match(data)
		if err != nil {
			return append(diags, hc.HandleError(fmt.Errorf("failed to filter accounts: %v", err))...)
		}
		if !match {
			continue
		}

		arr = append(arr, data)
	}

	diags = append(diags, hc.SafeSet(d, "list", arr, "Failed to set accounts list")...)

	// Always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
