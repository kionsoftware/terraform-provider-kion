# Find all enabled users
data "kion_user" "active_users" {
  filter {
    enabled = true
  }
}

# Find specific user by username
data "kion_user" "devops_lead" {
  filter {
    username = "jsmith"
    enabled  = true
  }
}

# Find users by username pattern (multiple users)
data "kion_user" "engineering_team" {
  filter {
    username = "eng-"
    enabled  = true
  }
}

# Find disabled users
data "kion_user" "inactive_users" {
  filter {
    enabled = false
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

output "inactive_user_summary" {
  value = {
    count = length(data.kion_user.inactive_users.list)
    ids   = data.kion_user.inactive_users.list
  }
  description = "Summary of inactive users"
}