data "kion_idmss" "all" {}

output "identity_management_systems" {
  value = data.kion_idmss.all.data
}
