resource "kion_gcp_account" "example" {
  # Required
  account_name   = "example"
  payer_id       = 1
  project_id     = 1
  start_datecode = "2026-01"

  # Optional
  # account_alias           = "example"
  # account_type_id         = 1
  # google_cloud_project_id = "example"
  # skip_access_checking    = false
}
