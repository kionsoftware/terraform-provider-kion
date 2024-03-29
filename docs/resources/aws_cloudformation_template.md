---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_aws_cloudformation_template Resource - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_aws_cloudformation_template (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)
- `policy` (String)
- `regions` (Set of String)

### Optional

- `description` (String)
- `last_updated` (String)
- `owner_user_groups` (Block Set) Must provide at least the owner_user_groups field or the owner_users field. (see [below for nested schema](#nestedblock--owner_user_groups))
- `owner_users` (Block Set) Must provide at least the owner_user_groups field or the owner_users field. (see [below for nested schema](#nestedblock--owner_users))
- `region` (String)
- `sns_arns` (String)
- `tags` (Map of String) Stack-level tags will apply to all supported resources in a CloudFormation stack.  Requires Kion >= 3.7.1.
- `template_parameters` (String)
- `termination_protection` (Boolean)

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--owner_user_groups"></a>
### Nested Schema for `owner_user_groups`

Read-Only:

- `id` (Number) The ID of this resource.


<a id="nestedblock--owner_users"></a>
### Nested Schema for `owner_users`

Read-Only:

- `id` (Number) The ID of this resource.
