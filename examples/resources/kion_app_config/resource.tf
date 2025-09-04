# Configure Kion application settings
# Note: This resource manages global application configuration.
# Only specify the settings you want to manage; unspecified settings will be left unchanged.

resource "kion_app_config" "main" {
  # Budget and cost management settings
  budget_mode                        = true
  allocation_mode                    = true
  enforce_funding                    = true
  enforce_funding_sources           = true
  cost_savings_enabled              = true
  cost_savings_allow_terminate      = false
  cost_savings_post_token_life_hours = 24

  # Organization and UI settings
  default_org_chart_view    = "policy"
  all_users_see_ou_names   = true

  # Feature enablement
  forecasting_enabled        = true
  resource_inventory_enabled = true
  reserved_instances_enabled = true
  event_driven_enabled      = false

  # API key management
  app_api_key_creation_enabled = true
  app_api_key_lifespan        = 90  # 90 days
  app_api_key_limit           = 5   # 5 keys per user

  # AWS configuration
  aws_access_key_creation_enabled = true
  supported_aws_regions = [
    "us-east-1",
    "us-west-2",
    "eu-west-1",
    "ap-southeast-1"
  ]

  # Permission and access settings
  allow_custom_permission_schemes  = true
  cloud_rule_group_ownership_only  = false
  idms_groups_as_viewers_default   = false

  # SMTP configuration for email notifications
  smtp_enabled     = true
  smtp_host        = "smtp.company.com"
  smtp_port        = 587
  smtp_from        = "kion-notifications@company.com"
  smtp_username    = "kion-service"
  smtp_password    = var.smtp_password  # Use a variable for sensitive data
  smtp_skip_verify = false

  # Debug and development settings
  saml_debug = false
}

# Example with minimal configuration - only enabling key features
resource "kion_app_config" "minimal" {
  # Only configure essential settings
  budget_mode         = true
  enforce_funding     = true
  forecasting_enabled = true
  
  # Basic SMTP for notifications
  smtp_enabled = true
  smtp_host    = "localhost"
  smtp_port    = 25
  smtp_from    = "noreply@company.com"
}

# Variable for sensitive SMTP password
variable "smtp_password" {
  description = "SMTP server password for email notifications"
  type        = string
  sensitive   = true
}

# Outputs to show current configuration
output "app_config_summary" {
  description = "Summary of key application configuration settings"
  value = {
    budget_mode_enabled    = kion_app_config.main.budget_mode
    cost_savings_enabled   = kion_app_config.main.cost_savings_enabled
    default_chart_view     = kion_app_config.main.default_org_chart_view
    smtp_configured        = kion_app_config.main.smtp_enabled
    supported_aws_regions  = length(kion_app_config.main.supported_aws_regions)
    api_key_limit_per_user = kion_app_config.main.app_api_key_limit
  }
}

# Example of importing existing app-config
# terraform import kion_app_config.existing app-config

# Example of using data source to reference current settings
data "kion_app_config" "current" {}

output "current_vs_managed" {
  description = "Comparison of current settings vs managed settings"
  value = {
    current_budget_mode = data.kion_app_config.current.budget_mode
    managed_budget_mode = kion_app_config.main.budget_mode
    settings_match      = data.kion_app_config.current.budget_mode == kion_app_config.main.budget_mode
  }
}