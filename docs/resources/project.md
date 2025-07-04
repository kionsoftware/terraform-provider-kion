---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_project Resource - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_project (Resource)

## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)
- `ou_id` (Number)
- `permission_scheme_id` (Number)

### Optional

- `auto_pay` (Boolean)
- `budget` (Block Set) (see [below for nested schema](#nestedblock--budget))
- `default_aws_region` (String)
- `description` (String)
- `labels` (Map of String) A map of labels to assign to the project. The labels must already exist in Kion.
- `last_updated` (String)
- `owner_user_group_ids` (Block Set) Must provide at least the owner_user_groups field or the owner_users field. (see [below for nested schema](#nestedblock--owner_user_group_ids))
- `owner_user_ids` (Block Set) Must provide at least the owner_user_groups field or the owner_users field. (see [below for nested schema](#nestedblock--owner_user_ids))
- `project_funding` (Block Set) (see [below for nested schema](#nestedblock--project_funding))

### Read-Only

- `archived` (Boolean)
- `id` (String) The ID of this resource.

<a id="nestedblock--budget"></a>

### Nested Schema for `budget`

Required:

- `end_datecode` (String) Year and month the budget ends. This is an exclusive date.
- `start_datecode` (String) Year and month the budget starts.

Optional:

- `amount` (Number) Total amount for the budget. This is required if data is not specified. Budget entries are created between start_datecode and end_datecode (exclusive) with the amount evenly distributed across the months. When monthly data is provided, the sum of all monthly amounts must equal this value.
- `data` (Block Set) Total amount for the budget. This is required if data is not specified. Budget entries are created between start_datecode and end_datecode (exclusive) with the amount evenly distributed across the months. (see [below for nested schema](#nestedblock--budget--data))
- `funding_source_ids` (Set of Number) Funding source IDs to use when data is not specified. This value is ignored is data is specified. If specified, the amount is distributed evenly across months and funding sources. Funding sources will be processed in order from first to last.

<a id="nestedblock--budget--data"></a>

### Nested Schema for `budget.data`

Required:

- `amount` (Number) Amount of the budget entry in dollars.
- `datecode` (String) Year and month for the budget data entry (i.e 2023-01).

Optional:

- `funding_source_id` (Number) ID of funding source for the budget entry.
- `priority` (Number) Priority order of the budget entry. This is required if funding_source_id is specified

<a id="nestedblock--owner_user_group_ids"></a>

### Nested Schema for `owner_user_group_ids`

Optional:

- `id` (Number)

<a id="nestedblock--owner_user_ids"></a>

### Nested Schema for `owner_user_ids`

Optional:

- `id` (Number)

<a id="nestedblock--project_funding"></a>

### Nested Schema for `project_funding`

Optional:

- `amount` (Number)
- `end_datecode` (String)
- `funding_order` (Number)
- `funding_source_id` (Number)
- `start_datecode` (String)
