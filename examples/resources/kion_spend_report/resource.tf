# WARNING: This resource uses the v1 API endpoint which is not yet stable.
# The public API endpoint will be available soon. Usage of this resource
# may result in breaking changes until the stable API is released.

# Example: Basic spend report with monthly granularity
resource "kion_spend_report" "monthly_spend_by_account" {
  report_name      = "Monthly Spend by Account"
  global_visibility = true
  date_range       = "last_six_months"
  spend_type       = "billed"
  dimension        = "account"
  time_granularity_id = 1  # Monthly
  deduct_credits   = true
  deduct_refunds   = true
}

# Example: Daily spend report with custom date range
resource "kion_spend_report" "daily_spend_by_service" {
  report_name      = "Daily Spend by Service"
  global_visibility = true
  date_range       = "custom"
  start_date       = "2024-01-01"
  end_date         = "2024-01-31"
  spend_type       = "attributed"
  dimension        = "service"
  time_granularity_id = 2  # Daily
  
  # Filter by specific accounts
  account_ids = [123, 456]
  account_exclusive = false  # Include only these accounts
}

# Example: Scheduled spend report with email notifications
resource "kion_spend_report" "scheduled_project_spend" {
  report_name      = "Weekly Project Spend Report"
  global_visibility = false
  date_range       = "month"
  spend_type       = "attributed"
  dimension        = "project"
  time_granularity_id = 1  # Monthly
  
  # Schedule configuration
  scheduled = true
  scheduled_email_subject = "Weekly Project Spend Update"
  scheduled_email_message = "Please find attached the weekly project spend report."
  scheduled_file_types = [0, 1]  # CSV and Excel
  scheduled_file_orientation = 2
  
  scheduled_frequency {
    type = 1  # Weekly
    days_of_week = [1]  # Monday
    hour = 9
    minute = 0
    time_zone_identifier = "America/New_York"
    start_date = "2024-01-01T09:00:00Z"
  }
  
  # Send to external emails
  external_emails {
    email_address = "finance@example.com"
  }
  external_emails {
    email_address = "management@example.com"
  }
  
  # Restrict access to specific user groups
  owner_user_group_ids = [100, 101]
}

# Example: Spend report with OU scope and filters
resource "kion_spend_report" "ou_scoped_spend" {
  report_name      = "OU Spend Analysis"
  global_visibility = true
  date_range       = "year"
  spend_type       = "billed"
  dimension        = "cloudProvider"
  time_granularity_id = 1  # Monthly
  
  # Scope to specific OU
  scope = "ou"
  scope_id = 50
  
  # Include descendant OUs
  include_descendants = true
  
  # Filter by cloud providers
  cloud_provider_ids = [1, 2]  # AWS and Azure
  cloud_provider_exclusive = false
  
  # Filter by regions
  region_ids = [10, 11, 12]
  region_exclusive = true  # Exclude these regions
}

# Example: Funding source spend report
resource "kion_spend_report" "funding_source_spend" {
  report_name      = "Funding Source Utilization"
  global_visibility = true
  date_range       = "last_three_months"
  spend_type       = "attributed"
  dimension        = "fundingSource"
  time_granularity_id = 1  # Monthly (required for funding source)
  
  # Filter by specific funding sources
  funding_source_ids = [200, 201]
  funding_source_exclusive = false
  
  # Include specific projects
  project_ids = [300, 301, 302]
  project_exclusive = false
}