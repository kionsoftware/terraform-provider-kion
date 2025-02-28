# Create administrator permission mapping
resource "kion_project_permission_mapping" "admin_permissions" {
  project_id     = 10
  app_role_id    = 1  # Administrator role

  # Assign to specific users (must be in numerical order)
  user_ids = [
    15,  # DevOps Lead
    16   # Senior Developer
  ]

  # Assign to admin groups (must be in numerical order)
  user_groups_ids = [
    25,  # DevOps Team
    26   # Platform Engineers
  ]
}

# Create read-only permission mapping
resource "kion_project_permission_mapping" "viewer_permissions" {
  project_id     = 10
  app_role_id    = 2  # Viewer role

  # Assign to specific users (must be in numerical order)
  user_ids = [
    20,  # QA Lead
    21,  # Security Analyst
    22   # Compliance Officer
  ]

  # Assign to viewer groups (must be in numerical order)
  user_groups_ids = [
    30,  # QA Team
    31,  # Security Team
    32   # Compliance Team
  ]
}

# Create developer permission mapping
resource "kion_project_permission_mapping" "developer_permissions" {
  project_id     = 10
  app_role_id    = 3  # Developer role

  # Assign to specific users (must be in numerical order)
  user_ids = [
    25,  # Frontend Developer
    26,  # Backend Developer
    27   # Full Stack Developer
  ]

  # Assign to development groups (must be in numerical order)
  user_groups_ids = [
    35,  # Frontend Team
    36,  # Backend Team
    37   # Mobile Team
  ]
}

# Output permission mapping IDs
output "permission_mapping_ids" {
  value = {
    admin     = kion_project_permission_mapping.admin_permissions.id
    viewer    = kion_project_permission_mapping.viewer_permissions.id
    developer = kion_project_permission_mapping.developer_permissions.id
  }
  description = "IDs of created permission mappings"
}