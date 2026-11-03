resource "kion_aws_account" "example" {
  # Required
  name     = "example"
  payer_id = 1

  # Optional
  # account_alias                = "example"
  # account_number               = "example"
  # account_type_id              = 1
  # commercial_account_name      = "example"
  # create_govcloud              = false
  # email                        = "example"
  # gov_account_name             = "example"
  # include_linked_account_spend = false
  # labels                       = {}
  # last_updated                 = "example"
  # linked_role                  = "example"
  # project_id                   = 1
  # skip_access_checking         = false
  # start_datecode               = "example"
  # use_org_account_info         = false

  # aws_organizational_unit {
  #   name        = "example"
  #   org_unit_id = "example"
  # }
  # move_project_settings {
  #   financials    = "example"
  #   move_datecode = 1
  # }
}
