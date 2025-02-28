# Create a new GCP project with all available options
resource "kion_gcp_account" "complete_example" {
  name                    = "Production GCP Project"
  create_mode            = "create"
  google_cloud_project_id = "prod-project-123456"
  payer_id               = 1

  # Optional configurations
  account_alias          = "prod-gcp"
  account_type_id        = 1
  google_cloud_parent_name = "folders/987654321"
  skip_access_checking   = false

  # Project placement and financial settings
  project_id             = 42
  start_datecode        = "2024-01"

  # Settings for moving between projects
  move_project_settings {
    financials    = "move"
    move_datecode = 202401
  }

  # Labels for categorization
  labels = {
    environment = "production"
    team        = "platform"
    cost_center = "12345"
  }
}

# Import an existing GCP project into a Kion project
resource "kion_gcp_account" "import_example" {
  name                    = "Existing Dev Project"
  create_mode            = "import"
  google_cloud_project_id = "dev-project-654321"
  payer_id               = 1
  project_id             = 43
  start_datecode        = "2024-01"

  labels = {
    environment = "development"
    imported    = "true"
  }
}

# Create a new GCP project in the account cache
resource "kion_gcp_account" "cache_example" {
  name                    = "Staging GCP Project"
  create_mode            = "create"
  google_cloud_project_id = "staging-project-789012"
  payer_id               = 1

  account_alias          = "staging-gcp"
  google_cloud_parent_name = "organizations/123456789"

  labels = {
    environment = "staging"
    managed_by  = "kion"
  }
}

# Example of moving a project between Kion projects
resource "kion_gcp_account" "move_example" {
  name                    = "Project to Move"
  create_mode            = "create"
  google_cloud_project_id = "moving-project-345678"
  payer_id               = 1
  project_id             = 44
  start_datecode        = "2024-01"

  move_project_settings {
    financials    = "preserve"  # Keep financial history in the old project
  }
}

# Output examples
output "prod_project_id" {
  value = kion_gcp_account.complete_example.id
}

output "prod_project_location" {
  value = kion_gcp_account.complete_example.location
}

output "imported_project_id" {
  value = kion_gcp_account.import_example.id
}

output "staging_project_created_at" {
  value = kion_gcp_account.cache_example.created_at
}
