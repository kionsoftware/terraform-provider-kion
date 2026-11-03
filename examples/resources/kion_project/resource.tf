resource "kion_project" "example" {
  # Required
  name  = "example"
  ou_id = 1

  # Optional
  # archived             = false
  # auto_pay             = false
  # budget = [{
  #   end_datecode   = "example"
  #   start_datecode = "example"
  #   amount             = 0.0
  #   data = [{
  #     amount   = 0.0
  #     datecode = "example"
  #     funding_source_id = 1
  #     priority          = 1
  #   }]
  #   funding_source_ids = []
  # }]
  # default_aws_region   = "example"
  # description          = "example"
  # labels               = {}
  # last_updated         = "example"
  # move_ou_settings = [{
  #   cloud_rule_setting = "example"
  #   financial_setting  = "example"
  #   spend_plan_setting = "example"
  # }]
  # owner_user_group_ids = []
  # owner_user_ids       = []
  # permission_scheme_id = 1
  # project_funding = [{
  #   amount            = 0.0
  #   end_datecode      = "example"
  #   funding_order     = 1
  #   funding_source_id = 1
  #   start_datecode    = "example"
  # }]
}
