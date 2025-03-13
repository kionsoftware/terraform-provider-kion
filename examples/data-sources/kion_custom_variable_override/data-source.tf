# Get all custom variable overrides
data "kion_custom_variable_override" "all" {
}

# Filter overrides by custom variable ID
data "kion_custom_variable_override" "by_variable" {
  filter {
    name   = "custom_variable_id"
    values = ["123"]  # Custom Variable ID
  }
}

# Filter by entity type
data "kion_custom_variable_override" "project_overrides" {
  filter {
    name   = "entity_type"
    values = ["project"]
  }
}

# Filter by entity ID
data "kion_custom_variable_override" "specific_project" {
  filter {
    name   = "entity_id"
    values = ["321"]  # Project ID
  }
}

# Filter by multiple criteria
data "kion_custom_variable_override" "specific_overrides" {
  filter {
    name   = "entity_type"
    values = ["ou"]
  }
  filter {
    name   = "custom_variable_id"
    values = ["456"]  # Custom Variable ID
  }
}

# Example outputs
output "all_overrides" {
  value = data.kion_custom_variable_override.all.list
}

output "variable_overrides" {
  value = data.kion_custom_variable_override.by_variable.list
}

output "project_override_values" {
  value = [
    for override in data.kion_custom_variable_override.project_overrides.list : {
      entity_id    = override.entity_id
      value_string = override.value_string
      value_list   = override.value_list
      value_map    = override.value_map
    }
  ]
}

output "specific_project_override" {
  value = data.kion_custom_variable_override.specific_project.list[0]
}

output "ou_variable_overrides" {
  value = [
    for override in data.kion_custom_variable_override.specific_overrides.list : {
      entity_id = override.entity_id
      values    = coalesce(override.value_string,
                          length(override.value_list) > 0 ? jsonencode(override.value_list) : null,
                          length(override.value_map) > 0 ? jsonencode(override.value_map) : null)
    }
  ]
}