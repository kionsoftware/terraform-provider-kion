resource "kion_compliance_standard" "example" {
  # Required
  created_by_user_id = 1
  name               = "example"

  # Optional
  # cloud_rule_id        = 1
  # compliance_check_ids = []
  # description          = "example"
  # last_updated         = "example"
  # owner_user_group_ids = []
  # owner_user_ids       = []
}
