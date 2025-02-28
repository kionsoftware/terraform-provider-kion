# Create a cloud access role on an OU.
resource "kion_ou_cloud_access_role" "carou1" {
  name                   = "sample-car"
  ou_id                  = 3
  aws_iam_role_name      = "sample-car"
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = true
  aws_iam_policies { id = 628 }
  #aws_iam_permissions_boundary = 1
  users { id = 1 }
  user_groups { id = 1 }
}

# Output the ID of the resource created.
output "ou_car_id" {
  value = kion_ou_cloud_access_role.carou1.id
}

# Create cloud access roles for different cloud providers
resource "kion_ou_cloud_access_role" "aws_admin_role" {
  name                   = "aws-admin-role"
  ou_id                  = 3
  aws_iam_role_name      = "KionAWSAdminRole"
  aws_iam_path           = "/kion/admin/"
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = false

  # Attach multiple AWS IAM policies
  aws_iam_policies {
    id = 100  # AdminAccess policy
  }
  aws_iam_policies {
    id = 101  # SecurityAudit policy
  }

  # Set permissions boundary
  aws_iam_permissions_boundary = 200

  # Assign to specific users
  users {
    id = 15  # Senior AWS Admin
  }

  # Assign to groups
  user_groups {
    id = 25  # AWS Administrators
  }
}

# Create a role with Azure permissions
resource "kion_ou_cloud_access_role" "azure_admin_role" {
  name              = "azure-admin-role"
  ou_id             = 3
  aws_iam_role_name = "KionAzureAdminRole"  # Required even for Azure roles
  web_access        = true

  # Azure role definitions
  azure_role_definitions {
    id = 300  # Azure Administrator
  }

  # Assign to specific groups
  user_groups {
    id = 26  # Azure Administrators
  }
}

# Create a role with GCP permissions
resource "kion_ou_cloud_access_role" "gcp_admin_role" {
  name              = "gcp-admin-role"
  ou_id             = 3
  aws_iam_role_name = "KionGCPAdminRole"  # Required even for GCP roles
  web_access        = true

  # GCP IAM roles
  gcp_iam_roles {
    id = 400  # GCP Administrator
  }

  # Assign to both users and groups
  users {
    id = 16  # GCP Admin
  }
  user_groups {
    id = 27  # GCP Administrators
  }
}

# Output the role IDs
output "aws_admin_role_id" {
  value       = kion_ou_cloud_access_role.aws_admin_role.id
  description = "The ID of the AWS admin role"
}

output "azure_admin_role_id" {
  value       = kion_ou_cloud_access_role.azure_admin_role.id
  description = "The ID of the Azure admin role"
}

output "gcp_admin_role_id" {
  value       = kion_ou_cloud_access_role.gcp_admin_role.id
  description = "The ID of the GCP admin role"
}
