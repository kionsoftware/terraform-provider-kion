data "kion_app_api_keys" "all" {}

output "api_keys" {
  value = data.kion_app_api_keys.all.data
}
