data "kion_service_catalogs" "all" {}

output "service_catalogs" {
  value = data.kion_service_catalogs.all.data
}
