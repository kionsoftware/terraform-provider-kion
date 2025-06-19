# WARNING: This data source uses the v1 API endpoint which is not yet stable.
# The public API endpoint will be available soon. Usage of this data source
# may result in breaking changes until the stable API is released.

# Get a specific spend report by ID (this is the only supported method)
data "kion_spend_report" "example" {
  id = 4
}

# Use the data source results
output "report_name" {
  value = length(data.kion_spend_report.example.spend_reports) > 0 ? data.kion_spend_report.example.spend_reports[0].report_name : null
}

output "report_dimension" {
  value = length(data.kion_spend_report.example.spend_reports) > 0 ? data.kion_spend_report.example.spend_reports[0].dimension : null
}

output "report_spend_type" {
  value = length(data.kion_spend_report.example.spend_reports) > 0 ? data.kion_spend_report.example.spend_reports[0].spend_type : null
}

output "is_scheduled" {
  value = length(data.kion_spend_report.example.spend_reports) > 0 ? data.kion_spend_report.example.spend_reports[0].scheduled : false
}