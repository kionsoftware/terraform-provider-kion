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

func dataSourceOUCloudAccessRoleExemption() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOUCloudAccessRoleExemptionRead,
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
						"ou_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "ID of the OU.",
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
							Description: "The ID of the OU cloud access role exemption.",
						},
						"ou_cloud_access_role_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ID of the OU cloud access role being exempted from.",
						},
						"ou_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ID of the OU.",
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

func dataSourceOUCloudAccessRoleExemptionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	allExemptions := make([]hc.OUCloudAccessRoleExemptionV1, 0)

	// Apply filters if provided
	filterList, hasFilter := d.GetOk("filter")
	var ouIDFilter int
	var ouCloudAccessRoleIDFilter int

	if hasFilter {
		filters := filterList.([]interface{})[0].(map[string]interface{})
		if v, ok := filters["ou_id"].(int); ok && v != 0 {
			ouIDFilter = v
		}
		if v, ok := filters["ou_cloud_access_role_id"].(int); ok && v != 0 {
			ouCloudAccessRoleIDFilter = v
		}
	}

	// If we have a specific OU ID, we can make a single request
	if ouIDFilter != 0 {
		resp := new(hc.OUCloudAccessRoleV1Response)
		endpoint := fmt.Sprintf("/v1/ou/%d/ou-cloud-access-role", ouIDFilter)
		err := client.GET(endpoint, resp)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read OU Cloud Access Role Exemptions",
				Detail:   fmt.Sprintf("Error: %v", err.Error()),
			})
			return diags
		}

		if resp.Data.OUExemptions != nil {
			allExemptions = append(allExemptions, resp.Data.OUExemptions...)
		}
	} else {
		// Without an OU ID filter, we need to get all OUs first
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

		// Iterate through all OUs to get their exemptions
		for _, ou := range ouResp.Data {
			resp := new(hc.OUCloudAccessRoleV1Response)
			endpoint := fmt.Sprintf("/v1/ou/%d/ou-cloud-access-role", ou.ID)
			err := client.GET(endpoint, resp)
			if err != nil {
				// Skip OUs that return errors (might not have access)
				continue
			}

			if resp.Data.OUExemptions != nil {
				allExemptions = append(allExemptions, resp.Data.OUExemptions...)
			}
		}
	}

	// Apply additional filtering
	filteredExemptions := make([]hc.OUCloudAccessRoleExemptionV1, 0)
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
			"ou_id":                   exemption.OUID,
			"reason":                  exemption.Reason,
		}
	}

	diags = append(diags, hc.SafeSet(d, "list", exemptionList, "Unable to set OU Cloud Access Role Exemptions")...)

	// Set a unique ID for the data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}