# Example 1: Get all ARM templates
data "kion_azure_arm_template" "all_templates" {
}

# Example 2: Filter templates by name
data "kion_azure_arm_template" "storage_templates" {
  filter {
    name   = "name"
    values = ["storage", "Storage"]
    regex  = true
  }
}

# Example 3: Filter by multiple criteria
data "kion_azure_arm_template" "filtered_templates" {
  filter {
    name   = "resource_group_name"
    values = ["storage-rg"]
  }
  filter {
    name   = "deployment_mode"
    values = ["1"]  # Incremental deployment
  }
}

# Example 4: Filter by description content
data "kion_azure_arm_template" "webapp_templates" {
  filter {
    name   = "description"
    values = ["web", "app", "application"]
    regex  = true
  }
}

# Example 5: Filter by version
data "kion_azure_arm_template" "latest_templates" {
  filter {
    name   = "version"
    values = ["2"]  # Version 2 templates
  }
}

# Output examples
output "all_template_names" {
  description = "Names of all ARM templates"
  value       = [for template in data.kion_azure_arm_template.all_templates.list : template.name]
}

output "storage_template_details" {
  description = "Details of storage-related templates"
  value = {
    for template in data.kion_azure_arm_template.storage_templates.list :
    template.name => {
      id                      = template.id
      description            = template.description
      resource_group_name    = template.resource_group_name
      resource_group_region_id = template.resource_group_region_id
      deployment_mode        = template.deployment_mode
      version               = template.version
    }
  }
}

output "webapp_template_count" {
  description = "Number of web application templates"
  value       = length(data.kion_azure_arm_template.webapp_templates.list)
}

output "template_statistics" {
  description = "Statistics about ARM templates"
  value = {
    total_templates = length(data.kion_azure_arm_template.all_templates.list)
    storage_templates = length(data.kion_azure_arm_template.storage_templates.list)
    latest_templates = length(data.kion_azure_arm_template.latest_templates.list)
  }
}