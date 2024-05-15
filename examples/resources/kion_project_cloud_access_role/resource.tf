resource "kion_project_cloud_access_role" "carp1" {
  name                         = "sample-car"
  project_id                   = 1
  aws_iam_role_name            = "sample-car"
  web_access                   = true
  short_term_access_keys       = true
  long_term_access_keys        = true
  aws_iam_permissions_boundary = 1
  future_accounts              = true
  aws_iam_policies { id = 1 }
  #accounts { id = 1 }
  users { id = 1 }
  user_groups { id = 1 }
}

# Output the ID of the resource created.
output "project_car_id" {
  value = kion_project_cloud_access_role.carp1.id
}
