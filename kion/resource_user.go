package kion

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceUserRead,
		DeleteContext: resourceUserDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)
	idString := d.Get("id").(string)
	ID, err := strconv.Atoi(idString)

	if err != nil {
		return diag.Errorf("Error converting ID to integer: %s", err)
	}

	resp := new(hc.UserResponse)
	err = client.GET(fmt.Sprintf("/v3/user/%d", ID), resp)
	if diags := hc.HandleError(err); diags != nil {
		return diags
	}

	userID := strconv.Itoa(resp.Data.User.ID)
	d.SetId(userID)

	return nil
}

// resourceUserDelete implements a no-op delete
func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// This function does nothing, but Terraform requires it.
	// It clears the ID to signal that the resource has been "deleted" in Terraform's state.
	d.SetId("")
	return nil
}
