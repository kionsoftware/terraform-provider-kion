resource "kion_azure_arm_template" "example" {
  # Required
  deployment_mode          = 1
  name                     = "example"
  resource_group_name      = "example"
  resource_group_region_id = 1
  template                 = "example"

  # Optional
  # description          = "example"
  # owner_user_group_ids = []
  # owner_user_ids       = []
  # template_parameters  = "example"
}
