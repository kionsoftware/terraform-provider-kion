data "kion_service_control_policys" "all" {}

output "service_control_policies" {
  value = data.kion_service_control_policys.all.data
}
