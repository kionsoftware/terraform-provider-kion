# Plan-only. The provider is configured with a placeholder endpoint and is
# never contacted, so no Kion credentials are required.

provider "kion" {
  api_url = "http://127.0.0.1:1"
  api_key = "test"
}

run "plan" {
  command = plan

  variables {
    email      = "user@example.com"
    first_name = "example"
    idms_id    = 1
    last_name  = "example"
    username   = "example"
  }
}
