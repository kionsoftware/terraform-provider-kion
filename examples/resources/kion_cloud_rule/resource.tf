# Create a complete cloud rule with all available options
resource "kion_cloud_rule" "complete_example" {
  name        = "Complete Cloud Rule Example"
  description = "A comprehensive example of a cloud rule with all supported options."

  # Concurrent CFT deployment setting
  concurrent_cft_sync = true

  # AWS-specific configurations
  aws_iam_policies {
    id = 1
  }

  aws_cloudformation_templates {
    id = 1
  }

  internal_aws_amis {
    id = 1
  }

  internal_aws_service_catalog_portfolios {
    id = 1
  }

  service_control_policies {
    id = 1
  }

  # Azure-specific configurations
  azure_policy_definitions {
    id = 1
  }

  azure_role_definitions {
    id = 1
  }

  azure_arm_template_definitions {
    id = 1
  }

  # GCP-specific configurations
  gcp_iam_roles {
    id = 1
  }

  # Compliance configurations
  compliance_standards {
    id = 1
  }

  # Webhook configurations
  pre_webhook_id  = 1
  post_webhook_id = 2

  # Ownership and scope
  owner_users {
    id = 1
  }

  owner_user_groups {
    id = 2
  }

  projects {
    id = 1
  }

  ous {
    id = 1
  }

  # Labels
  labels = {
    environment = "production"
    team        = "security"
    cost_center = "12345"
  }
}

# Create a simple AWS-focused cloud rule
resource "kion_cloud_rule" "aws_example" {
  name        = "AWS Security Baseline"
  description = "Basic AWS security policies and templates"

  aws_iam_policies {
    id = 1
  }

  service_control_policies {
    id = 1
  }

  owner_users {
    id = 1
  }
}

# Create an Azure-focused cloud rule
resource "kion_cloud_rule" "azure_example" {
  name        = "Azure Governance"
  description = "Azure governance policies and roles"

  azure_policy_definitions {
    id = 1
  }

  azure_role_definitions {
    id = 1
  }

  owner_user_groups {
    id = 1
  }
}

# Output examples
output "complete_rule_id" {
  value = kion_cloud_rule.complete_example.id
}

output "aws_rule_id" {
  value = kion_cloud_rule.aws_example.id
}

output "azure_rule_id" {
  value = kion_cloud_rule.azure_example.id
}
