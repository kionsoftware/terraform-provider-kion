resource "kion_account_cache" "example" {
  # Required
  account_name = "example"
  payer_id     = 1

  # Optional
  # account_alias                = "example"
  # account_email                = "example"
  # account_number               = "example"
  # account_type_id              = 1
  # commercial_account_name      = "example"
  # create_govcloud              = false
  # gov_account_name             = "example"
  # include_linked_account_spend = false
  # last_updated                 = "example"
  # linked_role                  = "example"
  # organizational_unit = {
  #   name        = "example"
  #   org_unit_id = "example"
  # }
  # skip_access_checking         = false
}
