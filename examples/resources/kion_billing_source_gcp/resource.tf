resource "kion_billing_source_gcp" "example" {
  # Required
  account_type_id            = 1
  gcp_billing_account_create = {
    big_query_export = {
      # dataset_name    = "example"
      # focus_view_name = "example"
      # gcp_project_id  = "example"
      # table_format    = "example"
      # table_name      = "example"
    }
    billing_start_date = "2026-01"
    gcp_id             = "example"
    name               = "example"
    service_account_id = 1
    # billing_account_attribution_account_id = 1
    # is_reseller                            = false
    # use_focus                              = false
    # use_proprietary                        = false
  }
}
