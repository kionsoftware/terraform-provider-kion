# Create a GCP billing source
# Note: You must first create a GCP service account through the Kion UI or API
# and obtain its ID to use in the service_account_id field
resource "kion_billing_source_gcp" "example" {
  name               = "My GCP Billing Account"
  service_account_id = 123  # ID of the GCP service account created in Kion
  gcp_id             = "012345-ABCDEF-GHIJKL"  # Your GCP billing account ID
  billing_start_date = "2024-01"

  # BigQuery export configuration - where GCP exports billing data
  big_query_export {
    gcp_project_id  = "my-billing-project"
    dataset_name    = "cloud_billing_export"
    table_name      = "gcp_billing_export_v1"
    table_format    = "standard"  # Options: auto, standard, detailed
    focus_view_name = "focus_view_v1"  # Optional: Only if using FOCUS
  }

  # Optional: Configure billing data format preferences
  use_focus       = true   # Use FOCUS format for cost data
  use_proprietary = true   # Use GCP proprietary billing format
  is_reseller     = false  # Set to true if this is a reseller billing account
}

# Example with minimal configuration
resource "kion_billing_source_gcp" "minimal" {
  name               = "Simple GCP Billing"
  service_account_id = 456  # ID of the GCP service account created in Kion
  gcp_id             = "987654-ZYXWVU-TSRQPO"
  billing_start_date = "2024-06"

  big_query_export {
    gcp_project_id = "billing-exports"
    dataset_name   = "billing_data"
    table_name     = "cost_export"
  }
}