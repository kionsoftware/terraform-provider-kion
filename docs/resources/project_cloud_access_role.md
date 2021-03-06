---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_project_cloud_access_role Resource - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_project_cloud_access_role (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **aws_iam_role_name** (String)
- **name** (String)
- **project_id** (Number)

### Optional

- **accounts** (Block List) This field will be ignored if 'apply_to_all_accounts' is set to: true. (see [below for nested schema](#nestedblock--accounts))
- **apply_to_all_accounts** (Boolean)
- **aws_iam_path** (String)
- **aws_iam_permissions_boundary** (Number)
- **aws_iam_policies** (Block List) (see [below for nested schema](#nestedblock--aws_iam_policies))
- **azure_role_definitions** (Block List) (see [below for nested schema](#nestedblock--azure_role_definitions))
- **future_accounts** (Boolean)
- **id** (String) The ID of this resource.
- **last_updated** (String)
- **long_term_access_keys** (Boolean)
- **short_term_access_keys** (Boolean)
- **user_groups** (Block List) (see [below for nested schema](#nestedblock--user_groups))
- **users** (Block List) (see [below for nested schema](#nestedblock--users))
- **web_access** (Boolean)

<a id="nestedblock--accounts"></a>
### Nested Schema for `accounts`

Optional:

- **id** (Number) The ID of this resource.


<a id="nestedblock--aws_iam_policies"></a>
### Nested Schema for `aws_iam_policies`

Optional:

- **id** (Number) The ID of this resource.


<a id="nestedblock--azure_role_definitions"></a>
### Nested Schema for `azure_role_definitions`

Optional:

- **id** (Number) The ID of this resource.


<a id="nestedblock--user_groups"></a>
### Nested Schema for `user_groups`

Optional:

- **id** (Number) The ID of this resource.


<a id="nestedblock--users"></a>
### Nested Schema for `users`

Optional:

- **id** (Number) The ID of this resource.


