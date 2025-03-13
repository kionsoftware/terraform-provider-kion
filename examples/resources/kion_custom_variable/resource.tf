# Create a string custom variable
resource "kion_custom_variable" "string_example" {
  name                    = "environment"
  description            = "Environment type for the deployment"
  type                   = "string"
  key_validation_regex   = "^[a-zA-Z0-9_-]+$"
  key_validation_message = "Key must contain only alphanumeric characters, underscores, and hyphens"
  value_validation_regex = "^(dev|staging|prod)$"
  value_validation_message = "Value must be one of: dev, staging, prod"
  default_value_string   = "dev"

  owner_user_ids = [1]
}

# Create a list custom variable
resource "kion_custom_variable" "list_example" {
  name                    = "allowed_regions"
  description            = "List of allowed AWS regions"
  type                   = "list"
  key_validation_regex   = "^[a-zA-Z0-9_-]+$"
  key_validation_message = "Key must contain only alphanumeric characters, underscores, and hyphens"
  value_validation_regex = "^[a-z]{2}-[a-z]+-[0-9]$"
  value_validation_message = "Value must be a valid AWS region format (e.g., us-east-1)"

  default_value_list = [
    "us-east-1",
    "us-west-2",
    "eu-west-1"
  ]

  owner_user_group_ids = [1, 2]
}

# Create a map custom variable
resource "kion_custom_variable" "map_example" {
  name                    = "resource_tags"
  description            = "Default resource tags for cloud resources"
  type                   = "map"
  key_validation_regex   = "^[a-zA-Z0-9_-]+$"
  key_validation_message = "Key must contain only alphanumeric characters, underscores, and hyphens"
  value_validation_regex = "^[a-zA-Z0-9_-]+$"
  value_validation_message = "Value must contain only alphanumeric characters, underscores, and hyphens"

  default_value_map = {
    "Environment" = "production"
    "Owner"       = "platform-team"
    "CostCenter"  = "12345"
  }

  owner_user_ids = [1]
  owner_user_group_ids = [2]
}

# Output examples
output "string_var_id" {
  value = kion_custom_variable.string_example.id
}

output "list_var_id" {
  value = kion_custom_variable.list_example.id
}

output "map_var_id" {
  value = kion_custom_variable.map_example.id
}