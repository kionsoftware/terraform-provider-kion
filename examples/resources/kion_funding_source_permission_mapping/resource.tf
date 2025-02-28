# Create a permission mapping for administrators
resource "kion_funding_source_permission_mapping" "admin_mapping" {
  funding_source_id = kion_funding_source.complete_example.id
  app_role_id      = 1  # Administrator role

  # User IDs with admin access
  user_ids = [
    1,  # Platform Admin
    2   # Finance Admin
  ]

  # User group IDs with admin access
  user_groups_ids = [
    1,  # Cloud Admin Team
    2   # Finance Team
  ]
}

# Create a permission mapping for viewers
resource "kion_funding_source_permission_mapping" "viewer_mapping" {
  funding_source_id = kion_funding_source.complete_example.id
  app_role_id      = 2  # Viewer role

  # Individual users with view access
  user_ids = [
    3,  # Project Manager
    4,  # Cost Analyst
    5   # Auditor
  ]

  # Groups with view access
  user_groups_ids = [
    3,  # Project Management Team
    4   # Audit Team
  ]
}

# Create a permission mapping for project-specific access
resource "kion_funding_source_permission_mapping" "project_mapping" {
  funding_source_id = kion_funding_source.project_budget.id
  app_role_id      = 3  # Project Manager role

  # Project team members
  user_ids = [
    6,  # Project Lead
    7   # Technical Lead
  ]

  # Project teams
  user_groups_ids = [
    5,  # Development Team
    6   # QA Team
  ]
}

# Output examples
output "admin_mapping_id" {
  value = kion_funding_source_permission_mapping.admin_mapping.id
}

output "viewer_mapping_id" {
  value = kion_funding_source_permission_mapping.viewer_mapping.id
}

output "project_mapping_id" {
  value = kion_funding_source_permission_mapping.project_mapping.id
}