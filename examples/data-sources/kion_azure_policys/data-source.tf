data "kion_azure_policys" "all" {}

output "azure_policies" {
  value = data.kion_azure_policys.all.data
}
