data "kion_funding_sources" "all" {}

output "funding_sources" {
  value = data.kion_funding_sources.all.data
}
