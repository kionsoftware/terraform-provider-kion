# Create labels for different organizational purposes
# Environment labels
resource "kion_label" "env_production" {
  key   = "Environment"
  value = "Production"
  color = "#FF0000"  # Red for production
}

resource "kion_label" "env_staging" {
  key   = "Environment"
  value = "Staging"
  color = "#FFA500"  # Orange for staging
}

resource "kion_label" "env_development" {
  key   = "Environment"
  value = "Development"
  color = "#00FF00"  # Green for development
}

# Cost center labels
resource "kion_label" "cost_center_it" {
  key   = "CostCenter"
  value = "IT-1234"
  color = "#0000FF"  # Blue for IT department
}

resource "kion_label" "cost_center_marketing" {
  key   = "CostCenter"
  value = "MKT-5678"
  color = "#800080"  # Purple for marketing department
}

# Team labels
resource "kion_label" "team_platform" {
  key   = "Team"
  value = "Platform"
  color = "#FFD700"  # Gold for platform team
}

resource "kion_label" "team_security" {
  key   = "Team"
  value = "Security"
  color = "#4B0082"  # Indigo for security team
}

# Output label IDs for reference
output "environment_label_ids" {
  value = {
    production  = kion_label.env_production.id
    staging     = kion_label.env_staging.id
    development = kion_label.env_development.id
  }
  description = "IDs of environment labels"
}

output "cost_center_label_ids" {
  value = {
    it        = kion_label.cost_center_it.id
    marketing = kion_label.cost_center_marketing.id
  }
  description = "IDs of cost center labels"
}

output "team_label_ids" {
  value = {
    platform = kion_label.team_platform.id
    security = kion_label.team_security.id
  }
  description = "IDs of team labels"
}
