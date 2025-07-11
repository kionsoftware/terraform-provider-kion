# Basic OU Cloud Access Role Example
resource "kion_ou_cloud_access_role" "example" {
  name   = "example-ou-car"
  ou_id  = 1
  
  # Basic access permissions
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = false
  
  # Assign to users and groups
  users {
    id = 1
  }
  user_groups {
    id = 1
  }
}

# AWS-focused OU Cloud Access Role
resource "kion_ou_cloud_access_role" "aws_admin" {
  name              = "aws-admin-role"
  ou_id             = 1
  aws_iam_role_name = "AdminRole"  # Only needed if this role will be used for AWS accounts
  aws_iam_path      = "/kion/"
  
  # AWS access types
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = false
  
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

# Azure-focused OU Cloud Access Role
resource "kion_ou_cloud_access_role" "azure_admin" {
  name       = "azure-admin-role"
  ou_id      = 1
  web_access = true
  
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

# GCP-focused OU Cloud Access Role
resource "kion_ou_cloud_access_role" "gcp_admin" {
  name       = "gcp-admin-role"
  ou_id      = 1
  web_access = true
  
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

# Multi-cloud OU Cloud Access Role
resource "kion_ou_cloud_access_role" "multi_cloud" {
  name              = "multi-cloud-role"
  ou_id             = 1
  aws_iam_role_name = "CrossAccountRole"  # Only needed because this role includes AWS permissions
  
  # Access types
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = false
  
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

# Outputs
output "example_id" {
  value = kion_ou_cloud_access_role.example.id
}

output "aws_admin_id" {
  value = kion_ou_cloud_access_role.aws_admin.id
}

output "azure_admin_id" {
  value = kion_ou_cloud_access_role.azure_admin.id
}

output "gcp_admin_id" {
  value = kion_ou_cloud_access_role.gcp_admin.id
}

output "multi_cloud_id" {
  value = kion_ou_cloud_access_role.multi_cloud.id
}