# Basic example - Look up an OU cloud access role by ID
data "kion_ou_cloud_access_role" "example" {
  id = 1
}

# Output the retrieved cloud access role information
output "ou_car_name" {
  value = data.kion_ou_cloud_access_role.example.name
}

output "ou_car_ou_id" {
  value = data.kion_ou_cloud_access_role.example.ou_id
}

output "ou_car_aws_role_name" {
  value = data.kion_ou_cloud_access_role.example.aws_iam_role_name
}

# Example using the data source to reference in another resource
resource "kion_project" "example_project" {
  name                 = "Project using OU CAR ${data.kion_ou_cloud_access_role.example.name}"
  ou_id                = data.kion_ou_cloud_access_role.example.ou_id
  permission_scheme_id = 1

  # Use the same user groups as the OU cloud access role
  owner_user_group_ids {
    id = tolist(data.kion_ou_cloud_access_role.example.user_groups)[0].id
  }
}

# Look up a specific OU cloud access role and display its permissions
data "kion_ou_cloud_access_role" "aws_admin" {
  id = 5
}

# Output AWS-specific information
output "aws_admin_details" {
  value = {
    name                   = data.kion_ou_cloud_access_role.aws_admin.name
    aws_iam_role_name      = data.kion_ou_cloud_access_role.aws_admin.aws_iam_role_name
    aws_iam_path           = data.kion_ou_cloud_access_role.aws_admin.aws_iam_path
    web_access             = data.kion_ou_cloud_access_role.aws_admin.web_access
    short_term_access_keys = data.kion_ou_cloud_access_role.aws_admin.short_term_access_keys
    long_term_access_keys  = data.kion_ou_cloud_access_role.aws_admin.long_term_access_keys
    aws_policy_count       = length(data.kion_ou_cloud_access_role.aws_admin.aws_iam_policies)
  }
}

# Example showing how to iterate over associated resources
locals {
  # Extract AWS IAM policy IDs from the cloud access role
  aws_policy_ids = [for policy in data.kion_ou_cloud_access_role.aws_admin.aws_iam_policies : policy.id]

  # Extract user group IDs
  user_group_ids = [for group in data.kion_ou_cloud_access_role.aws_admin.user_groups : group.id]
}

output "associated_policy_ids" {
  value = local.aws_policy_ids
}

output "associated_user_group_ids" {
  value = local.user_group_ids
}

# Cloud access role types
#
# 1 = User (default, multi-cloud, assumable by Kion users and groups)
# 2 = Custom Trust (AWS only, custom trust policy)
# 3 = Account      (AWS only, trusts another AWS account)
# 4 = Service      (AWS only, trusts AWS service principals)
data "kion_ou_cloud_access_role" "service_role" {
  id = 8
}

output "service_role_details" {
  value = {
    role_type        = data.kion_ou_cloud_access_role.service_role.cloud_access_role_type_id
    trust_policy     = data.kion_ou_cloud_access_role.service_role.aws_iam_role_trust_policy
    trusted_accounts = data.kion_ou_cloud_access_role.service_role.aws_trusted_account_numbers
    trusted_services = data.kion_ou_cloud_access_role.service_role.aws_trusted_services
    aws_partition    = data.kion_ou_cloud_access_role.service_role.aws_partition
    instance_profile = data.kion_ou_cloud_access_role.service_role.aws_create_instance_profile
    session_tags     = data.kion_ou_cloud_access_role.service_role.aws_session_tags
  }
}

# Only User roles accept Kion users and user groups, so branching on the type
# tells you whether the users and user_groups attributes can be populated.
output "ou_role_accepts_kion_users" {
  value = data.kion_ou_cloud_access_role.service_role.cloud_access_role_type_id == 1
}
