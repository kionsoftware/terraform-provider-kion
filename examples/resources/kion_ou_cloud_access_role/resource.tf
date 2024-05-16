# Create a cloud access role on an OU.
resource "kion_ou_cloud_access_role" "carou1" {
  name                   = "sample-car"
  ou_id                  = 3
  aws_iam_role_name      = "sample-car"
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = true
  aws_iam_policies { id = 628 }
  #aws_iam_permissions_boundary = 1
  users { id = 1 }
  user_groups { id = 1 }
}

# Output the ID of the resource created.
output "ou_car_id" {
  value = kion_ou_cloud_access_role.carou1.id
}
