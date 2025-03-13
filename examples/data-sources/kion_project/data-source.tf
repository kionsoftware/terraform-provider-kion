# Find all development projects
data "kion_project" "dev_projects" {
  filter {
    name   = "name"
    values = ["dev-", "development-"]
    regex  = true
  }
}

# Find projects in a specific OU
data "kion_project" "platform_projects" {
  filter {
    name   = "ou_id"
    values = ["3"]  # Platform OU
  }
}

# Find projects by description
data "kion_project" "analytics_projects" {
  filter {
    name   = "description"
    values = [".*analytics.*", ".*data.*"]
    regex  = true
  }
}

# Find projects by AWS region
data "kion_project" "east_projects" {
  filter {
    name   = "default_aws_region"
    values = ["us-east-1"]
  }
}

# Find archived projects
data "kion_project" "archived_projects" {
  filter {
    name   = "archived"
    values = ["true"]
  }
}

# Output project information
output "development_projects" {
  value = {
    for project in data.kion_project.dev_projects.list :
    project.name => {
      id          = project.id
      description = project.description
      ou_id       = project.ou_id
      region      = project.default_aws_region
    }
  }
  description = "List of development projects"
}

output "platform_project_names" {
  value = [
    for project in data.kion_project.platform_projects.list :
    project.name
  ]
  description = "Names of projects in the platform OU"
}

output "analytics_project_details" {
  value = {
    for project in data.kion_project.analytics_projects.list :
    project.name => {
      id          = project.id
      description = project.description
      auto_pay    = project.auto_pay
    }
  }
  description = "Details of analytics-related projects"
}

output "east_region_projects" {
  value = {
    for project in data.kion_project.east_projects.list :
    project.name => project.id
  }
  description = "Projects in us-east-1 region"
}

output "archived_project_count" {
  value = length(data.kion_project.archived_projects.list)
  description = "Number of archived projects"
}