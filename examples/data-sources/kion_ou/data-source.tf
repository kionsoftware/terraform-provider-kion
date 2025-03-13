# Example 1: List all OUs
data "kion_ou" "all" {
}

# Example 2: Filter OUs by name
data "kion_ou" "development" {
  filter {
    name   = "name"
    values = ["Development"]
  }
}

# Example 3: Filter OUs using regex pattern
data "kion_ou" "prod_related" {
  filter {
    name   = "name"
    values = [".*prod.*"]
    regex  = true
  }
}

# Example 4: Filter by parent OU ID
data "kion_ou" "sub_ous" {
  filter {
    name   = "parent_ou_id"
    values = ["123"]  # Replace with actual parent OU ID
  }
}

# Example outputs
output "all_ous" {
  description = "List of all organizational units"
  value       = data.kion_ou.all.list
}

output "dev_ou_details" {
  description = "Details of development OUs"
  value       = data.kion_ou.development.list
}

output "prod_ou_names" {
  description = "Names of production-related OUs"
  value       = [for ou in data.kion_ou.prod_related.list : ou.name]
}

output "ou_hierarchy" {
  description = "Map of OUs with their parent IDs"
  value = {
    for ou in data.kion_ou.all.list :
    ou.name => {
      id           = ou.id
      parent_ou_id = ou.parent_ou_id
      description  = ou.description
    }
  }
}
