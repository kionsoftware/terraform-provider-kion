package kion

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

// OUChanges allows moving an OU if the parent ID changes and updating permissions.
func OUChanges(k *hc.Client, d *schema.ResourceData, diags diag.Diagnostics, hasChanged int) (diag.Diagnostics, int) {
	// Handle OU move.
	if d.HasChanges("parent_ou_id") {
		hasChanged++
		arrParentOUID, _, _, err := hc.AssociationChangedInt(d, "parent_ou_id")
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to determine changeset for ParentOU on OrganizationalUnit",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), d.Id()),
			})
			return diags, hasChanged
		}
		_, err = k.POST(fmt.Sprintf("/v2/ou/%s/move", d.Id()), arrParentOUID)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to add update Parent OU for Organizational Unit",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), d.Id()),
			})
			return diags, hasChanged
		}
	}

	return diags, hasChanged
}
