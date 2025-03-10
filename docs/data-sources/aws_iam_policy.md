---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_aws_iam_policy Data Source - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_aws_iam_policy (Data Source)



## Example Usage

```terraform
# Declare a data source to get all IAM policies.
data "kion_aws_iam_policy" "p1" {}

# Use v4 query parameter to search for policies by name
data "kion_aws_iam_policy" "read_only" {
  query = "ReadOnly"
}

# Use v4 policy_type filter to get only AWS managed policies
data "kion_aws_iam_policy" "aws_managed" {
  policy_type = "aws"
}

# Use v4 pagination
data "kion_aws_iam_policy" "paginated" {
  page      = 1
  page_size = 50
}

# Combine v4 filtering with existing filter blocks
data "kion_aws_iam_policy" "combined" {
  query       = "Admin"
  policy_type = "user"

  filter {
    name   = "owner_users.id"
    values = ["20"]
  }
}

# Output the list of all policies.
output "policies" {
  value = data.kion_aws_iam_policy.p1.list
}

# Output the first policy (which by default is the newest policy).
output "first" {
  value = data.kion_aws_iam_policy.p1.list[0]
}

# Output the first policy name.
output "policy_name" {
  value = data.kion_aws_iam_policy.p1.list[0].name
}

# Output a list of all policy names.
output "policy_names" {
  value = data.kion_aws_iam_policy.p1.list.*.name
}

# Output a list of all owner users for all policies.
output "policy_owner_users" {
  value = data.kion_aws_iam_policy.p1.list.*.owner_users
}

# Declare a data source to get 1 IAM policy that matches the name filter.
data "kion_aws_iam_policy" "p1" {
  filter {
    name   = "name"
    values = ["SystemReadOnlyAccess"]
  }
}

# Declare a data source to get 2 IAM policies that matches the name filter.
data "kion_aws_iam_policy" "p1" {
  filter {
    name   = "name"
    values = ["SystemReadOnlyAccess", "test-policy"]
  }
}

# Declare a data source to get 1 IAM policy that matches both of the filters.
# SystemReadOnlyAccess has the id of 1 so only that policy matches all of the filters.
data "kion_aws_iam_policy" "p1" {
  filter {
    name   = "name"
    values = ["SystemReadOnlyAccess", "test-policy"]
  }

  filter {
    name   = "id"
    values = [1]
  }
}

# Declare a data source to get all IAM policies that matches the owner filter.
# Syntax to filter on an array.
data "kion_aws_iam_policy" "p1" {
  filter {
    name   = "owner_users.id"
    values = ["20"]
  }
}

# Declare a data source to get all IAM policies that matches the id filter.
# Notice that terraform will convert these to strings even though you
# passed in an integer.
data "kion_aws_iam_policy" "p1" {
  filter {
    name   = "id"
    values = [1, "3"]
    # Terraform will convert all of these to strings.
    # + values = [
    #     + "1",
    #     + "3",
    #   ]
  }
}

# Declare a data source to get all IAM policies that matches the query.
output "policy_access" {
  value = {
    # Loop through each policy
    for k in data.kion_aws_iam_policy.p1.list :
    # Create a map with a key of: id
    k.id => k
    # Filter out an names that don't match the passed in variable
    if k.name == "SystemReadOnlyAccess"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `filter` (Block List) (see [below for nested schema](#nestedblock--filter))
- `page` (Number) Page number of results
- `page_size` (Number) Number of results per page
- `policy_type` (String) Policy type filter. Valid values are 'user', 'aws', or 'system'
- `query` (String) Query string for IAM policy name matching

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

- `aws_iam_path` (String)
- `aws_managed_policy` (Boolean)
- `description` (String)
- `id` (Number)
- `name` (String)
- `owner_user_groups` (List of Object) (see [below for nested schema](#nestedobjatt--list--owner_user_groups))
- `owner_users` (List of Object) (see [below for nested schema](#nestedobjatt--list--owner_users))
- `path_suffix` (String)
- `policy` (String)
- `system_managed_policy` (Boolean)

<a id="nestedobjatt--list--owner_user_groups"></a>
### Nested Schema for `list.owner_user_groups`

Read-Only:

- `id` (Number)


<a id="nestedobjatt--list--owner_users"></a>
### Nested Schema for `list.owner_users`

Read-Only:

- `id` (Number)
