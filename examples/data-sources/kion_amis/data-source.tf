data "kion_amis" "all" {}

output "all_amis" {
  value = data.kion_amis.all.data
}
