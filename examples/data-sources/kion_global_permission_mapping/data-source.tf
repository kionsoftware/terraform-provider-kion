# Example 1: Get all global permission mappings
data "kion_global_permission_mapping" "all" {
}

# Example outputs
output "all_mappings" {
  description = "List of all global permission mappings"
  value       = data.kion_global_permission_mapping.all.list
}

output "app_role_ids" {
  description = "List of all app role IDs in use"
  value       = [for mapping in data.kion_global_permission_mapping.all.list : mapping.app_role_id]
}

output "user_groups_per_role" {
  description = "Map of app role IDs to their associated user group IDs"
  value = {
    for mapping in data.kion_global_permission_mapping.all.list :
    mapping.app_role_id => mapping.user_groups_ids
  }
}

output "users_per_role" {
  description = "Map of app role IDs to their associated user IDs"
  value = {
    for mapping in data.kion_global_permission_mapping.all.list :
    mapping.app_role_id => mapping.user_ids
  }
}
