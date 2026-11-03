// Package sweep_test is the acceptance-test sweeper entrypoint: blank-importing
// a service package runs its init, which registers that package's sweeper.
//
// Only packages that register one are listed. A resource whose sweep.go says no
// sweeper is registered cannot be enumerated or deleted through the API, and is
// deliberately absent rather than present-and-silent. TestSweeperRegistration
// keeps this list honest.
package sweep_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	// Import every service package that registers a sweeper.
	_ "terraform-provider-kion/internal/service/account_cache"
	_ "terraform-provider-kion/internal/service/account_linkage"
	_ "terraform-provider-kion/internal/service/ami"
	_ "terraform-provider-kion/internal/service/app_api_key"
	_ "terraform-provider-kion/internal/service/azure_account"
	_ "terraform-provider-kion/internal/service/azure_arm_template"
	_ "terraform-provider-kion/internal/service/azure_policy"
	_ "terraform-provider-kion/internal/service/azure_role"
	_ "terraform-provider-kion/internal/service/category"
	_ "terraform-provider-kion/internal/service/cft"
	_ "terraform-provider-kion/internal/service/cloud_rule"
	_ "terraform-provider-kion/internal/service/compliance_check"
	_ "terraform-provider-kion/internal/service/compliance_family"
	_ "terraform-provider-kion/internal/service/compliance_level"
	_ "terraform-provider-kion/internal/service/compliance_program"
	_ "terraform-provider-kion/internal/service/compliance_standard"
	_ "terraform-provider-kion/internal/service/custom_account"
	_ "terraform-provider-kion/internal/service/custom_variable"
	_ "terraform-provider-kion/internal/service/funding_source"
	_ "terraform-provider-kion/internal/service/gcp_account"
	_ "terraform-provider-kion/internal/service/gcp_iam_role"
	_ "terraform-provider-kion/internal/service/gcp_service_account"
	_ "terraform-provider-kion/internal/service/iam_policy"
	_ "terraform-provider-kion/internal/service/idms"
	_ "terraform-provider-kion/internal/service/idms_group_association"
	_ "terraform-provider-kion/internal/service/label"
	_ "terraform-provider-kion/internal/service/oauth2_client"
	_ "terraform-provider-kion/internal/service/ou"
	_ "terraform-provider-kion/internal/service/ou_cloud_access_role"
	_ "terraform-provider-kion/internal/service/ou_note"
	_ "terraform-provider-kion/internal/service/project"
	_ "terraform-provider-kion/internal/service/project_cloud_access_role"
	_ "terraform-provider-kion/internal/service/project_line_item"
	_ "terraform-provider-kion/internal/service/saml_group_association"
	_ "terraform-provider-kion/internal/service/scope"
	_ "terraform-provider-kion/internal/service/service_control_policy"
	_ "terraform-provider-kion/internal/service/user_group"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
