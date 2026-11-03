resource "kion_project_enforcement" "example" {
  # Required
  project_id = 1
  threshold  = 1
  timeframe  = "example"

  # Optional
  # amount_type            = "example"
  # cloud_rule_id          = 1
  # description            = "example"
  # enabled                = false
  # notification_emails    = []
  # notification_frequency = "example"
  # overburn               = false
  # service_id             = 1
  # spend_option           = "example"
  # threshold_type         = "example"
  # user_group_ids         = []
  # user_ids               = []
}
