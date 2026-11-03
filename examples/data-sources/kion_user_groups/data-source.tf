data "kion_user_groups" "all" {}

output "user_groups" {
  value = data.kion_user_groups.all.data
}
