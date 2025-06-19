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

func resourceProjectCloudAccessRoleExemption() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCloudAccessRoleExemptionCreate,
		ReadContext:   resourceProjectCloudAccessRoleExemptionRead,
		UpdateContext: resourceProjectCloudAccessRoleExemptionUpdate,
		DeleteContext: resourceProjectCloudAccessRoleExemptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceProjectCloudAccessRoleExemptionRead(ctx, d, m)
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
			"project_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the project in the application.",
			},
			"reason": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Reason the Cloud Access Role is being exempted.",
			},
		},
	}
}

func resourceProjectCloudAccessRoleExemptionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	post := hc.ProjectCloudAccessRoleExemptionCreate{
		OUCloudAccessRoleID: d.Get("ou_cloud_access_role_id").(int),
		ProjectID:           d.Get("project_id").(int),
		Reason:              d.Get("reason").(string),
	}

	resp, err := client.POST("/v3/project-cloud-access-role-exemption", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Project Cloud Access Role Exemption",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	return resourceProjectCloudAccessRoleExemptionRead(ctx, d, m)
}

func resourceProjectCloudAccessRoleExemptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	// Get project_id from state or try to find it
	projectID := d.Get("project_id").(int)
	if projectID == 0 {
		// If project_id is not in state (e.g., during import), we need to search all projects
		projectResp := new(hc.ProjectListResponse)
		err := client.GET("/v3/project", projectResp)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read Projects",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
			return diags
		}

		// Search through all projects to find the exemption
		found := false
		for _, project := range projectResp.Data {
			resp := new(hc.ProjectCloudAccessRoleV1Response)
			endpoint := fmt.Sprintf("/v1/project/%d/ou-cloud-access-role", project.ID)
			err := client.GET(endpoint, resp)
			if err != nil {
				continue
			}

			if resp.Data.ProjectExemptions != nil {
				if hc.FindProjectExemptionByID(resp.Data.ProjectExemptions, exemptionID) != nil {
					projectID = project.ID
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
	resp := new(hc.ProjectCloudAccessRoleV1Response)
	endpoint := fmt.Sprintf("/v1/project/%d/ou-cloud-access-role", projectID)
	err = client.GET(endpoint, resp)
	if err != nil {
		if resErr, ok := err.(*hc.RequestError); ok && resErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Project Cloud Access Role Exemption",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// Find the specific exemption
	var found *hc.ProjectCloudAccessRoleExemptionV1
	if resp.Data.ProjectExemptions != nil {
		found = hc.FindProjectExemptionByID(resp.Data.ProjectExemptions, exemptionID)
	}

	if found == nil {
		d.SetId("")
		return diags
	}

	// Use SafeSet for cleaner error handling
	diags = append(diags, hc.SafeSet(d, "ou_cloud_access_role_id", found.OUCloudAccessRoleID, "Unable to set ou_cloud_access_role_id")...)
	diags = append(diags, hc.SafeSet(d, "project_id", found.ProjectID, "Unable to set project_id")...)
	diags = append(diags, hc.SafeSet(d, "reason", found.Reason, "Unable to set reason")...)

	return diags
}

func resourceProjectCloudAccessRoleExemptionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := false

	// Check if reason has changed
	if d.HasChange("reason") {
		hasChanged = true
	}

	if hasChanged {
		patch := hc.ProjectCloudAccessRoleExemptionUpdate{
			Reason: d.Get("reason").(string),
		}

		err := client.PATCH(fmt.Sprintf("/v3/project-cloud-access-role-exemption/%s", ID), patch)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update Project Cloud Access Role Exemption",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return resourceProjectCloudAccessRoleExemptionRead(ctx, d, m)
}

func resourceProjectCloudAccessRoleExemptionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/project-cloud-access-role-exemption/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete Project Cloud Access Role Exemption",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// Give the API time to delete
	projectID := d.Get("project_id").(int)
	exemptionID, _ := strconv.Atoi(ID)

	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{"200"},
		Target:  []string{"404"},
		Refresh: func() (interface{}, string, error) {
			resp := new(hc.ProjectCloudAccessRoleV1Response)
			endpoint := fmt.Sprintf("/v1/project/%d/ou-cloud-access-role", projectID)
			err := client.GET(endpoint, resp)
			if err != nil {
				if resErr, ok := err.(*hc.RequestError); ok && resErr.StatusCode == http.StatusNotFound {
					return resp, "404", nil
				}
				return nil, "", err
			}

			// Check if the exemption still exists
			found := false
			if resp.Data.ProjectExemptions != nil {
				if hc.FindProjectExemptionByID(resp.Data.ProjectExemptions, exemptionID) != nil {
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
			Summary:  "Unable to delete Project Cloud Access Role Exemption",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	d.SetId("")

	return diags
}
