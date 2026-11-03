resource "kion_compliance_check" "example" {
  # Required
  cloud_provider_id        = 1
  compliance_check_type_id = 1
  name                     = "example"

  # Optional
  # azure_policy_id        = 1
  # body                   = "example"
  # compliance_control_ids = []
  # compliance_standard_id = 1
  # created_by_user_id     = 1
  # description            = "example"
  # frequency_minutes      = 1
  # frequency_type_id      = 1
  # is_all_regions         = false
  # is_auto_archived       = false
  # last_updated           = "example"
  # owner_user_group_ids   = []
  # owner_user_ids         = []
  # regions                = []
  # severity_type_id       = 1
}
