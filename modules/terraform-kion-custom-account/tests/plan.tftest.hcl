# Plan-only. The provider is configured with a placeholder endpoint and is
# never contacted, so no Kion credentials are required.

provider "kion" {
  api_url = "http://127.0.0.1:1"
  api_key = "test"
}

run "plan" {
  command = plan

  variables {
    account_name   = "example"
    account_number = "example"
    payer_id       = 1
    project_id     = 1
    start_datecode = "2026-01"
  }
}
