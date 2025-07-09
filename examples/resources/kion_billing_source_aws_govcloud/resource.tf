# Create an AWS GovCloud billing source attached to a commercial billing source
resource "kion_billing_source_aws_govcloud" "example" {
  commercial_billing_source_id = 123  # ID of the commercial AWS billing source
  name                         = "GovCloud Billing Account"
  aws_account_number           = "123456789012"
  account_creation_enabled     = true
}

# Example with data source for commercial billing source
data "kion_funding_source" "commercial" {
  filter {
    name   = "name"
    values = ["Commercial AWS Billing Source"]
  }
}

resource "kion_billing_source_aws_govcloud" "example_with_data" {
  commercial_billing_source_id = data.kion_funding_source.commercial.list[0].id
  name                         = "GovCloud Billing Account"
  aws_account_number           = "123456789012"
  account_creation_enabled     = false
}