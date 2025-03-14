---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_azure_arm_template Data Source - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_azure_arm_template (Data Source)



## Example Usage

```terraform
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `filter` (Block List) (see [below for nested schema](#nestedblock--filter))

### Read-Only

- `id` (String) The ID of this resource.
- `list` (List of Object) This is where Kion makes the discovered data available as a list of resources. (see [below for nested schema](#nestedatt--list))

<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

Required:

- `name` (String) The field name whose values you wish to filter by.
- `values` (List of String) The values of the field name you specified.

Optional:

- `regex` (Boolean) Dictates if the values provided should be treated as regular expressions.


<a id="nestedatt--list"></a>
### Nested Schema for `list`

Read-Only:

- `ct_managed` (Boolean)
- `deployment_mode` (Number)
- `description` (String)
- `id` (Number)
- `name` (String)
- `owner_user_groups` (List of Object) (see [below for nested schema](#nestedobjatt--list--owner_user_groups))
- `owner_users` (List of Object) (see [below for nested schema](#nestedobjatt--list--owner_users))
- `resource_group_name` (String)
- `resource_group_region_id` (Number)
- `template` (String)
- `template_parameters` (String)
- `version` (Number)

<a id="nestedobjatt--list--owner_user_groups"></a>
### Nested Schema for `list.owner_user_groups`

Read-Only:

- `id` (Number)


<a id="nestedobjatt--list--owner_users"></a>
### Nested Schema for `list.owner_users`

Read-Only:

- `id` (Number)
