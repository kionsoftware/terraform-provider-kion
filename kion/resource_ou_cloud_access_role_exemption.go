package kion

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceOUCloudAccessRoleExemption() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOUCloudAccessRoleExemptionCreate,
		ReadContext:   resourceOUCloudAccessRoleExemptionRead,
		UpdateContext: resourceOUCloudAccessRoleExemptionUpdate,
		DeleteContext: resourceOUCloudAccessRoleExemptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceOUCloudAccessRoleExemptionRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"ou_cloud_access_role_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the OU cloud access role in the application being exempted from.",
			},
			"ou_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the OU in the application.",
			},
			"reason": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Reason the Cloud Access Role is being exempted.",
			},
		},
	}
}

func resourceOUCloudAccessRoleExemptionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	post := hc.OUCloudAccessRoleExemptionCreate{
		OUCloudAccessRoleID: d.Get("ou_cloud_access_role_id").(int),
		OUID:                d.Get("ou_id").(int),
		Reason:              d.Get("reason").(string),
	}

	resp, err := client.POST("/v3/ou-cloud-access-role-exemption", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create OU Cloud Access Role Exemption",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	return resourceOUCloudAccessRoleExemptionRead(ctx, d, m)
}

func resourceOUCloudAccessRoleExemptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()
	exemptionID, err := strconv.Atoi(ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse exemption ID",
			Detail:   fmt.Sprintf("Error: %v", err.Error()),
		})
		return diags
	}

	// Get ou_id from state or try to find it
	ouID := d.Get("ou_id").(int)
	if ouID == 0 {
		// If ou_id is not in state (e.g., during import), we need to search all OUs
		ouResp := new(hc.OUListResponse)
		err := client.GET("/v3/ou", ouResp)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read OUs",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
			return diags
		}

		// Search through all OUs to find the exemption
		found := false
		for _, ou := range ouResp.Data {
			resp := new(hc.OUCloudAccessRoleV1Response)
			endpoint := fmt.Sprintf("/v1/ou/%d/ou-cloud-access-role", ou.ID)
			err := client.GET(endpoint, resp)
			if err != nil {
				continue
			}

			if resp.Data.OUExemptions != nil {
				if hc.FindOUExemptionByID(resp.Data.OUExemptions, exemptionID) != nil {
					ouID = ou.ID
					found = true
				}
			}
			if found {
				break
			}
		}

		if !found {
			d.SetId("")
			return diags
		}
	}

	// Now fetch the exemption using the v1 endpoint
	resp := new(hc.OUCloudAccessRoleV1Response)
	endpoint := fmt.Sprintf("/v1/ou/%d/ou-cloud-access-role", ouID)
	err = client.GET(endpoint, resp)
	if err != nil {
		if resErr, ok := err.(*hc.RequestError); ok && resErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read OU Cloud Access Role Exemption",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// Find the specific exemption
	var found *hc.OUCloudAccessRoleExemptionV1
	if resp.Data.OUExemptions != nil {
		found = hc.FindOUExemptionByID(resp.Data.OUExemptions, exemptionID)
	}

	if found == nil {
		d.SetId("")
		return diags
	}

	// Use SafeSet for cleaner error handling
	diags = append(diags, hc.SafeSet(d, "ou_cloud_access_role_id", found.OUCloudAccessRoleID, "Unable to set ou_cloud_access_role_id")...)
	diags = append(diags, hc.SafeSet(d, "ou_id", found.OUID, "Unable to set ou_id")...)
	diags = append(diags, hc.SafeSet(d, "reason", found.Reason, "Unable to set reason")...)

	return diags
}

func resourceOUCloudAccessRoleExemptionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := false

	// Check if reason has changed
	if d.HasChange("reason") {
		hasChanged = true
	}

	if hasChanged {
		patch := hc.OUCloudAccessRoleExemptionUpdate{
			Reason: d.Get("reason").(string),
		}

		err := client.PATCH(fmt.Sprintf("/v3/ou-cloud-access-role-exemption/%s", ID), patch)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update OU Cloud Access Role Exemption",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return resourceOUCloudAccessRoleExemptionRead(ctx, d, m)
}

func resourceOUCloudAccessRoleExemptionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/ou-cloud-access-role-exemption/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete OU Cloud Access Role Exemption",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// Give the API time to delete
	ouID := d.Get("ou_id").(int)
	exemptionID, _ := strconv.Atoi(ID)
	
	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{"200"},
		Target:  []string{"404"},
		Refresh: func() (interface{}, string, error) {
			resp := new(hc.OUCloudAccessRoleV1Response)
			endpoint := fmt.Sprintf("/v1/ou/%d/ou-cloud-access-role", ouID)
			err := client.GET(endpoint, resp)
			if err != nil {
				if resErr, ok := err.(*hc.RequestError); ok && resErr.StatusCode == http.StatusNotFound {
					return resp, "404", nil
				}
				return nil, "", err
			}
			
			// Check if the exemption still exists
			found := false
			if resp.Data.OUExemptions != nil {
				if hc.FindOUExemptionByID(resp.Data.OUExemptions, exemptionID) != nil {
					found = true
				}
			}
			
			if !found {
				return resp, "404", nil
			}
			return resp, "200", nil
		},
		Timeout:                   1 * time.Minute,
		Delay:                     5 * time.Second,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 1,
	}

	_, err = deleteStateConf.WaitForStateContext(ctx)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete OU Cloud Access Role Exemption",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	d.SetId("")

	return diags
}