data "kion_webhooks" "all" {}

output "webhooks" {
  value = data.kion_webhooks.all.data
}
