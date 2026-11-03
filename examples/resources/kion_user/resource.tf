resource "kion_user" "example" {
  # Required
  email      = "example"
  first_name = "example"
  idms_id    = 1
  last_name  = "example"
  username   = "example"

  # Optional
  # mfa            = 1
  # phone          = "example"
  # user_group_ids = []
}
