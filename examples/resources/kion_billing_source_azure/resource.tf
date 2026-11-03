resource "kion_billing_source_azure" "example" {
  # Required
  account_type_id    = 1
  billing_start_date = "example"
  name               = "example"

  # Optional
  # account_creation                           = false
  # app_id                                     = "example"
  # azure_csp_id                               = 1
  # client_secret                              = "example"
  # customer_id                                = "example"
  # domain                                     = "example"
  # focus_storage_container                    = "example"
  # focus_storage_prefix                       = "example"
  # focus_storage_primary_endpoint             = "example"
  # primary_management_group_azure_id          = "example"
  # resource_group_creation                    = false
  # skip_validation                            = false
  # storage_container                          = "example"
  # storage_prefix                             = "example"
  # storage_primary_endpoint                   = "example"
  # subscription_creation                      = false
  # tenant_level_charge_attribution_account_id = 1
  # use_focus_reports                          = false
  # use_proprietary_reports                    = false
}
