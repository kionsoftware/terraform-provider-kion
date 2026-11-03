resource "kion_ou_permission_mapping" "example" {
  # Required
  app_role_id = 1
  ou_id       = 1

  # Optional
  # user_groups_ids = []
  # user_ids        = []
}
