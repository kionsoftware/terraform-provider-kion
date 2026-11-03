resource "kion_service_catalog" "example" {
  # Required
  account_id   = 1
  name         = "example"
  portfolio_id = "example"
  region       = "example"

  # Optional
  # description          = "example"
  # owner_user_group_ids = []
  # owner_user_ids       = []
  # tag_option           = false
}
