# Create a new GCP project and place it in a Kion project:
resource "kion_gcp_account" "test4" {
  create_mode             = "create"
  name                    = "Terraform Created GCP Project - 4"
  google_cloud_project_id = "terraform-test-create"
  payer_id                = 2
  project_id              = kion_project.test1.id
  start_datecode          = "2023-01"
}

# Output the ID of the resource created.
output "kion_project_id" {
  value = kion_gcp_account.test4.id
}
