data "kion_users" "all" {}

output "users" {
  value = data.kion_users.all.data
}
