# Get all custom variables
data "kion_custom_variable" "all" {
}

# Filter custom variables by name
data "kion_custom_variable" "by_name" {
  filter {
    name   = "name"
    values = ["environment"]
  }
}

# Filter by type
data "kion_custom_variable" "string_vars" {
  filter {
    name   = "type"
    values = ["string"]
  }
}

# Filter by description with regex
data "kion_custom_variable" "tag_vars" {
  filter {
    name   = "description"
    values = [".*tags.*"]
    regex  = true
  }
}

# Filter by multiple criteria
data "kion_custom_variable" "owned_list_vars" {
  filter {
    name   = "type"
    values = ["list"]
  }
  filter {
    name   = "owner_user_ids"
    values = ["1"]
  }
}

# Example outputs
output "all_variables" {
  value = data.kion_custom_variable.all.custom_variables
}

output "environment_var" {
  value = data.kion_custom_variable.by_name.custom_variables[0]
}

output "string_var_names" {
  value = [for v in data.kion_custom_variable.string_vars.custom_variables : v.custom_variable_name]
}

output "tag_related_vars" {
  value = [
    for v in data.kion_custom_variable.tag_vars.custom_variables : {
      name = v.custom_variable_name
      type = v.custom_variable_type
    }
  ]
}

output "owned_list_var_details" {
  value = [
    for v in data.kion_custom_variable.owned_list_vars.custom_variables : {
      name              = v.custom_variable_name
      description       = v.custom_variable_description
      default_value_list = v.custom_variable_default_value_list
    }
  ]
}