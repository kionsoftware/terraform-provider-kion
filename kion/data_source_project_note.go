package kion

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceProjectNote() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProjectNoteRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"create_user_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"create_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_update_user_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_update_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"text": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceProjectNoteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Get("id").(string)
	d.SetId(ID) // Set the ID early so it's available even if we error out

	// Get project notes with project_id filter
	resp := new(hc.ProjectNoteListResponse)
	params := url.Values{}
	params.Add("project_id", strconv.Itoa(d.Get("project_id").(int)))

	err := client.GET(fmt.Sprintf("/v3/project-note?%s", params.Encode()), resp)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read Project Note: %v", err))
	}

	// Find the specific note by ID
	var found bool
	for _, item := range resp.Data {
		if strconv.Itoa(int(item.ID)) == ID {
			data := make(map[string]interface{})
			data["create_user_id"] = item.CreateUserID
			data["create_user_name"] = item.CreateUserName
			data["last_update_user_id"] = item.LastUpdateUserID.Int
			data["last_update_user_name"] = item.LastUpdateUserName
			data["name"] = item.Name
			data["project_id"] = item.ProjectID
			data["text"] = item.Text

			for k, v := range data {
				if err := hc.SafeSet(d, k, v, "Unable to read Project Note"); err != nil {
					diags = append(diags, err...)
				}
			}
			found = true
			break
		}
	}

	if !found {
		return diag.FromErr(fmt.Errorf("unable to find Project Note"))
	}

	return diags
}
