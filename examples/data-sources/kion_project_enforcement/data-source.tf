# Find all enabled enforcements
data "kion_project_enforcement" "enabled_enforcements" {
  filter {
    name   = "enabled"
    values = ["true"]
  }
}

# Find enforcements for specific project
data "kion_project_enforcement" "project_enforcements" {
  filter {
    name   = "project_id"
    values = ["10"]  # Development project
  }
}

# Find triggered enforcements
data "kion_project_enforcement" "triggered_enforcements" {
  filter {
    name   = "triggered"
    values = ["true"]
  }
}

# Find enforcements by threshold type
data "kion_project_enforcement" "percentage_enforcements" {
  filter {
    name   = "threshold_type"
    values = ["percent"]
  }
}

# Find enforcements by description
data "kion_project_enforcement" "budget_enforcements" {
  filter {
    name   = "description"
    values = [".*budget.*", ".*spend.*"]
    regex  = true
  }
}

# Output enforcement information
output "enabled_enforcements" {
  value = {
    for enforcement in data.kion_project_enforcement.enabled_enforcements.enforcements :
    enforcement.id => {
      project_id = enforcement.project_id
      threshold  = enforcement.threshold
      timeframe  = enforcement.timeframe
      triggered  = enforcement.triggered
    }
  }
  description = "List of enabled enforcements"
}

output "project_enforcement_details" {
  value = {
    for enforcement in data.kion_project_enforcement.project_enforcements.enforcements :
    enforcement.id => {
      description = enforcement.description
      threshold   = enforcement.threshold
      enabled     = enforcement.enabled
      triggered   = enforcement.triggered
    }
  }
  description = "Enforcements for specific project"
}

output "triggered_enforcement_summary" {
  value = {
    count = length(data.kion_project_enforcement.triggered_enforcements.enforcements)
    details = {
      for enforcement in data.kion_project_enforcement.triggered_enforcements.enforcements :
      enforcement.id => {
        project_id  = enforcement.project_id
        description = enforcement.description
      }
    }
  }
  description = "Summary of triggered enforcements"
}

output "percentage_threshold_enforcements" {
  value = {
    for enforcement in data.kion_project_enforcement.percentage_enforcements.enforcements :
    enforcement.id => {
      project_id = enforcement.project_id
      threshold  = enforcement.threshold
      timeframe  = enforcement.timeframe
    }
  }
  description = "Enforcements using percentage thresholds"
}

output "budget_enforcement_notifications" {
  value = {
    for enforcement in data.kion_project_enforcement.budget_enforcements.enforcements :
    enforcement.id => {
      user_ids       = enforcement.user_ids
      user_group_ids = enforcement.user_group_ids
      frequency      = enforcement.notification_frequency
    }
  }
  description = "Notification settings for budget-related enforcements"
}