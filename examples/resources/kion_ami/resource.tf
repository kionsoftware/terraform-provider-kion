resource "kion_ami" "example" {
  # Required
  account_id = 1
  aws_ami_id = "example"
  name       = "example"
  region     = "example"

  # Optional
  # description               = "example"
  # expiration_alert_number   = 1
  # expiration_alert_unit     = "example"
  # expiration_notify         = false
  # expiration_warning_number = 1
  # expiration_warning_unit   = "example"
  # expires_at                = "example"
  # owner_user_group_ids      = []
  # owner_user_ids            = []
  # sync_deprecation          = false
  # sync_tags                 = false
}
