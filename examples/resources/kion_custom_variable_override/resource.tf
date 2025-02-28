# Create a string variable override for a project
resource "kion_custom_variable_override" "project_string_override" {
  custom_variable_id = kion_custom_variable.string_example.id
  entity_type       = "project"
  entity_id         = "123"  # Project ID
  value_string      = "prod"
}

# Create a list variable override for an OU
resource "kion_custom_variable_override" "ou_list_override" {
  custom_variable_id = kion_custom_variable.list_example.id
  entity_type       = "ou"
  entity_id         = "456"  # OU ID

  value_list = [
    "us-east-1",
    "us-east-2"
  ]
}

# Create a map variable override for a cloud rule
resource "kion_custom_variable_override" "rule_map_override" {
  custom_variable_id = kion_custom_variable.map_example.id
  entity_type       = "cloud_rule"
  entity_id         = "789"  # Cloud Rule ID

  value_map = {
    "Environment" = "staging"
    "Owner"       = "dev-team"
    "Project"     = "terraform-demo"
  }
}

# Example of multiple overrides for different entities
resource "kion_custom_variable_override" "multi_env_override" {
  custom_variable_id = kion_custom_variable.string_example.id
  entity_type       = "project"
  entity_id         = "321"  # Different Project ID
  value_string      = "staging"
}

# Output examples
output "project_override_id" {
  value = kion_custom_variable_override.project_string_override.id
}

output "ou_override_id" {
  value = kion_custom_variable_override.ou_list_override.id
}

output "rule_override_id" {
  value = kion_custom_variable_override.rule_map_override.id
}