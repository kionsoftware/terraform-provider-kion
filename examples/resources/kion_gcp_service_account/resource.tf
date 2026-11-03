resource "kion_gcp_service_account" "example" {
  # Required
  email                     = "example"
  enable_federation_support = false
  gcp_project_id            = "example"
  name                      = "example"
  unique_id                 = "example"

  # Optional
  # description         = "example"
  # disabled            = false
  # display_name        = "example"
  # oauth_client_id     = "example"
  # oauth_client_secret = "example"
}
