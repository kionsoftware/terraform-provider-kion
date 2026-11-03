data "kion_projects" "all" {}

output "projects" {
  value = data.kion_projects.all.data
}
