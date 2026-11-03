data "kion_accounts" "all" {}

output "all_accounts" {
  value = data.kion_accounts.all.data
}
