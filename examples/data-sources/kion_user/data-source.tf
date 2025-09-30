# Find all enabled users (legacy approach)
data "kion_user" "active_users" {
  filter {
    enabled = true
  }
}

# Find specific user by username (legacy approach)
data "kion_user" "devops_lead" {
  filter {
    username = "jsmith"
    enabled  = true
  }
}

# Find users by username pattern using new generic filtering with regex
data "kion_user" "engineering_team" {
  filter {
    name   = "username"
    values = ["eng-.*", "dev-.*"]
    regex  = true
  }
}

# Find disabled users using new generic filtering
data "kion_user" "inactive_users" {
  filter {
    name   = "enabled"
    values = ["false"]
  }
}

# Find users by specific IDs using new generic filtering
data "kion_user" "specific_users" {
  filter {
    name   = "id"
    values = ["123", "456", "789"]
  }
}

# Find users with multiple username patterns
data "kion_user" "admin_users" {
  filter {
    name   = "username"
    values = ["admin.*", "root.*", "super.*"]
    regex  = true
  }
}

# Output user information
output "active_user_count" {
  value       = length(data.kion_user.active_users.list)
  description = "Number of active users in the system"
}

output "devops_lead_id" {
  value       = length(data.kion_user.devops_lead.list) > 0 ? data.kion_user.devops_lead.list[0] : null
  description = "User ID of the DevOps lead"
}

output "engineering_team_ids" {
  value       = data.kion_user.engineering_team.list
  description = "List of engineering team user IDs"
}

output "specific_users_count" {
  value       = length(data.kion_user.specific_users.list)
  description = "Count of specific users found by ID"
}

output "admin_users_summary" {
  value = {
    count = length(data.kion_user.admin_users.list)
    ids   = data.kion_user.admin_users.list
  }
  description = "Summary of admin users found by regex pattern"
}

output "inactive_user_summary" {
  value = {
    count = length(data.kion_user.inactive_users.list)
    ids   = data.kion_user.inactive_users.list
  }
  description = "Summary of inactive users"
}