# Example 1: Get all CloudFormation templates
data "kion_aws_cloudformation_template" "all" {
}

# Example 2: Filter templates by name
data "kion_aws_cloudformation_template" "by_name" {
  filter {
    name   = "name"
    values = ["vpc-template"]
  }
}

# Example 3: Filter templates by region
data "kion_aws_cloudformation_template" "by_region" {
  filter {
    name   = "regions"
    values = ["us-east-1", "us-west-2"]
  }
}

# Example 4: Filter templates by multiple criteria
data "kion_aws_cloudformation_template" "multi_filter" {
  filter {
    name   = "name"
    values = ["security"]
    regex  = true
  }

  filter {
    name   = "owner_users.id"
    values = ["1"]
  }
}

# Example 5: Filter templates by tag
data "kion_aws_cloudformation_template" "by_tag" {
  filter {
    name   = "tags"
    values = ["production"]
    regex  = true
  }
}

# Example 6: Search for templates with specific parameters
data "kion_aws_cloudformation_template" "with_params" {
  filter {
    name   = "template_parameters"
    values = ["VpcCidr"]
    regex  = true
  }
}

# Output examples
output "all_templates" {
  description = "List of all CloudFormation templates"
  value = data.kion_aws_cloudformation_template.all.list
}

output "template_names" {
  description = "Names of all templates"
  value = [for template in data.kion_aws_cloudformation_template.all.list : template.name]
}

# Output template details in a map format
output "template_map" {
  description = "Map of templates with key details"
  value = {
    for template in data.kion_aws_cloudformation_template.all.list :
    template.id => {
      name        = template.name
      description = template.description
      regions     = template.regions
      tags        = template.tags
    }
  }
}

# Output filtered templates
output "security_templates" {
  description = "Templates matching security filter"
  value = data.kion_aws_cloudformation_template.multi_filter.list
}

# Output first matching template with parameters
output "first_template_with_params" {
  description = "First template with matching parameters"
  value = try(data.kion_aws_cloudformation_template.with_params.list[0], null)
}

# Count templates per region
output "templates_by_region" {
  description = "Count of templates per region"
  value = {
    for template in data.kion_aws_cloudformation_template.all.list :
    template.region => count(template.region)...
  }
}