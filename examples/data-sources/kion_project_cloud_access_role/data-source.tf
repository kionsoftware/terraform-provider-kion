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

# Example using the data source to create a similar role in another project.
# This copies a User role. Copying a non-User role would also need to carry over
# cloud_access_role_type_id and that type's trust attributes.
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

# Cloud access role types
#
# 1 = User (default, multi-cloud, assumable by Kion users and groups)
# 2 = Custom Trust (AWS only, custom trust policy)
# 3 = Account      (AWS only, trusts another AWS account)
# 4 = Service      (AWS only, trusts AWS service principals)
data "kion_project_cloud_access_role" "custom_trust" {
  id = "20"
}

output "custom_trust_details" {
  value = {
    role_type        = data.kion_project_cloud_access_role.custom_trust.cloud_access_role_type_id
    trust_policy     = data.kion_project_cloud_access_role.custom_trust.aws_iam_role_trust_policy
    trusted_accounts = data.kion_project_cloud_access_role.custom_trust.aws_trusted_account_numbers
    trusted_services = data.kion_project_cloud_access_role.custom_trust.aws_trusted_services
    aws_partition    = data.kion_project_cloud_access_role.custom_trust.aws_partition
    instance_profile = data.kion_project_cloud_access_role.custom_trust.aws_create_instance_profile
    session_tags     = data.kion_project_cloud_access_role.custom_trust.aws_session_tags
  }
}

# Only User roles accept Kion users and user groups, so branching on the type
# tells you whether the users and user_groups attributes can be populated.
output "accepts_kion_users" {
  value = data.kion_project_cloud_access_role.custom_trust.cloud_access_role_type_id == 1
}

# Note: cloud_provider_ids is not exposed on this data source. Kion accepts it
# when creating or updating a role but omits it when reading one back.
