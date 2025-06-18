# List all Project Cloud Access Role Exemptions
data "kion_project_cloud_access_role_exemption" "all" {}

# Filter by specific OU Cloud Access Role
data "kion_project_cloud_access_role_exemption" "by_role" {
  filter {
    ou_cloud_access_role_id = 107
  }
}

# Filter by specific Project
data "kion_project_cloud_access_role_exemption" "by_project" {
  filter {
    project_id = 42
  }
}

# Filter by both Project and Cloud Access Role
data "kion_project_cloud_access_role_exemption" "specific" {
  filter {
    ou_cloud_access_role_id = 107
    project_id              = 42
  }
}