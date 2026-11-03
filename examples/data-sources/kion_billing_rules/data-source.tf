data "kion_billing_rules" "all" {}

output "billing_rules" {
  value = data.kion_billing_rules.all.items
}
