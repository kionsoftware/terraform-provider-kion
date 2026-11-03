# Plan-only. The provider is configured with a placeholder endpoint and is
# never contacted, so no Kion credentials are required.

provider "kion" {
  api_url = "http://127.0.0.1:1"
  api_key = "test"
}

run "plan" {
  command = plan

  variables {
    create_user_id = 1
    name           = "example"
    ou_id          = 1
    text           = "example"
  }
}
