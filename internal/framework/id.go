package framework

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// StringToInt64 converts a string ID (from import) to int64.
func StringToInt64(s string) (int64, diag.Diagnostics) {
	var diags diag.Diagnostics

	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		diags.AddError(
			"Invalid ID",
			fmt.Sprintf("Could not parse %q as an integer ID: %s", s, err),
		)
	}

	return id, diags
}

// Int64ToString converts an int64 ID to a string for Terraform state.
func Int64ToString(id int64) string {
	return strconv.FormatInt(id, 10)
}

// Uint64ToString converts a uint64 ID to a string for Terraform state.
func Uint64ToString(id uint64) string {
	return strconv.FormatUint(id, 10)
}
