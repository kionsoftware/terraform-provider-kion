# Get all cloud rules
data "kion_cloud_rule" "all" {
}

# Filter cloud rules by name
data "kion_cloud_rule" "by_name" {
  filter {
    name   = "name"
    values = ["Security Baseline"]
  }
}

# Filter cloud rules by description with regex
data "kion_cloud_rule" "by_description" {
  filter {
    name   = "description"
    values = [".*security.*"]
    regex  = true
  }
}

# Filter by multiple criteria
data "kion_cloud_rule" "by_multiple" {
  filter {
    name   = "built_in"
    values = ["true"]
  }
  filter {
    name   = "concurrent_cft_sync"
    values = ["true"]
  }
}

# Example outputs
output "all_rules" {
  value = data.kion_cloud_rule.all.list
}

output "security_rule_id" {
  value = data.kion_cloud_rule.by_name.list[0].id
}

output "security_related_rules" {
  value = [for rule in data.kion_cloud_rule.by_description.list : rule.name]
}

output "built_in_concurrent_rules" {
  value = data.kion_cloud_rule.by_multiple.list
}