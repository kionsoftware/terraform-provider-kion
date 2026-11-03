# Plan-only. The provider is configured with a placeholder endpoint and is
# never contacted, so no Kion credentials are required.

provider "kion" {
  api_url = "http://127.0.0.1:1"
  api_key = "test"
}

run "plan" {
  command = plan

  variables {
    email                     = "user@example.com"
    enable_federation_support = false
    gcp_project_id            = "example"
    name                      = "example"
    unique_id                 = "example"
  }
}
