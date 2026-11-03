resource "kion_funding_source_enforcement" "example" {
  # Required
  funding_source_id = 1
  threshold         = 1
  timeframe         = "example"

  # Optional
  # cloud_rule_id  = 1
  # description    = "example"
  # enabled        = false
  # overburn       = false
  # spend_option   = "example"
  # user_group_ids = []
  # user_ids       = []
}
