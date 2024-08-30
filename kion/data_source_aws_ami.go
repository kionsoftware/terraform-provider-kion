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

func dataSourceAwsAmi() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsAmiRead,
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
				Description: "This is where Kion makes the discovered AMI data available as a list of resources.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"aws_ami_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expires_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sync_deprecation": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"sync_tags": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"unavailable_in_aws": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"owner_user_groups": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"owner_users": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsAmiRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	resp := new(hc.AmiListResponse)
	err := client.GET("/v3/ami", resp)
	if err != nil {
		return hc.HandleError(fmt.Errorf("unable to read AWS AMI: %v", err))
	}

	f := hc.NewFilterable(d)

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data {
		ami := item.Ami // Access the nested Ami struct

		data := make(map[string]interface{})
		data["account_id"] = ami.AccountID
		data["aws_ami_id"] = ami.AwsAmiID
		data["description"] = ami.Description
		data["expires_at"] = ami.ExpiresAt.Format(time.RFC3339)
		data["name"] = ami.Name
		data["region"] = ami.Region
		data["sync_deprecation"] = ami.SyncDeprecation
		data["sync_tags"] = ami.SyncTags
		data["unavailable_in_aws"] = ami.UnavailableInAws
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)

		match, err := f.Match(data)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to filter AWS AMI",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "filter"),
			})
			return diags
		} else if !match {
			continue
		}

		arr = append(arr, data)
	}

	diags = append(diags, hc.SafeSet(d, "list", arr, "Unable to read AWS AMI")...)

	// Always run.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
