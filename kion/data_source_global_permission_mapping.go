package kion

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceGlobalPermissionsMapping() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceGlobalPermissionMappingRead,
		Schema: map[string]*schema.Schema{
			"list": {
				Description: "List of global permission mappings.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_role_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"user_groups_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
						"user_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
					},
				},
			},
		},
	}
}

func dataSourceGlobalPermissionMappingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	resp := new(hc.ProjectPermissionMappingListResponse)
	err := client.GET("/v3/global/permission-mapping", resp)
	if err != nil {
		return diag.FromErr(err)
	}

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data {
		data := map[string]interface{}{
			"app_role_id":     item.AppRoleID,
			"user_groups_ids": item.UserGroupsIDs,
			"user_ids":        item.UserIDs,
		}
		arr = append(arr, data)
	}

	if err := d.Set("list", arr); err != nil {
		return diag.FromErr(err)
	}

	// Set the ID of the datasource to a unique value, which is the current timestamp
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return nil
}
