# Plan-only. The provider is configured with a placeholder endpoint and is
# never contacted, so no Kion credentials are required.

provider "kion" {
  api_url = "http://127.0.0.1:1"
  api_key = "test"
}

run "plan" {
  command = plan

  variables {
    account_type_id            = 1
    gcp_billing_account_create = { big_query_export = {}, billing_start_date = "2026-01", gcp_id = "example", name = "example", service_account_id = 1 }
  }
}
