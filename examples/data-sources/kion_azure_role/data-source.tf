# Example 1: Get all Azure roles
data "kion_azure_role" "all_roles" {
}

# Example 2: Filter roles by name pattern
data "kion_azure_role" "admin_roles" {
  filter {
    name   = "name"
    values = ["admin", "administrator"]
    regex  = true
  }
}

# Example 3: Filter by description content
data "kion_azure_role" "security_roles" {
  filter {
    name   = "description"
    values = ["security", "secure", "protection"]
    regex  = true
  }
}

# Example 4: Filter by management type
data "kion_azure_role" "system_roles" {
  filter {
    name   = "system_managed_policy"
    values = ["true"]
  }
}

# Example 5: Filter by Azure managed status
data "kion_azure_role" "azure_managed_roles" {
  filter {
    name   = "azure_managed_policy"
    values = ["true"]
  }
}

# Example 6: Filter by owner user group
data "kion_azure_role" "team_roles" {
  filter {
    name   = "owner_user_groups"
    values = ["3"]  # Roles owned by user group ID 3
  }
}

# Output examples
output "all_role_names" {
  description = "Names of all Azure roles"
  value       = [for role in data.kion_azure_role.all_roles.list : role.name]
}

output "admin_role_details" {
  description = "Details of administrator roles"
  value = {
    for role in data.kion_azure_role.admin_roles.list :
    role.name => {
      id                    = role.id
      description          = role.description
      azure_managed_policy = role.azure_managed_policy
      system_managed_policy = role.system_managed_policy
      role_permissions     = role.role_permissions
    }
  }
}

output "security_roles_summary" {
  description = "Summary of security-related roles"
  value = [
    for role in data.kion_azure_role.security_roles.list : {
      name        = role.name
      description = role.description
      permissions = role.role_permissions
    }
  ]
}

output "role_statistics" {
  description = "Statistics about Azure roles"
  value = {
    total_roles       = length(data.kion_azure_role.all_roles.list)
    admin_roles      = length(data.kion_azure_role.admin_roles.list)
    system_roles     = length(data.kion_azure_role.system_roles.list)
    azure_managed    = length(data.kion_azure_role.azure_managed_roles.list)
    team_roles      = length(data.kion_azure_role.team_roles.list)
  }
}

output "role_ownership_map" {
  description = "Map of roles to their owner groups"
  value = {
    for role in data.kion_azure_role.all_roles.list :
    role.name => {
      user_groups = [for group in role.owner_user_groups : group.id]
      users      = [for user in role.owner_users : user.id]
    }
  }
}