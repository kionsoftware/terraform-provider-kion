# Example: Basic AWS billing source with CUR reports
resource "kion_billing_source_aws" "example" {
  name               = "Production AWS Billing"
  aws_account_number = "123456789012"
  billing_start_date = "2024-01"
  account_creation   = true
  
  # CUR configuration
  billing_report_type = "cur"
  cur_bucket          = "my-billing-reports-bucket"
  cur_bucket_region   = "us-east-1"
  cur_name            = "my-cost-and-usage-report"
  cur_prefix          = "reports"
}

# Example: AWS billing source with FOCUS reports
resource "kion_billing_source_aws" "focus_example" {
  name               = "AWS with FOCUS Reports"
  aws_account_number = "987654321098"
  billing_start_date = "2024-01"
  
  # FOCUS billing configuration
  focus_billing_bucket_account_number = "987654321098"
  focus_billing_report_bucket         = "my-focus-reports-bucket"
  focus_billing_report_bucket_region  = "us-east-1"
  focus_billing_report_name           = "my-focus-report"
  focus_billing_report_prefix         = "focus-reports"
}

# Example: AWS billing source with IAM role access
resource "kion_billing_source_aws" "role_based" {
  name                          = "AWS with IAM Role Access"
  aws_account_number            = "111222333444"
  billing_bucket_account_number = "555666777888" # Different account holds the billing data
  billing_start_date            = "2024-01"
  
  # Use IAM role instead of access keys
  bucket_access_role = "BillingReportAccessRole"
  linked_role        = "OrganizationAccountAccessRole"
  
  # CUR configuration
  billing_report_type = "cur"
  cur_bucket          = "cross-account-billing-reports"
  cur_bucket_region   = "us-east-1"
  cur_name            = "organization-cur-report"
  cur_prefix          = "cur"
}

# Example: AWS billing source with access keys
resource "kion_billing_source_aws" "key_based" {
  name               = "AWS with Access Keys"
  aws_account_number = "999888777666"
  billing_start_date = "2024-01"
  
  # Authentication via access keys
  key_id     = var.aws_access_key_id
  key_secret = var.aws_secret_access_key
  
  # Skip validation during creation
  skip_validation = true
  
  # DBR configuration
  billing_report_type = "dbrrt"
  mr_bucket           = "detailed-billing-reports"
  billing_region      = "us-west-2"
}