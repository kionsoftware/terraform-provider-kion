# Create a project with multiple owners, labels, and budget configurations
resource "kion_project" "development_project" {
  name                 = "Development Infrastructure"
  description          = "Infrastructure resources for the development team"
  ou_id                = 3
  permission_scheme_id = 2
  default_aws_region   = "us-east-1"
  auto_pay            = true

  # Project owners - both users and groups
  owner_user_ids {
    id = 10  # Lead Developer
  }
  owner_user_ids {
    id = 11  # DevOps Engineer
  }

  owner_user_group_ids {
    id = 5  # Development Team
  }

  # Labels for organization and tracking
  labels = {
    "Environment" = "Development"
    "CostCenter" = "IT-1234"
    "Team"       = "Platform"
    "Owner"      = "DevOps"
  }

  # Budget configuration with monthly allocations
  budget {
    start_datecode = "2024-01"
    end_datecode   = "2024-12"
    amount         = 120000  # $120,000 total budget

    # Specific monthly allocations with different funding sources
    data {
      datecode          = "2024-01"
      amount            = 12000
      funding_source_id = 1
      priority          = 1
    }
    data {
      datecode          = "2024-02"
      amount            = 8000
      funding_source_id = 2
      priority          = 1
    }

    # Specify funding sources for remaining months
    funding_source_ids = [1, 2]
  }

  # Additional project funding
  project_funding {
    amount            = 50000
    funding_source_id = 1
    funding_order     = 1
    start_datecode    = "2024-01"
    end_datecode      = "2024-06"
  }

  project_funding {
    amount            = 70000
    funding_source_id = 2
    funding_order     = 2
    start_datecode    = "2024-07"
    end_datecode      = "2024-12"
  }
}

# Create a production project with different settings
resource "kion_project" "production_project" {
  name                 = "Production Infrastructure"
  description          = "Production environment infrastructure and services"
  ou_id                = 3
  permission_scheme_id = 3
  default_aws_region   = "us-west-2"
  auto_pay            = false

  # Production project owners
  owner_user_ids {
    id = 12  # Production Lead
  }

  owner_user_group_ids {
    id = 6  # Operations Team
  }

  # Production-specific labels
  labels = {
    "Environment" = "Production"
    "CostCenter" = "IT-5678"
    "Team"       = "Operations"
    "Critical"   = "Yes"
  }

  # Annual budget
  budget {
    start_datecode = "2024-01"
    end_datecode   = "2025-01"
    amount         = 240000  # $240,000 annual budget

    # Distribute across two funding sources
    funding_source_ids = [3, 4]
  }
}

# Output project information
output "development_project_details" {
  value = {
    id          = kion_project.development_project.id
    name        = kion_project.development_project.name
    archived    = kion_project.development_project.archived
  }
  description = "Development project details"
}

output "production_project_details" {
  value = {
    id          = kion_project.production_project.id
    name        = kion_project.production_project.name
    archived    = kion_project.production_project.archived
  }
  description = "Production project details"
}
