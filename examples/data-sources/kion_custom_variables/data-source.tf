data "kion_custom_variables" "all" {}

output "custom_variables" {
  value = data.kion_custom_variables.all.items
}
