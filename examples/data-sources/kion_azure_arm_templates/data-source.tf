data "kion_azure_arm_templates" "all" {}

output "arm_templates" {
  value = data.kion_azure_arm_templates.all.data
}
