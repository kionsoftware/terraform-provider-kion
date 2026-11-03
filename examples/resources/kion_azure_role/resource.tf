resource "kion_azure_role" "example" {
  # Required
  name             = "example"
  role_permissions = "example"

  # Optional
  # car_restricted_user_group_ids = []
  # car_restricted_user_ids       = []
  # description                   = "example"
  # owner_user_group_ids          = []
  # owner_user_ids                = []
}
