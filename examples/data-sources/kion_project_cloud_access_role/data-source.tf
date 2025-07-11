# Basic example - Look up a project cloud access role by ID
data "kion_project_cloud_access_role" "example" {
  id = "1"
}

# Output the retrieved cloud access role information
output "project_car_name" {
  value = data.kion_project_cloud_access_role.example.name
}

output "project_car_project_id" {
  value = data.kion_project_cloud_access_role.example.project_id
}

output "project_car_applies_to_all" {
  value = data.kion_project_cloud_access_role.example.apply_to_all_accounts
}

# Example using the data source to create a similar role in another project
resource "kion_project_cloud_access_role" "copy_role" {
  name                   = "Copy of ${data.kion_project_cloud_access_role.example.name}"
  project_id             = 2 # Different project
  aws_iam_role_name      = data.kion_project_cloud_access_role.example.aws_iam_role_name
  aws_iam_path           = data.kion_project_cloud_access_role.example.aws_iam_path
  web_access             = data.kion_project_cloud_access_role.example.web_access
  short_term_access_keys = data.kion_project_cloud_access_role.example.short_term_access_keys
  long_term_access_keys  = data.kion_project_cloud_access_role.example.long_term_access_keys
  apply_to_all_accounts  = data.kion_project_cloud_access_role.example.apply_to_all_accounts
  future_accounts        = data.kion_project_cloud_access_role.example.future_accounts

  # Copy the same AWS IAM policies
  dynamic "aws_iam_policies" {
    for_each = data.kion_project_cloud_access_role.example.aws_iam_policies
    content {
      id = aws_iam_policies.value.id
    }
  }

  # Copy the same user groups
  dynamic "user_groups" {
    for_each = data.kion_project_cloud_access_role.example.user_groups
    content {
      id = user_groups.value.id
    }
  }
}

# Look up an AWS-focused project cloud access role
data "kion_project_cloud_access_role" "aws_admin" {
  id = "10"
}

# Output detailed AWS information
output "aws_project_role_details" {
  value = {
    name                   = data.kion_project_cloud_access_role.aws_admin.name
    project_id             = data.kion_project_cloud_access_role.aws_admin.project_id
    aws_iam_role_name      = data.kion_project_cloud_access_role.aws_admin.aws_iam_role_name
    aws_iam_path           = data.kion_project_cloud_access_role.aws_admin.aws_iam_path
    apply_to_all_accounts  = data.kion_project_cloud_access_role.aws_admin.apply_to_all_accounts
    future_accounts        = data.kion_project_cloud_access_role.aws_admin.future_accounts
    web_access             = data.kion_project_cloud_access_role.aws_admin.web_access
    short_term_access_keys = data.kion_project_cloud_access_role.aws_admin.short_term_access_keys
    long_term_access_keys  = data.kion_project_cloud_access_role.aws_admin.long_term_access_keys
    account_count          = length(data.kion_project_cloud_access_role.aws_admin.accounts)
    aws_policy_count       = length(data.kion_project_cloud_access_role.aws_admin.aws_iam_policies)
    azure_role_count       = length(data.kion_project_cloud_access_role.aws_admin.azure_role_definitions)
    gcp_role_count         = length(data.kion_project_cloud_access_role.aws_admin.gcp_iam_roles)
  }
}

# Example showing account management details
locals {
  # Extract account IDs from the cloud access role
  account_ids = [for account in data.kion_project_cloud_access_role.aws_admin.accounts : account.id]

  # Extract policy IDs by cloud provider
  aws_policy_ids = [for policy in data.kion_project_cloud_access_role.aws_admin.aws_iam_policies : policy.id]
  azure_role_ids = [for role in data.kion_project_cloud_access_role.aws_admin.azure_role_definitions : role.id]
  gcp_role_ids   = [for role in data.kion_project_cloud_access_role.aws_admin.gcp_iam_roles : role.id]

  # Extract user and group IDs
  user_ids       = [for user in data.kion_project_cloud_access_role.aws_admin.users : user.id]
  user_group_ids = [for group in data.kion_project_cloud_access_role.aws_admin.user_groups : group.id]
}

output "associated_accounts" {
  value = local.account_ids
}

output "cloud_permissions_summary" {
  value = {
    aws_policies   = local.aws_policy_ids
    azure_roles    = local.azure_role_ids
    gcp_roles      = local.gcp_role_ids
    total_policies = length(local.aws_policy_ids) + length(local.azure_role_ids) + length(local.gcp_role_ids)
  }
}

output "assigned_users_and_groups" {
  value = {
    users       = local.user_ids
    user_groups = local.user_group_ids
  }
}

# Multi-cloud role example
data "kion_project_cloud_access_role" "multi_cloud" {
  id = "15"
}

# Conditional output based on cloud providers used
output "multi_cloud_analysis" {
  value = {
    role_name = data.kion_project_cloud_access_role.multi_cloud.name
    has_aws   = length(data.kion_project_cloud_access_role.multi_cloud.aws_iam_policies) > 0
    has_azure = length(data.kion_project_cloud_access_role.multi_cloud.azure_role_definitions) > 0
    has_gcp   = length(data.kion_project_cloud_access_role.multi_cloud.gcp_iam_roles) > 0
    cloud_providers = compact([
      length(data.kion_project_cloud_access_role.multi_cloud.aws_iam_policies) > 0 ? "AWS" : "",
      length(data.kion_project_cloud_access_role.multi_cloud.azure_role_definitions) > 0 ? "Azure" : "",
      length(data.kion_project_cloud_access_role.multi_cloud.gcp_iam_roles) > 0 ? "GCP" : ""
    ])
  }
}
