data "kion_iam_policys" "all" {}

output "iam_policies" {
  value = data.kion_iam_policys.all.data
}
