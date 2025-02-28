# Example 1: Create a new Azure MCA subscription
resource "kion_azure_account" "mca_example" {
  # Required fields
  name              = "MCA Subscription Example"
  payer_id          = 1
  subscription_name = "terraform-mca-subscription"

  # MCA-specific configuration
  mca {
    billing_account         = "5e98e158-xxxx-xxxx-xxxx-xxxxxxxxxxxx:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx_xxxx-xx-xx"
    billing_profile         = "AW4F-xxxx-xxx-xxx"
    billing_profile_invoice = "SH3V-xxxx-xxx-xxx"
  }

  # Optional fields
  account_alias    = "mca-dev"
  account_type_id  = 2
  project_id       = 42  # Add to project instead of cache
  start_datecode   = "2024-03"  # Required when adding to project
}

# Example 2: Create a new Azure EA subscription
resource "kion_azure_account" "ea_example" {
  name              = "EA Subscription Example"
  payer_id          = 1
  subscription_name = "terraform-ea-subscription"

  # EA-specific configuration
  ea {
    account          = "12345678"
    billing_account  = "98765432"
  }

  # Optional: Place under specific management group
  parent_management_group_id = "mg-platform-prod"

  # Labels for organization
  labels = {
    environment = "production"
    department  = "engineering"
    cost_center = "12345"
  }
}

# Example 3: Create a new Azure CSP subscription
resource "kion_azure_account" "csp_example" {
  name              = "CSP Subscription Example"
  payer_id          = 1
  subscription_name = "terraform-csp-subscription"

  # CSP-specific configuration
  csp {
    offer_id       = "MS-AZR-0146P"
    billing_cycle  = "Monthly"  # Optional
  }

  # Optional configuration
  skip_access_checking = true
}

# Example 4: Import an existing Azure subscription
resource "kion_azure_account" "import_example" {
  name              = "Imported Subscription"
  payer_id          = 1
  subscription_uuid = "12345678-90ab-cdef-1234-567890abcdef"
  project_id        = 43
  start_datecode    = "2024-03"
}

# Example 5: Account with move project settings
resource "kion_azure_account" "movable_example" {
  name              = "Movable Subscription"
  payer_id          = 1
  subscription_name = "terraform-movable-sub"

  mca {
    billing_account         = "5e98e158-xxxx-xxxx-xxxx-xxxxxxxxxxxx:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx_xxxx-xx-xx"
    billing_profile         = "AW4F-xxxx-xxx-xxx"
    billing_profile_invoice = "SH3V-xxxx-xxx-xxx"
  }

  # Initial project placement
  project_id     = 44
  start_datecode = "2024-03"

  # Settings for when moving between projects
  move_project_settings {
    financials     = "move"  # or "preserve"
    move_datecode  = 202403  # Only move finances from March 2024 onwards
  }
}

# Output examples
output "mca_account_id" {
  description = "ID of the MCA subscription"
  value       = kion_azure_account.mca_example.id
}

output "ea_account_id" {
  description = "ID of the EA subscription"
  value       = kion_azure_account.ea_example.id
}

output "csp_account_id" {
  description = "ID of the CSP subscription"
  value       = kion_azure_account.csp_example.id
}

output "import_account_location" {
  description = "Location of the imported subscription (project or cache)"
  value       = kion_azure_account.import_example.location
}

output "account_creation_time" {
  description = "When the MCA subscription was created"
  value       = kion_azure_account.mca_example.created_at
}
