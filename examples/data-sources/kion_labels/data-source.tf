data "kion_labels" "all" {}

output "labels" {
  value = data.kion_labels.all.items
}
