data "kion_ous" "all" {}

output "organizational_units" {
  value = data.kion_ous.all.data
}
