# Create a webhook for funding source events
resource "kion_webhook" "funding_alerts" {
  name        = "Funding Source Alerts"
  description = "Notifies Slack channel about funding source changes"
  url         = "https://hooks.slack.com/services/XXXXX/YYYYY/ZZZZZ"

  # Event types to monitor
  event_types = [
    "funding-source.created",
    "funding-source.updated",
    "funding-source.deleted"
  ]

  # Headers for Slack webhook
  headers = {
    "Content-Type" = "application/json"
  }

  # Owner configuration
  owner_users {
    id = 30  # Finance Team Lead
  }
  owner_user_groups {
    id = 40  # Finance Team
  }
}

# Create a webhook for security events
resource "kion_webhook" "security_alerts" {
  name        = "Security Event Monitor"
  description = "Sends security-related events to Security Information and Event Management (SIEM) system"
  url         = "https://siem.example.com/api/events"

  event_types = [
    "compliance.check.failed",
    "compliance.check.passed",
    "aws-cloudtrail.security.event",
    "user.login.failed"
  ]

  # Custom headers for SIEM API
  headers = {
    "Authorization" = "Bearer ${var.siem_api_token}"
    "Content-Type"  = "application/json"
    "X-Source"      = "Kion"
  }

  owner_users {
    id = 31  # Security Operations Lead
  }
  owner_user_groups {
    id = 41  # SecOps Team
  }
}

# Create a webhook for project lifecycle events
resource "kion_webhook" "project_lifecycle" {
  name        = "Project Lifecycle Monitor"
  description = "Monitors project creation, updates, and deletion events"
  url         = "https://automation.example.com/webhooks/projects"

  event_types = [
    "project.created",
    "project.updated",
    "project.deleted",
    "project.archived",
    "project.owner.added",
    "project.owner.removed"
  ]

  # Headers for internal automation system
  headers = {
    "X-API-Key"    = var.automation_api_key
    "Content-Type" = "application/json"
  }

  owner_users {
    id = 32  # DevOps Lead
  }
  owner_user_groups {
    id = 42  # DevOps Team
  }
}

# Output webhook information
output "webhook_configs" {
  value = {
    funding_alerts = {
      id          = kion_webhook.funding_alerts.id
      event_count = length(kion_webhook.funding_alerts.event_types)
    }
    security_alerts = {
      id          = kion_webhook.security_alerts.id
      event_count = length(kion_webhook.security_alerts.event_types)
    }
    project_lifecycle = {
      id          = kion_webhook.project_lifecycle.id
      event_count = length(kion_webhook.project_lifecycle.event_types)
    }
  }
  description = "Details of created webhooks including event type counts"
}