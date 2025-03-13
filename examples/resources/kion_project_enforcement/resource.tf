# Create monthly budget enforcement for development project
resource "kion_project_enforcement" "dev_budget_enforcement" {
  project_id             = 10
  description           = "Monthly budget enforcement for development project"
  threshold             = 5000  # $5,000 threshold
  threshold_type        = "dollar"
  timeframe             = "month"
  enabled               = true
  overburn             = true
  notification_frequency = "daily"

  # Notify development team
  user_ids = [15, 16]  # Development leads
  user_group_ids = [25] # Development team group
}

# Create percentage-based enforcement for production project
resource "kion_project_enforcement" "prod_spend_enforcement" {
  project_id             = 11
  description           = "Production spending enforcement"
  threshold             = 80  # 80% of budget
  threshold_type        = "percent"
  timeframe             = "funding_source"
  enabled               = true
  overburn             = false
  notification_frequency = "weekly"
  spend_option         = "remaining"

  # Notify production team and finance
  user_ids = [20, 21]  # Production lead and finance manager
  user_group_ids = [30, 31]  # Production and finance teams
}

# Create service-specific enforcement for analytics project
resource "kion_project_enforcement" "analytics_service_enforcement" {
  project_id             = 12
  description           = "BigQuery usage enforcement"
  threshold             = 10000  # $10,000 threshold
  threshold_type        = "dollar"
  timeframe             = "month"
  enabled               = true
  service_id           = 50  # BigQuery service ID
  amount_type          = "last_month"
  cloud_rule_id        = 100  # Specific cloud rule for BigQuery

  # Notify analytics team
  user_ids = [25]  # Analytics lead
  user_group_ids = [35]  # Analytics team
}

# Create annual budget enforcement
resource "kion_project_enforcement" "annual_budget_enforcement" {
  project_id             = 11
  description           = "Annual budget enforcement"
  threshold             = 200000  # $200,000 threshold
  threshold_type        = "dollar"
  timeframe             = "year"
  enabled               = true
  overburn             = true
  notification_frequency = "weekly"

  # Notify multiple stakeholders
  user_ids = [20, 21, 22]  # Project leads and finance
  user_group_ids = [30, 31, 32]  # Project teams and management
}

# Output enforcement IDs
output "enforcement_ids" {
  value = {
    dev_budget     = kion_project_enforcement.dev_budget_enforcement.id
    prod_spend     = kion_project_enforcement.prod_spend_enforcement.id
    analytics      = kion_project_enforcement.analytics_service_enforcement.id
    annual_budget  = kion_project_enforcement.annual_budget_enforcement.id
  }
  description = "IDs of created enforcement rules"
}

# Output triggered status
output "enforcement_status" {
  value = {
    dev_budget     = kion_project_enforcement.dev_budget_enforcement.triggered
    prod_spend     = kion_project_enforcement.prod_spend_enforcement.triggered
    analytics      = kion_project_enforcement.analytics_service_enforcement.triggered
    annual_budget  = kion_project_enforcement.annual_budget_enforcement.triggered
  }
  description = "Current triggered status of enforcement rules"
}