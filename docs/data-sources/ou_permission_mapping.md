---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_ou_permission_mapping Data Source - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_ou_permission_mapping (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `ou_id` (Number) ID of the OU to fetch permission mappings for.

### Read-Only

- `id` (String) The ID of this resource.
- `list` (List of Object) List of permission mappings. (see [below for nested schema](#nestedatt--list))

<a id="nestedatt--list"></a>
### Nested Schema for `list`

Read-Only:

- `app_role_id` (Number)
- `user_groups_ids` (Set of Number)
- `user_ids` (Set of Number)
