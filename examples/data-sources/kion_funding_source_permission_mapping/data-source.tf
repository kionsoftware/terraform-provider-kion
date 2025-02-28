# Get all permission mappings
data "kion_funding_source_permission_mapping" "all" {
}

# Filter by funding source ID
data "kion_funding_source_permission_mapping" "by_funding_source" {
  filter {
    name   = "funding_source_id"
    values = ["1"]
  }
}

# Filter by application role
data "kion_funding_source_permission_mapping" "admin_roles" {
  filter {
    name   = "app_role_id"
    values = ["1"]  # Administrator role
  }
}

# Filter by user ID
data "kion_funding_source_permission_mapping" "user_mappings" {
  filter {
    name   = "user_ids"
    values = ["1"]  # Find mappings for specific user
  }
}

# Filter by user group
data "kion_funding_source_permission_mapping" "group_mappings" {
  filter {
    name   = "user_groups_ids"
    values = ["3"]  # Find mappings for specific group
  }
}

# Filter by multiple criteria
data "kion_funding_source_permission_mapping" "specific_mappings" {
  filter {
    name   = "funding_source_id"
    values = ["1"]
  }
  filter {
    name   = "app_role_id"
    values = ["2"]  # Viewer role
  }
}

# Example outputs
output "all_mappings" {
  value = data.kion_funding_source_permission_mapping.all.list
}

output "funding_source_mappings" {
  value = [
    for mapping in data.kion_funding_source_permission_mapping.by_funding_source.list : {
      app_role_id = mapping.app_role_id
      users       = mapping.user_ids
      groups      = mapping.user_groups_ids
    }
  ]
}

output "admin_role_mappings" {
  value = [
    for mapping in data.kion_funding_source_permission_mapping.admin_roles.list : {
      funding_source_id = mapping.funding_source_id
      users            = mapping.user_ids
      groups           = mapping.user_groups_ids
    }
  ]
}

output "user_access_details" {
  value = [
    for mapping in data.kion_funding_source_permission_mapping.user_mappings.list : {
      funding_source_id = mapping.funding_source_id
      app_role_id      = mapping.app_role_id
    }
  ]
}

output "group_access_summary" {
  value = [
    for mapping in data.kion_funding_source_permission_mapping.group_mappings.list : {
      funding_source_id = mapping.funding_source_id
      app_role_id      = mapping.app_role_id
      all_users        = mapping.user_ids
    }
  ]
}