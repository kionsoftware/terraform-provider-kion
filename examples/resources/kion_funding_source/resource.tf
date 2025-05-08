# Create a complete funding source with all available options
resource "kion_funding_source" "complete_example" {
  name                 = "Annual Cloud Budget 2024"
  description         = "Main cloud infrastructure budget for fiscal year 2024"
  amount              = 1000000
  start_datecode      = "2024-01"
  end_datecode        = "2024-12"
  ou_id               = 1
  permission_scheme_id = 4

  # Ownership configuration
  owner_users {
    id = 1  # Platform Team Lead
  }
  owner_users {
    id = 2  # Finance Manager
  }

  owner_user_groups {
    id = 1  # Cloud Platform Team
  }
  owner_user_groups {
    id = 2  # Finance Team
  }

  # Labels for categorization
  labels = {
    environment = "production"
    fiscal_year = "2024"
    department  = "it"
    cost_center = "12345"
  }
}

# Create a project-specific funding source
resource "kion_funding_source" "project_budget" {
  name                 = "Project Alpha Budget"
  description         = "Dedicated budget for Project Alpha development"
  amount              = 50000
  start_datecode      = "2024-01"
  end_datecode        = "2024-06"
  ou_id               = 2
  permission_scheme_id = 4

  owner_users {
    id = 3  # Project Manager
  }

  labels = {
    project     = "alpha"
    environment = "development"
  }
}

# Create a quarterly funding source
resource "kion_funding_source" "quarterly_budget" {
  name                 = "Q1 2024 Operations Budget"
  description         = "Quarterly budget for cloud operations"
  amount              = 250000
  start_datecode      = "2024-01"
  end_datecode        = "2024-03"
  ou_id               = 3
  permission_scheme_id = 4

  owner_user_groups {
    id = 3  # Operations Team
  }

  labels = {
    quarter     = "Q1"
    fiscal_year = "2024"
    type        = "operational"
  }
}

# Output examples
output "annual_budget_id" {
  value = kion_funding_source.complete_example.id
}

output "project_budget_id" {
  value = kion_funding_source.project_budget.id
}

output "quarterly_budget_id" {
  value = kion_funding_source.quarterly_budget.id
}
