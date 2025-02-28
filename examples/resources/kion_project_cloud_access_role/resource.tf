# Create AWS cloud access role for development project
resource "kion_project_cloud_access_role" "dev_aws_role" {
  name                   = "dev-aws-admin"
  project_id             = 10
  aws_iam_role_name      = "DevAWSAdmin"
  aws_iam_path           = "/kion/dev/"
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = false

  # Apply to all current and future accounts
  apply_to_all_accounts = true
  future_accounts       = true

  # AWS IAM policies
  aws_iam_policies {
    id = 100  # Developer access policy
  }
  aws_iam_policies {
    id = 101  # CloudWatch access policy
  }

  aws_iam_permissions_boundary = 200  # Developer boundary

  # Assign to development team
  users {
    id = 15  # Lead Developer
  }
  user_groups {
    id = 25  # Development Team
  }
}

# Create Azure cloud access role for production project
resource "kion_project_cloud_access_role" "prod_azure_role" {
  name                   = "prod-azure-admin"
  project_id             = 11
  aws_iam_role_name      = "ProdAzureAdmin"  # Required even for Azure roles
  web_access             = true

  # Apply to specific accounts
  accounts {
    id = 50  # Production Azure subscription
  }
  accounts {
    id = 51  # DR Azure subscription
  }

  # Azure role definitions
  azure_role_definitions {
    id = 300  # Azure Administrator
  }
  azure_role_definitions {
    id = 301  # Azure Security Admin
  }

  # Assign to production team
  users {
    id = 20  # Production Lead
  }
  user_groups {
    id = 30  # Production Team
  }
}

# Create GCP cloud access role for analytics project
resource "kion_project_cloud_access_role" "analytics_gcp_role" {
  name                   = "analytics-gcp-admin"
  project_id             = 12
  aws_iam_role_name      = "AnalyticsGCPAdmin"  # Required even for GCP roles
  web_access             = true

  # Apply to specific accounts and future accounts
  accounts {
    id = 70  # Analytics GCP project
  }
  future_accounts = true

  # GCP IAM roles
  gcp_iam_roles {
    id = 400  # BigQuery Admin
  }
  gcp_iam_roles {
    id = 401  # Storage Admin
  }

  # Assign to analytics team
  users {
    id = 25  # Analytics Lead
  }
  user_groups {
    id = 35  # Analytics Team
  }
}

# Output role information
output "dev_aws_role_id" {
  value       = kion_project_cloud_access_role.dev_aws_role.id
  description = "Development AWS role ID"
}

output "prod_azure_role_id" {
  value       = kion_project_cloud_access_role.prod_azure_role.id
  description = "Production Azure role ID"
}

output "analytics_gcp_role_id" {
  value       = kion_project_cloud_access_role.analytics_gcp_role.id
  description = "Analytics GCP role ID"
}
