---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_user_group Resource - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_user_group (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **idms_id** (Number)
- **name** (String)

### Optional

- **description** (String)
- **id** (String) The ID of this resource.
- **last_updated** (String)
- **owner_groups** (Block List) (see [below for nested schema](#nestedblock--owner_groups))
- **owner_users** (Block List) (see [below for nested schema](#nestedblock--owner_users))
- **users** (Block List) (see [below for nested schema](#nestedblock--users))

### Read-Only

- **created_at** (String)
- **enabled** (Boolean)

<a id="nestedblock--owner_groups"></a>
### Nested Schema for `owner_groups`

Optional:

- **id** (Number) The ID of this resource.


<a id="nestedblock--owner_users"></a>
### Nested Schema for `owner_users`

Optional:

- **id** (Number) The ID of this resource.


<a id="nestedblock--users"></a>
### Nested Schema for `users`

Optional:

- **id** (Number) The ID of this resource.


