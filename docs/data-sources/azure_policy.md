---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_azure_policy Data Source - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_azure_policy (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **filter** (Block List) (see [below for nested schema](#nestedblock--filter))
- **id** (String) The ID of this resource.

### Read-Only

- **list** (List of Object) (see [below for nested schema](#nestedatt--list))

<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

Required:

- **name** (String)
- **values** (List of String)

Optional:

- **regex** (Boolean)


<a id="nestedatt--list"></a>
### Nested Schema for `list`

Read-Only:

- **azure_managed_policy_def_id** (String)
- **ct_managed** (Boolean)
- **description** (String)
- **id** (Number)
- **name** (String)
- **owner_user_groups** (List of Object) (see [below for nested schema](#nestedobjatt--list--owner_user_groups))
- **owner_users** (List of Object) (see [below for nested schema](#nestedobjatt--list--owner_users))
- **parameters** (String)
- **policy** (String)

<a id="nestedobjatt--list--owner_user_groups"></a>
### Nested Schema for `list.owner_user_groups`

Read-Only:

- **id** (Number)


<a id="nestedobjatt--list--owner_users"></a>
### Nested Schema for `list.owner_users`

Read-Only:

- **id** (Number)


