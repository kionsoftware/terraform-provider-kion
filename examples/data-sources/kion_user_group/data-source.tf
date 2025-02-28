# Find all enabled user groups
data "kion_user_group" "enabled_groups" {
  filter {
    name   = "enabled"
    values = ["true"]
  }
}

# Find groups by name pattern
data "kion_user_group" "team_groups" {
  filter {
    name   = "name"
    values = [".*Team.*"]
    regex  = true
  }
}

# Find groups by specific IDMS
data "kion_user_group" "idms_groups" {
  filter {
    name   = "idms_id"
    values = ["1"]  # Internal IDMS
  }
}

# Find groups by description content
data "kion_user_group" "engineering_groups" {
  filter {
    name   = "description"
    values = [".*engineer.*", ".*development.*"]
    regex  = true
  }
}

# Find groups by creation date
data "kion_user_group" "recent_groups" {
  filter {
    name   = "created_at"
    values = ["2024-01.*"]  # Groups created in January 2024
    regex  = true
  }
}

# Output group information
output "enabled_groups" {
  value = {
    for group in data.kion_user_group.enabled_groups.list :
    group.name => {
      id          = group.id
      description = group.description
      idms_id     = group.idms_id
    }
  }
  description = "List of all enabled groups"
}

output "team_group_names" {
  value = [
    for group in data.kion_user_group.team_groups.list :
    group.name
  ]
  description = "Names of groups containing 'Team'"
}

output "idms_group_details" {
  value = {
    for group in data.kion_user_group.idms_groups.list :
    group.name => {
      id          = group.id
      description = group.description
      enabled     = group.enabled
    }
  }
  description = "Details of groups in specific IDMS"
}

output "engineering_group_summary" {
  value = {
    count = length(data.kion_user_group.engineering_groups.list)
    groups = {
      for group in data.kion_user_group.engineering_groups.list :
      group.name => group.id
    }
  }
  description = "Summary of engineering-related groups"
}

output "recent_group_details" {
  value = {
    for group in data.kion_user_group.recent_groups.list :
    group.name => {
      id         = group.id
      created_at = group.created_at
    }
  }
  description = "Details of recently created groups"
}