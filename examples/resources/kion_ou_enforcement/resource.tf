resource "kion_ou_enforcement" "example" {
  # Required
  ou_id     = 1
  threshold = 1
  timeframe = "example"

  # Optional
  # cloud_rule_id               = 1
  # description                 = "example"
  # enabled                     = false
  # overburn                    = false
  # service_id                  = 1
  # threshold_type              = "example"
  # trigger_planned_amount_type = "example"
  # user_group_ids              = []
  # user_ids                    = []
}
