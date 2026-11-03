data "kion_categorys" "all" {}

output "categories" {
  value = data.kion_categorys.all.data
}
