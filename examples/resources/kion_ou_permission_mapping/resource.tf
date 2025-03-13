# Create permission mappings for different roles and users
resource "kion_ou_permission_mapping" "admin_permissions" {
  ou_id       = 3
  app_role_id = 1  # Administrator role

  # Assign to specific users (must be in numerical order)
  user_ids = [10, 11, 12]  # DevOps and Security leads

  # Assign to specific groups (must be in numerical order)
  user_groups_ids = [5, 6]  # DevOps and Security teams
}

resource "kion_ou_permission_mapping" "viewer_permissions" {
  ou_id       = 3
  app_role_id = 2  # Viewer role

  # Assign to specific users (must be in numerical order)
  user_ids = [20, 21, 22]  # Auditors and compliance team

  # Assign to specific groups (must be in numerical order)
  user_groups_ids = [7, 8]  # Audit and compliance groups
}

# Output permission mapping IDs
output "admin_permission_mapping_id" {
  value       = kion_ou_permission_mapping.admin_permissions.id
  description = "The ID of the admin permission mapping"
}

output "viewer_permission_mapping_id" {
  value       = kion_ou_permission_mapping.viewer_permissions.id
  description = "The ID of the viewer permission mapping"
}