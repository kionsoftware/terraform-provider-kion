package kion

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Legacy specific filter fields (deprecated but supported for backwards compatibility)
						"username": {
							Description: "The username you wish to filter by.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"enabled": {
							Description: "Filter by whether the user is enabled.",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						// New generic filter fields
						"name": {
							Description: "The field name whose values you wish to filter by.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"values": {
							Description: "The values of the field name you specified.",
							Type:        schema.TypeList,
							Optional:    true,
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
				Description: "This is where Kion makes the discovered data available as a list of user IDs.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	resp := new(hc.UserListResponse) // Use the UserListResponse struct from models_user.go
	err := client.GET("/v3/user", resp)
	if diags := hc.HandleError(err); diags != nil {
		return diags
	}

	f := hc.NewFilterable(d)

	// Create a list to hold the filtered user IDs
	var userIDs []int
	for _, item := range resp.Data {
		data := map[string]interface{}{
			"id":       item.ID,
			"username": item.Username,
			"enabled":  item.Enabled,
		}

		match, err := f.Match(data)
		if diags := hc.HandleError(err); diags != nil {
			return diags
		}

		if match {
			userIDs = append(userIDs, item.ID)
		}
	}

	diags := hc.SafeSet(d, "list", userIDs, "list of user IDs")

	// Set a unique ID for this data source instance
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
