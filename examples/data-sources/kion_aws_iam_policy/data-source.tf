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

# Example 1: Get all IAM policies
data "kion_aws_iam_policy" "all" {
}

# Example 2: Search for AWS managed policies by name
data "kion_aws_iam_policy" "aws_managed" {
  query       = "Administrator"
  policy_type = "aws"
}

# Example 3: Search for user-created policies
data "kion_aws_iam_policy" "user_policies" {
  policy_type = "user"
}

# Example 4: Use pagination
data "kion_aws_iam_policy" "paginated" {
  page      = 1
  page_size = 10
}

# Example 5: Filter policies by name
data "kion_aws_iam_policy" "by_name" {
  filter {
    name   = "name"
    values = ["ReadOnlyAccess"]
  }
}

# Example 6: Filter policies by multiple criteria
data "kion_aws_iam_policy" "multi_filter" {
  filter {
    name   = "name"
    values = ["Admin", "Administrator"]
    regex  = true
  }

  filter {
    name   = "aws_managed_policy"
    values = ["true"]
  }
}

# Example 7: Filter by owner
data "kion_aws_iam_policy" "by_owner" {
  filter {
    name   = "owner_users.id"
    values = ["1", "2"]
  }
}

# Example 8: Combine query, policy type and filter
data "kion_aws_iam_policy" "combined" {
  query       = "S3"
  policy_type = "aws"

  filter {
    name   = "description"
    values = ["bucket"]
    regex  = true
  }
}

# Output examples
output "all_policies" {
  value = data.kion_aws_iam_policy.all.list
}

output "aws_admin_policies" {
  value = data.kion_aws_iam_policy.aws_managed.list
}

# Output specific fields from filtered policies
output "filtered_policy_names" {
  value = [for policy in data.kion_aws_iam_policy.multi_filter.list : policy.name]
}

# Output policy details in a map format
output "policy_map" {
  value = {
    for policy in data.kion_aws_iam_policy.all.list :
    policy.id => {
      name        = policy.name
      description = policy.description
      is_aws     = policy.aws_managed_policy
    }
  }
}

# Output count of policies found
output "total_policies" {
  value = length(data.kion_aws_iam_policy.all.list)
}

# Output first matching policy's full details
output "first_matching_policy" {
  value = try(data.kion_aws_iam_policy.by_name.list[0], null)
}
