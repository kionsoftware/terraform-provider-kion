# Read current Kion application configuration
data "kion_app_config" "current" {
}

# Example outputs to demonstrate available configuration settings
output "budget_mode_enabled" {
  description = "Whether budget mode is enabled"
  value       = data.kion_app_config.current.budget_mode
}

output "default_org_chart_view" {
  description = "Default organization chart view setting"
  value       = data.kion_app_config.current.default_org_chart_view
}

output "cost_savings_settings" {
  description = "Cost savings configuration"
  value = {
    enabled           = data.kion_app_config.current.cost_savings_enabled
    allow_terminate   = data.kion_app_config.current.cost_savings_allow_terminate
    post_token_hours  = data.kion_app_config.current.cost_savings_post_token_life_hours
  }
}

output "smtp_configuration" {
  description = "SMTP server configuration (password excluded)"
  value = {
    enabled     = data.kion_app_config.current.smtp_enabled
    host        = data.kion_app_config.current.smtp_host
    port        = data.kion_app_config.current.smtp_port
    from        = data.kion_app_config.current.smtp_from
    username    = data.kion_app_config.current.smtp_username
    skip_verify = data.kion_app_config.current.smtp_skip_verify
  }
}

output "feature_flags" {
  description = "Various feature flag settings"
  value = {
    allocation_mode              = data.kion_app_config.current.allocation_mode
    forecasting_enabled          = data.kion_app_config.current.forecasting_enabled
    resource_inventory_enabled   = data.kion_app_config.current.resource_inventory_enabled
    reserved_instances_enabled   = data.kion_app_config.current.reserved_instances_enabled
    event_driven_enabled         = data.kion_app_config.current.event_driven_enabled
    saml_debug                   = data.kion_app_config.current.saml_debug
  }
}

output "api_key_settings" {
  description = "App API key configuration"
  value = {
    creation_enabled = data.kion_app_config.current.app_api_key_creation_enabled
    lifespan_days    = data.kion_app_config.current.app_api_key_lifespan
    limit_per_user   = data.kion_app_config.current.app_api_key_limit
  }
}

output "aws_settings" {
  description = "AWS-related configuration"
  value = {
    access_key_creation_enabled = data.kion_app_config.current.aws_access_key_creation_enabled
    supported_regions           = data.kion_app_config.current.supported_aws_regions
  }
}

output "permission_settings" {
  description = "Permission and access configuration"
  value = {
    all_users_see_ou_names           = data.kion_app_config.current.all_users_see_ou_names
    allow_custom_permission_schemes  = data.kion_app_config.current.allow_custom_permission_schemes
    cloud_rule_group_ownership_only  = data.kion_app_config.current.cloud_rule_group_ownership_only
    idms_groups_as_viewers_default   = data.kion_app_config.current.idms_groups_as_viewers_default
  }
}

output "funding_enforcement" {
  description = "Funding and budget enforcement settings"
  value = {
    enforce_funding         = data.kion_app_config.current.enforce_funding
    enforce_funding_sources = data.kion_app_config.current.enforce_funding_sources
  }
}