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

func dataSourceProjectCloudAccessRoleExemption() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceProjectCloudAccessRoleExemptionRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ou_cloud_access_role_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "ID of the OU cloud access role.",
						},
						"project_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "ID of the project.",
						},
					},
				},
			},
			"list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the project cloud access role exemption.",
						},
						"ou_cloud_access_role_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ID of the OU cloud access role being exempted from.",
						},
						"project_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ID of the project.",
						},
						"reason": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Reason the Cloud Access Role is being exempted.",
						},
					},
				},
			},
		},
	}
}

func dataSourceProjectCloudAccessRoleExemptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	allExemptions := make([]hc.ProjectCloudAccessRoleExemptionV1, 0)

	// Apply filters if provided
	filterList, hasFilter := d.GetOk("filter")
	var projectIDFilter int
	var ouCloudAccessRoleIDFilter int

	if hasFilter {
		filters := filterList.([]interface{})[0].(map[string]interface{})
		if v, ok := filters["project_id"].(int); ok && v != 0 {
			projectIDFilter = v
		}
		if v, ok := filters["ou_cloud_access_role_id"].(int); ok && v != 0 {
			ouCloudAccessRoleIDFilter = v
		}
	}

	// If we have a specific project ID, we can make a single request
	if projectIDFilter != 0 {
		resp := new(hc.ProjectCloudAccessRoleV1Response)
		endpoint := fmt.Sprintf("/v1/project/%d/ou-cloud-access-role", projectIDFilter)
		err := client.GET(endpoint, resp)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read Project Cloud Access Role Exemptions",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
			return diags
		}

		if resp.Data.ProjectExemptions != nil {
			allExemptions = append(allExemptions, resp.Data.ProjectExemptions...)
		}
	} else {
		// Without a project ID filter, we need to get all projects first
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

		// Iterate through all projects to get their exemptions
		for _, project := range projectResp.Data {
			resp := new(hc.ProjectCloudAccessRoleV1Response)
			endpoint := fmt.Sprintf("/v1/project/%d/ou-cloud-access-role", project.ID)
			err := client.GET(endpoint, resp)
			if err != nil {
				// Skip projects that return errors (might not have access)
				continue
			}

			if resp.Data.ProjectExemptions != nil {
				allExemptions = append(allExemptions, resp.Data.ProjectExemptions...)
			}
		}
	}

	// Apply additional filtering
	filteredExemptions := make([]hc.ProjectCloudAccessRoleExemptionV1, 0)
	for _, exemption := range allExemptions {
		if ouCloudAccessRoleIDFilter != 0 && exemption.OUCloudAccessRoleID != ouCloudAccessRoleIDFilter {
			continue
		}
		filteredExemptions = append(filteredExemptions, exemption)
	}

	// Convert to list for schema
	exemptionList := make([]map[string]interface{}, len(filteredExemptions))
	for i, exemption := range filteredExemptions {
		exemptionList[i] = map[string]interface{}{
			"id":                      exemption.ID,
			"ou_cloud_access_role_id": exemption.OUCloudAccessRoleID,
			"project_id":              exemption.ProjectID,
			"reason":                  exemption.Reason,
		}
	}

	diags = append(diags, hc.SafeSet(d, "list", exemptionList, "Unable to set Project Cloud Access Role Exemptions")...)

	// Set a unique ID for the data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
