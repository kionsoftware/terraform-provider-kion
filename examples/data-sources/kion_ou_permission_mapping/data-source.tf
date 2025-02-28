# Fetch existing permission mappings for an OU
data "kion_ou_permission_mapping" "existing_permissions" {
  ou_id = 3
}

# Output all permission mappings
output "all_permission_mappings" {
  value       = data.kion_ou_permission_mapping.existing_permissions.list
  description = "List of all permission mappings for the OU"
}

# Output specific information about the permission mappings
output "app_role_ids" {
  value       = [for mapping in data.kion_ou_permission_mapping.existing_permissions.list : mapping.app_role_id]
  description = "List of all app role IDs in use"
}

output "user_assignments" {
  value = {
    for mapping in data.kion_ou_permission_mapping.existing_permissions.list :
    mapping.app_role_id => mapping.user_ids
  }
  description = "Map of app role IDs to assigned user IDs"
}

output "group_assignments" {
  value = {
    for mapping in data.kion_ou_permission_mapping.existing_permissions.list :
    mapping.app_role_id => mapping.user_groups_ids
  }
  description = "Map of app role IDs to assigned user group IDs"
}