# Create a project with multiple owners, labels, and budget configurations
resource "kion_project" "development_project" {
  name                 = "Development Infrastructure"
  description          = "Infrastructure resources for the development team"
  ou_id                = 3
  permission_scheme_id = 2
  default_aws_region   = "us-east-1"
  auto_pay             = true

  # Project owners - both users and groups
  owner_user_ids {
    id = 10 # Lead Developer
  }
  owner_user_ids {
    id = 11 # DevOps Engineer
  }

  owner_user_group_ids {
    id = 5 # Development Team
  }

  # Labels for organization and tracking
  labels = {
    "Environment" = "Development"
    "CostCenter"  = "IT-1234"
    "Team"        = "Platform"
    "Owner"       = "DevOps"
  }

  # Budget configuration with monthly allocations
  budget {
    start_datecode = "2024-01"
    end_datecode   = "2024-12"
    amount         = 120000 # $120,000 total budget

    # Monthly allocations - distributing budget across the year
    # Alternating between two funding sources
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
    data {
      datecode          = "2024-03"
      amount            = 10000
      funding_source_id = 1
      priority          = 1
    }
    data {
      datecode          = "2024-04"
      amount            = 10000
      funding_source_id = 2
      priority          = 1
    }
    data {
      datecode          = "2024-05"
      amount            = 10000
      funding_source_id = 1
      priority          = 1
    }
    data {
      datecode          = "2024-06"
      amount            = 10000
      funding_source_id = 2
      priority          = 1
    }
    data {
      datecode          = "2024-07"
      amount            = 10000
      funding_source_id = 1
      priority          = 1
    }
    data {
      datecode          = "2024-08"
      amount            = 10000
      funding_source_id = 2
      priority          = 1
    }
    data {
      datecode          = "2024-09"
      amount            = 10000
      funding_source_id = 1
      priority          = 1
    }
    data {
      datecode          = "2024-10"
      amount            = 10000
      funding_source_id = 2
      priority          = 1
    }
    data {
      datecode          = "2024-11"
      amount            = 10000
      funding_source_id = 1
      priority          = 1
    }
    data {
      datecode          = "2024-12"
      amount            = 10000
      funding_source_id = 2
      priority          = 1
    }

    # Specify both funding sources used in monthly allocations
    funding_source_ids = [1, 2]
  }
}

# Create a production project with different settings
resource "kion_project" "production_project" {
  name                 = "Production Infrastructure"
  description          = "Production environment infrastructure and services"
  ou_id                = 3
  permission_scheme_id = 3
  default_aws_region   = "us-west-2"
  auto_pay             = false

  # Production project owners
  owner_user_ids {
    id = 12 # Production Lead
  }

  owner_user_group_ids {
    id = 6 # Operations Team
  }

  # Production-specific labels
  labels = {
    "Environment" = "Production"
    "CostCenter"  = "IT-5678"
    "Team"        = "Operations"
    "Critical"    = "Yes"
  }

  # Annual budget
  budget {
    start_datecode = "2024-01"
    end_datecode   = "2025-01"
    amount         = 240000 # $240,000 annual budget

    # Distribute across two funding sources
    funding_source_ids = [3, 4]
  }
}

# Output project information
output "development_project_details" {
  value = {
    id       = kion_project.development_project.id
    name     = kion_project.development_project.name
    archived = kion_project.development_project.archived
  }
  description = "Development project details"
}

output "production_project_details" {
  value = {
    id       = kion_project.production_project.id
    name     = kion_project.production_project.name
    archived = kion_project.production_project.archived
  }
  description = "Production project details"
}

# =============================================================================
# Moving a Project Between OUs
# =============================================================================
# Projects can be moved between OUs without being destroyed and recreated.
# When you change the `ou_id`, the provider uses the Kion move API to relocate
# the project while preserving its attached accounts.

# Import an existing project to manage it with Terraform
import {
  to = kion_project.existing_project
  id = "4"
}

# Project that can be moved between OUs
# To move: simply change the ou_id value and apply
resource "kion_project" "existing_project" {
  name                 = "My Project"
  ou_id                = 11 # Change this to move the project to a different OU
  permission_scheme_id = 1

  owner_user_group_ids {
    id = 1
  }

  labels = {
    "Department" = "Engineering"
  }

  # Optional: Settings that control how the move is performed
  # If not specified, defaults are: cloud_rule_setting="convert", financial_setting="move"
  move_ou_settings {
    # "convert" = inherited cloud rules from old OU become local rules on the project
    # "remove" = cloud rules are removed from the project
    cloud_rule_setting = "convert"

    # "move" = financial history transfers to the new OU (keeps same project ID) - RECOMMENDED
    # "preserve" = financial history stays with the original OU (WARNING: creates new project ID)
    financial_setting = "move"
  }
}

# =============================================================================
# Move Settings Reference
# =============================================================================
#
# | Setting              | Value       | Description                                                |
# |----------------------|-------------|------------------------------------------------------------|
# | cloud_rule_setting   | "convert"   | Inherited cloud rules become local rules on the project    |
# | cloud_rule_setting   | "remove"    | Cloud rules are removed from the project                   |
# | financial_setting    | "move"      | Financial history moves with project (keeps same ID)       |
# | financial_setting    | "preserve"  | Financial history stays in old OU (WARNING: new project ID)|
#
# Note: Any accounts attached to the project will automatically move with it.
