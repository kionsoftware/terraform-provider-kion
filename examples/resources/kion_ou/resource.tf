resource "kion_ou" "example" {
  # Required
  name                 = "example"
  parent_ou_id         = 1
  permission_scheme_id = 1

  # Optional
  # description          = "example"
  # labels               = {}
  # last_updated         = "example"
  # owner_user_group_ids = []
  # owner_user_ids       = []
}
