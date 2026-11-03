data "kion_cfts" "all" {}

output "cloudformation_templates" {
  value = data.kion_cfts.all.data
}
