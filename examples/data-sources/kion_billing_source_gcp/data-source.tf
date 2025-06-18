# Read an existing GCP billing source by ID
data "kion_billing_source" "gcp_billing" {
  id = 123
}

# Example outputs to show available GCP billing source information
output "gcp_billing_name" {
  value = data.kion_billing_source.gcp_billing.gcp_payer[0].name
}

output "gcp_billing_account_id" {
  value = data.kion_billing_source.gcp_billing.gcp_payer[0].gcp_id
}

output "gcp_service_account_id" {
  value = data.kion_billing_source.gcp_billing.gcp_payer[0].service_account_id
}

output "bigquery_project" {
  value = data.kion_billing_source.gcp_billing.gcp_payer[0].gcp_billing_account[0].big_query_export[0].gcp_project_id
}

output "bigquery_dataset" {
  value = data.kion_billing_source.gcp_billing.gcp_payer[0].gcp_billing_account[0].big_query_export[0].dataset_name
}