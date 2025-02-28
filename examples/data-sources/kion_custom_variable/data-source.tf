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
  value = data.kion_custom_variable.all.list
}

output "environment_var" {
  value = data.kion_custom_variable.by_name.list[0]
}

output "string_var_names" {
  value = [for var in data.kion_custom_variable.string_vars.list : var.name]
}

output "tag_related_vars" {
  value = [
    for var in data.kion_custom_variable.tag_vars.list : {
      name = var.name
      type = var.type
    }
  ]
}

output "owned_list_var_details" {
  value = [
    for var in data.kion_custom_variable.owned_list_vars.list : {
      name        = var.name
      description = var.description
      default_value_list = var.default_value_list
    }
  ]
}