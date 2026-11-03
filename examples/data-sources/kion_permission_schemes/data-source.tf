data "kion_permission_schemes" "all" {}

output "permission_schemes" {
  value = data.kion_permission_schemes.all.data
}
