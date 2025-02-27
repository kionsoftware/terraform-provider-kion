package kion

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceProjectNote() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectNoteCreate,
		ReadContext:   resourceProjectNoteRead,
		UpdateContext: resourceProjectNoteUpdate,
		DeleteContext: resourceProjectNoteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceProjectNoteRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"create_user_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"text": {
				Type:     schema.TypeString,
				Required: true,
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceProjectNoteCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	post := hc.ProjectNoteCreate{
		CreateUserID: uint(d.Get("create_user_id").(int)),
		Name:         d.Get("name").(string),
		ProjectID:    uint(d.Get("project_id").(int)),
		Text:         d.Get("text").(string),
	}

	resp, err := client.POST("/v3/project-note", post)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to create Project Note: %v", err))
	} else if resp.RecordID == 0 {
		return diag.FromErr(fmt.Errorf("unable to create Project Note: received item ID of 0"))
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	return resourceProjectNoteRead(ctx, d, m)
}

func resourceProjectNoteRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

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
		d.SetId("")
		return nil
	}

	return diags
}

func resourceProjectNoteUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	// Get current state to populate required fields
	resp := new(hc.ProjectNoteListResponse)
	params := url.Values{}
	params.Add("project_id", strconv.Itoa(d.Get("project_id").(int)))

	err := client.GET(fmt.Sprintf("/v3/project-note?%s", params.Encode()), resp)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to read Project Note for update: %v", err))
	}

	// Find the current note data
	var found bool
	var patch hc.ProjectNoteUpdate

	for _, item := range resp.Data {
		if strconv.Itoa(int(item.ID)) == ID {
			patch = hc.ProjectNoteUpdate{
				ID:           item.ID,
				CreateUserID: uint(d.Get("create_user_id").(int)),
				ProjectID:    uint(d.Get("project_id").(int)),
				Name:         d.Get("name").(string),
				Text:         d.Get("text").(string),
			}
			found = true
			break
		}
	}

	if !found {
		return diag.FromErr(fmt.Errorf("unable to find Project Note for update"))
	}

	// Use v2 endpoint for update
	err = client.PATCH(fmt.Sprintf("/v2/project-note/%s", ID), patch)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to update Project Note: %v", err))
	}

	// Update last_updated timestamp
	if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
		return diag.FromErr(err)
	}

	return resourceProjectNoteRead(ctx, d, m)
}

func resourceProjectNoteDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	ID := d.Id()

	// For delete, we'll need to find the note first
	params := url.Values{}
	params.Add("project_id", strconv.Itoa(d.Get("project_id").(int)))

	resp := new(hc.ProjectNoteListResponse)
	err := client.GET(fmt.Sprintf("/v3/project-note?%s", params.Encode()), resp)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to find Project Note for deletion: %v", err))
	}

	// Find the specific note by ID
	var found bool
	for _, item := range resp.Data {
		if strconv.Itoa(int(item.ID)) == ID {
			found = true
			break
		}
	}

	if !found {
		// Note already deleted or doesn't exist
		d.SetId("")
		return nil
	}

	err = client.DELETE(fmt.Sprintf("/v3/project-note/%s", ID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to delete Project Note: %v", err))
	}

	d.SetId("")
	return nil
}
