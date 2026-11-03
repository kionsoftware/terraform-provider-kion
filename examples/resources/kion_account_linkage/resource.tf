resource "kion_account_linkage" "example" {
  # Required
  azure_object_id      = "example"
  azure_principal_name = "example"
  payer_id             = 1
  user_id              = 1
}
