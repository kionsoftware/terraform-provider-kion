data "kion_azure_roles" "all" {}

output "azure_roles" {
  value = data.kion_azure_roles.all.data
}
