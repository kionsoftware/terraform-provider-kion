# Basic Project Cloud Access Role Example
resource "kion_project_cloud_access_role" "example" {
  name       = "example-project-car"
  project_id = 1

  # Basic access permissions
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = false

  # Apply to all accounts
  apply_to_all_accounts = true

  # Assign to users and groups
  users {
    id = 1
  }
  user_groups {
    id = 1
  }
}

# AWS-focused Project Cloud Access Role
resource "kion_project_cloud_access_role" "aws_admin" {
  name              = "aws-admin-role"
  project_id        = 1
  aws_iam_role_name = "AdminRole" # Only needed if this role will be used for AWS accounts
  aws_iam_path      = "/kion/"

  # AWS access types
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = false

  # Apply to specific accounts
  accounts {
    id = 1
  }
  accounts {
    id = 2
  }

  # Include future accounts
  future_accounts = true

  # AWS IAM policies
  aws_iam_policies {
    id = 1
  }
  aws_iam_policies {
    id = 2
  }

  # AWS permissions boundary
  aws_iam_permissions_boundary = 1

  # Assign to users and groups
  users {
    id = 1
  }
  user_groups {
    id = 1
  }
}

# Azure-focused Project Cloud Access Role
resource "kion_project_cloud_access_role" "azure_admin" {
  name       = "azure-admin-role"
  project_id = 2
  web_access = true

  # Apply to specific accounts
  accounts {
    id = 3
  }
  accounts {
    id = 4
  }

  # Azure role definitions
  azure_role_definitions {
    id = 1
  }
  azure_role_definitions {
    id = 2
  }

  # Assign to groups
  user_groups {
    id = 2
  }
}

# GCP-focused Project Cloud Access Role
resource "kion_project_cloud_access_role" "gcp_admin" {
  name       = "gcp-admin-role"
  project_id = 3
  web_access = true

  # Apply to all accounts and future accounts
  apply_to_all_accounts = true
  future_accounts       = true

  # GCP IAM roles
  gcp_iam_roles {
    id = 1
  }
  gcp_iam_roles {
    id = 2
  }

  # Assign to users and groups
  users {
    id = 2
  }
  user_groups {
    id = 2
  }
}

# Multi-cloud Project Cloud Access Role
resource "kion_project_cloud_access_role" "multi_cloud" {
  name              = "multi-cloud-role"
  project_id        = 4
  aws_iam_role_name = "CrossAccountRole" # Only needed because this role includes AWS permissions

  # Access types
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = false

  # Apply to all accounts
  apply_to_all_accounts = true
  future_accounts       = true

  # AWS permissions
  aws_iam_policies {
    id = 1
  }

  # Azure permissions
  azure_role_definitions {
    id = 1
  }

  # GCP permissions
  gcp_iam_roles {
    id = 1
  }

  # Assign to users
  users {
    id = 1
  }
}

# Custom Trust role (cloud_access_role_type_id = 2). AWS only. The trust policy
# is supplied verbatim and the role cannot be assigned to Kion users or groups.
resource "kion_project_cloud_access_role" "custom_trust" {
  name                      = "custom-trust-role"
  project_id                = 1
  aws_iam_role_name         = "CustomTrustRole"
  cloud_access_role_type_id = 2

  aws_iam_role_trust_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { AWS = "arn:aws:iam::123456789012:root" }
      Action    = "sts:AssumeRole"
    }]
  })

  aws_iam_policies {
    id = 1
  }
}

# Account role (cloud_access_role_type_id = 3). Trusts a single AWS account.
resource "kion_project_cloud_access_role" "trusted_account" {
  name                        = "trusted-account-role"
  project_id                  = 1
  aws_iam_role_name           = "TrustedAccountRole"
  cloud_access_role_type_id   = 3
  aws_trusted_account_numbers = ["123456789012"]

  aws_iam_policies {
    id = 1
  }
}

# Service role (cloud_access_role_type_id = 4). Trusts AWS service principals.
# GovCloud and ISO partitions must set aws_partition explicitly.
resource "kion_project_cloud_access_role" "service" {
  name                        = "lambda-execution-role"
  project_id                  = 1
  aws_iam_role_name           = "LambdaExecutionRole"
  cloud_access_role_type_id   = 4
  aws_trusted_services        = ["lambda.amazonaws.com", "ec2.amazonaws.com"]
  aws_partition               = "aws"
  aws_create_instance_profile = true

  aws_iam_policies {
    id = 1
  }
}

# Session tags and an explicit cloud provider restriction on a User role.
resource "kion_project_cloud_access_role" "tagged" {
  name       = "tagged-role"
  project_id = 1
  web_access = true

  # 1 = AWS, 2 = Azure, 3 = GCP
  cloud_provider_ids = [1]

  aws_session_tags = {
    department  = "engineering"
    cost_center = "1234"
  }

  users {
    id = 1
  }
}

# Outputs
output "example_id" {
  value = kion_project_cloud_access_role.example.id
}

output "custom_trust_id" {
  value = kion_project_cloud_access_role.custom_trust.id
}

output "trusted_account_id" {
  value = kion_project_cloud_access_role.trusted_account.id
}

output "service_id" {
  value = kion_project_cloud_access_role.service.id
}

output "aws_admin_id" {
  value = kion_project_cloud_access_role.aws_admin.id
}

output "azure_admin_id" {
  value = kion_project_cloud_access_role.azure_admin.id
}

output "gcp_admin_id" {
  value = kion_project_cloud_access_role.gcp_admin.id
}

output "multi_cloud_id" {
  value = kion_project_cloud_access_role.multi_cloud.id
}