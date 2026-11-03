# Plan-only. The provider is configured with a placeholder endpoint and is
# never contacted, so no Kion credentials are required.

provider "kion" {
  api_url = "http://127.0.0.1:1"
  api_key = "test"
}

run "plan" {
  command = plan

  variables {
    account_type_id    = 1
    aws_account_number = "example"
    billing_start_date = "2026-01"
    linked_role        = "example"
    name               = "example"
  }
}
