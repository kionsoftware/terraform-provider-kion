---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_azure_policy Resource - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_azure_policy (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **id** (String) The ID of this resource.
- **last_updated** (String)
- **owner_user_groups** (Block List) (see [below for nested schema](#nestedblock--owner_user_groups))
- **owner_users** (Block List) (see [below for nested schema](#nestedblock--owner_users))

### Read-Only

- **azure_managed_policy_def_id** (String)
- **ct_managed** (Boolean)
- **description** (String)
- **name** (String)
- **parameters** (String)
- **policy** (String)

<a id="nestedblock--owner_user_groups"></a>
### Nested Schema for `owner_user_groups`

Optional:

- **id** (Number) The ID of this resource.


<a id="nestedblock--owner_users"></a>
### Nested Schema for `owner_users`

Optional:

- **id** (Number) The ID of this resource.


