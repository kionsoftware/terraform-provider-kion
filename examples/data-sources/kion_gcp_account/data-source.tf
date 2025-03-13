# Get all GCP accounts
data "kion_gcp_account" "all" {
}

# Filter accounts by name
data "kion_gcp_account" "by_name" {
  filter {
    name   = "name"
    values = ["Production GCP Project"]
  }
}

# Filter by project ID
data "kion_gcp_account" "by_project" {
  filter {
    name   = "project_id"
    values = ["42"]
  }
}

# Filter by GCP project ID with regex
data "kion_gcp_account" "prod_projects" {
  filter {
    name   = "google_cloud_project_id"
    values = ["prod-.*"]
    regex  = true
  }
}

# Filter by location (project or cache)
data "kion_gcp_account" "cached_accounts" {
  filter {
    name   = "location"
    values = ["cache"]
  }
}

# Filter by multiple criteria
data "kion_gcp_account" "filtered_accounts" {
  filter {
    name   = "account_type_id"
    values = ["1"]
  }
  filter {
    name   = "skip_access_checking"
    values = ["false"]
  }
}

# Example outputs
output "all_accounts" {
  value = data.kion_gcp_account.all.list
}

output "prod_account_details" {
  value = data.kion_gcp_account.by_name.list[0]
}

output "project_accounts" {
  value = [
    for account in data.kion_gcp_account.by_project.list : {
      name                    = account.name
      google_cloud_project_id = account.google_cloud_project_id
      created_at             = account.created_at
    }
  ]
}

output "production_projects" {
  value = [
    for account in data.kion_gcp_account.prod_projects.list : {
      name       = account.name
      project_id = account.google_cloud_project_id
      alias      = account.account_alias
    }
  ]
}

output "cached_account_summary" {
  value = [
    for account in data.kion_gcp_account.cached_accounts.list : {
      name                     = account.name
      google_cloud_project_id  = account.google_cloud_project_id
      google_cloud_parent_name = account.google_cloud_parent_name
    }
  ]
}

output "filtered_account_ids" {
  value = [for account in data.kion_gcp_account.filtered_accounts.list : account.id]
}