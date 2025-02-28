# Get all GCP IAM roles
data "kion_gcp_iam_role" "all" {
}

# Filter roles by name
data "kion_gcp_iam_role" "by_name" {
  filter {
    name   = "name"
    values = ["Custom Admin Role"]
  }
}

# Filter by description with regex
data "kion_gcp_iam_role" "admin_roles" {
  filter {
    name   = "description"
    values = [".*admin.*"]
    regex  = true
  }
}

# Filter by launch stage
data "kion_gcp_iam_role" "ga_roles" {
  filter {
    name   = "gcp_role_launch_stage"
    values = ["4"]  # GA stage
  }
}

# Filter by system managed status
data "kion_gcp_iam_role" "custom_roles" {
  filter {
    name   = "system_managed_policy"
    values = ["false"]
  }
}

# Filter by multiple criteria
data "kion_gcp_iam_role" "filtered_roles" {
  filter {
    name   = "gcp_managed_policy"
    values = ["false"]
  }
  filter {
    name   = "system_managed_policy"
    values = ["false"]
  }
}

# Example outputs
output "all_roles" {
  value = data.kion_gcp_iam_role.all.list
}

output "admin_role_details" {
  value = data.kion_gcp_iam_role.by_name.list[0]
}

output "admin_role_names" {
  value = [
    for role in data.kion_gcp_iam_role.admin_roles.list : {
      name        = role.name
      description = role.description
      gcp_id      = role.gcp_id
    }
  ]
}

output "ga_role_summary" {
  value = [
    for role in data.kion_gcp_iam_role.ga_roles.list : {
      name             = role.name
      role_permissions = role.role_permissions
    }
  ]
}

output "custom_role_ids" {
  value = [
    for role in data.kion_gcp_iam_role.custom_roles.list : {
      name    = role.name
      id      = role.id
      gcp_id  = role.gcp_id
    }
  ]
}

output "filtered_role_count" {
  value = length(data.kion_gcp_iam_role.filtered_roles.list)
}