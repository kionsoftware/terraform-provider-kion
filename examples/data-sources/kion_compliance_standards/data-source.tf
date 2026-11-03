data "kion_compliance_standards" "all" {}

output "compliance_standards" {
  value = data.kion_compliance_standards.all.data
}
