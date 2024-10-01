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
			"account_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "AWS account application ID where the AMI is stored.",
			},
			"aws_ami_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Image ID of the AMI from AWS.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description for the AMI in the application.",
			},
			"expires_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The expiration date and time of the AMI. This may be null.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the AMI.",
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "AWS region where the AMI exists.",
			},
			"sync_deprecation": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Will sync the expiration date from the system into the AMI in AWS.",
			},
			"sync_tags": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Will sync the AWS tags from the source AMI into all the accounts where the AMI is shared.",
			},
			"unavailable_in_aws": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if the AMI is unavailable in AWS.",
			},
			"owner_user_group_ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of group IDs who own the AMI.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"owner_user_ids": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of user IDs who own the AMI.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
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
		ami := item.Ami

		data := make(map[string]interface{})
		data["account_id"] = ami.AccountID
		data["aws_ami_id"] = ami.AwsAmiID
		data["description"] = ami.Description

		if ami.ExpiresAt.Valid {
			data["expires_at"] = ami.ExpiresAt.Time.Format(time.RFC3339)
		} else {
			data["expires_at"] = nil
		}

		data["id"] = ami.ID
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
