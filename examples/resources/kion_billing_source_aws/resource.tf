resource "kion_billing_source_aws" "example" {
  # Required
  account_type_id    = 1
  aws_account_number = "example"
  billing_start_date = "2026-01"
  linked_role        = "example"
  name               = "example"

  # Optional
  # account_creation                    = false
  # billing_bucket_account_number       = "example"
  # billing_region                      = "example"
  # billing_report_type                 = "example"
  # bucket_access_role                  = "example"
  # cur_bucket                          = "example"
  # cur_bucket_region                   = "example"
  # cur_name                            = "example"
  # cur_prefix                          = "example"
  # focus_billing_bucket_account_number = "example"
  # focus_billing_report_bucket         = "example"
  # focus_billing_report_bucket_region  = "example"
  # focus_billing_report_name           = "example"
  # focus_billing_report_prefix         = "example"
  # focus_bucket_access_role            = "example"
  # key_id                              = "example"
  # key_secret                          = "example"
  # mr_bucket                           = "example"
  # only_dbr                            = false
  # skip_validation                     = false
  # use_focus_reports                   = false
}
