data "kion_account_caches" "all" {}

output "cached_accounts" {
  value = data.kion_account_caches.all.data
}
