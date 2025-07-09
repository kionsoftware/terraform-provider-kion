# Retrieve a billing source by ID
data "kion_billing_source" "example" {
  id = 123
}

# Example output usage
output "billing_source_name" {
  value = data.kion_billing_source.example.aws_payer[0].name
}

output "billing_source_account_creation" {
  value = data.kion_billing_source.example.account_creation
}