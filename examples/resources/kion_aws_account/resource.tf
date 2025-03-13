# Create a new AWS account and place it in a project with all available options
resource "kion_aws_account" "complete_example" {
  # Required fields
  name      = "Terraform Complete Example Account"
  payer_id  = 1

  # Optional fields for account creation/import
  account_alias            = "tf-complete-example"
  account_number          = "123456789012"  # Include to import existing account
  account_type_id         = 1
  commercial_account_name = "Complete Example Commercial"
  create_govcloud         = true
  email                   = "root@example.com"
  gov_account_name        = "Complete Example GovCloud"
  include_linked_account_spend = true
  linked_role             = "OrganizationAccountAccessRole"
  project_id              = 42
  skip_access_checking    = false
  start_datecode         = "2024-03"
  use_org_account_info    = true

  # AWS Organizational Unit configuration
  aws_organizational_unit {
    name        = "Development"
    org_unit_id = "ou-1234-5678abcd"
  }

  # Settings for moving account between projects
  move_project_settings {
    financials     = "move"
    move_datecode  = 202403
  }

  # Labels for the account
  labels = {
    environment = "production"
    team        = "platform"
    cost_center = "12345"
  }
}

# Example of importing an existing account into a project
resource "kion_aws_account" "import_example" {
  name           = "Terraform Import Example"
  payer_id       = 1
  account_number = "987654321098"
  project_id     = 43
  start_datecode = "2024-03"
}

# Example of creating an account in the account cache
resource "kion_aws_account" "cache_example" {
  name                    = "Terraform Cache Example"
  payer_id                = 1
  commercial_account_name = "Cache Account"
  create_govcloud         = false
}

# Output examples
output "complete_example_id" {
  value = kion_aws_account.complete_example.id
}

output "complete_example_car_external_id" {
  value = kion_aws_account.complete_example.car_external_id
}

output "complete_example_service_external_id" {
  value = kion_aws_account.complete_example.service_external_id
}
