# List all OU Cloud Access Role Exemptions
data "kion_ou_cloud_access_role_exemption" "all" {}

# Filter by specific OU Cloud Access Role
data "kion_ou_cloud_access_role_exemption" "by_role" {
  filter {
    ou_cloud_access_role_id = 107
  }
}

# Filter by specific OU
data "kion_ou_cloud_access_role_exemption" "by_ou" {
  filter {
    ou_id = 104
  }
}

# Filter by both OU and Cloud Access Role
data "kion_ou_cloud_access_role_exemption" "specific" {
  filter {
    ou_cloud_access_role_id = 107
    ou_id                   = 104
  }
}