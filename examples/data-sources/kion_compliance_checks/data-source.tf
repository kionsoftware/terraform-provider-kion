data "kion_compliance_checks" "all" {}

output "compliance_checks" {
  value = data.kion_compliance_checks.all.data
}
