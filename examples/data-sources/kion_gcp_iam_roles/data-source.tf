data "kion_gcp_iam_roles" "all" {}

output "gcp_iam_roles" {
  value = data.kion_gcp_iam_roles.all.data
}
