resource "kion_automation_policy" "example" {
  # Required
  cloud_provider_policies = [{
    # apply_to_all_regions = false
    # cloud_provider_id    = 1
    # policy               = "example"
    # regions              = []
    # resources            = []
  }]
  name                    = "example"

  # Optional
  # cloud_rule_ids       = []
  # description          = "example"
  # enabled              = false
  # engine               = 1
  # owner_user_group_ids = []
  # owner_user_ids       = []
  # scheduled_frequency = {
  #   days_of_month        = []
  #   days_of_week         = []
  #   end_date             = "example"
  #   hour                 = 1
  #   interval_hours       = 1
  #   minute               = 1
  #   quarterly_recurrence = 1
  #   start_date           = "example"
  #   time_zone_identifier = "example"
  #   type                 = 1
  #   weekly_recurrence    = 1
  # }
}
