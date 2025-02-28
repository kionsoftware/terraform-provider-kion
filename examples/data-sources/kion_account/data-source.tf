# Get all accounts
data "kion_account" "all" {
}

# Filter accounts by name
data "kion_account" "by_name" {
  filter {
    name   = "name"
    values = ["Production Account"]
  }
}

# Filter accounts by account number with regex
data "kion_account" "by_account_number" {
  filter {
    name   = "account_number"
    values = ["^123.*"]  # Matches account numbers starting with 123
    regex  = true
  }
}

# Filter accounts by multiple criteria
data "kion_account" "by_multiple" {
  filter {
    name   = "project_id"
    values = ["42"]
  }
  filter {
    name   = "account_type_id"
    values = ["1"]
  }
}

# Example outputs
output "all_accounts" {
  value = data.kion_account.all.list
}

output "production_account_id" {
  value = data.kion_account.by_name.list[0].id
}

output "filtered_account_numbers" {
  value = [for account in data.kion_account.by_account_number.list : account.account_number]
}

output "project_accounts" {
  value = data.kion_account.by_multiple.list
}