data "kion_cloud_rules" "all" {}

output "cloud_rules" {
  value = data.kion_cloud_rules.all.data
}
