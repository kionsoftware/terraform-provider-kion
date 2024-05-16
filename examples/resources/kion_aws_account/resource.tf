# Create a new AWS account and place it in the account cache
resource "kion_aws_account" "test2" {
  name                    = "Terraform Created AWS Account - 2"
  payer_id                = 1
  commercial_account_name = "Test Account"
  create_govcloud         = false

  aws_organizational_unit {
    name        = "test name"
    org_unit_id = "123456"
  }
}

# Output the ID of the resource created.
output "kion_account_id" {
  value = kion_aws_account.test2.id
}
