resource "kion_funding_source_permission_mapping" "example" {
  # Required
  app_role_id       = 1
  funding_source_id = 1

  # Optional
  # user_groups_ids = []
  # user_ids        = []
}
