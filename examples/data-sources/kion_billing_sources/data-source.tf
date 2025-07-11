# Example: Get all billing sources
data "kion_billing_sources" "all" {
}

# Example: Filter billing sources by name
data "kion_billing_sources" "production" {
  filter {
    name   = "name"
    values = ["Production*"]
    regex  = true
  }
}

# Example: Filter billing sources by type
data "kion_billing_sources" "aws_sources" {
  filter {
    name   = "type"
    values = ["aws"]
  }
}

# Example: Filter billing sources that support account creation
data "kion_billing_sources" "account_creation_enabled" {
  filter {
    name   = "account_creation"
    values = ["true"]
  }
}

# Output examples
output "all_billing_sources" {
  value = data.kion_billing_sources.all.list
}

output "production_billing_sources" {
  value = data.kion_billing_sources.production.list
}

output "aws_billing_source_names" {
  value = [for source in data.kion_billing_sources.aws_sources.list : source.name]
}
