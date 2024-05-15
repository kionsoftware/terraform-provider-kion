# Create a GCP IAM role.
resource "kion_gcp_iam_role" "gr1" {
  name                  = "Read permissions"
  description           = "Allow user to list & get IAM roles."
  role_permissions      = ["iam.roles.get", "iam.roles.list"]
  gcp_role_launch_stage = 4
  owner_users { id = 1 }
}

# Output the ID of the resource created.
output "gcp_role" {
  value = kion_gcp_iam_role.gr1.id
}
