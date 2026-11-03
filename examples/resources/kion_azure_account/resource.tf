resource "kion_azure_account" "example" {
  # Required
  account_name   = "example"
  payer_id       = 1
  project_id     = 1
  start_datecode = "2026-01"

  # Optional
  # account_alias        = "example"
  # account_type_id      = 1
  # skip_access_checking = false
  # subscription_uuid    = "example"
}
