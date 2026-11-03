resource "kion_custom_account" "example" {
  # Required
  account_name   = "example"
  account_number = "example"
  payer_id       = 1
  project_id     = 1
  start_datecode = "2026-01"

  # Optional
  # account_alias = "example"
}
