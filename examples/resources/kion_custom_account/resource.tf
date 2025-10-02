resource "kion_custom_account" "complete_example" {
  # Required fields
  name           = "Terraform Complete Custom Account"
  payer_id       = 1
  account_number = "CUSTOM-123456"

  # Optional fields
  project_id          = 42
  start_datecode      = "2024-03"

  # Labels for the account (only supported for project accounts)
  labels = {
    environment = "production"
    team        = "platform"
    cost_center = "12345"
  }
}