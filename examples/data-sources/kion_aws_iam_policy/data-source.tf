
# Declare a data source to get all IAM policies.
data "kion_aws_iam_policy" "p1" {}

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
