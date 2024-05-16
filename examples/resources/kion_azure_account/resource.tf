resource "kion_azure_account" "test6" {
  name              = "Terraform Created Azure Subscription - 5"
  subscription_name = "terraform-test-create"
  payer_id          = 3
  mca {
    billing_account         = "5e98e158-xxxx-xxxx-xxxx-xxxxxxxxxxxx:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx_xxxx-xx-xx"
    billing_profile         = "AW4F-xxxx-xxx-xxx"
    billing_profile_invoice = "SH3V-xxxx-xxx-xxx"
  }
}

# Output the ID of the resource created.
output "kion_azure_account_id" {
  value = kion_azure_account.test6.id
}
