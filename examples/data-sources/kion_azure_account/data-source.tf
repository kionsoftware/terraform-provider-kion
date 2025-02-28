# Example 1: Get all Azure accounts
data "kion_azure_account" "all_accounts" {
}

# Example 2: Filter Azure accounts by name
data "kion_azure_account" "dev_accounts" {
  filter {
    name   = "name"
    values = ["dev-", "development-"]
    regex  = true
  }
}

# Example 3: Filter by multiple criteria
data "kion_azure_account" "filtered_accounts" {
  filter {
    name   = "account_type_id"
    values = ["1"]
  }
  filter {
    name   = "payer_id"
    values = ["3"]
  }
  filter {
    name   = "skip_access_checking"
    values = ["false"]
  }
}

# Example 4: Filter by project ID
data "kion_azure_account" "project_accounts" {
  filter {
    name   = "project_id"
    values = ["42"]
  }
}

# Example 5: Filter by account alias
data "kion_azure_account" "aliased_accounts" {
  filter {
    name   = "account_alias"
    values = ["prod-"]
    regex  = true
  }
}

# Example 6: Filter by creation date
data "kion_azure_account" "recent_accounts" {
  filter {
    name   = "created_at"
    values = ["2024-"]
    regex  = true
  }
}

# Output examples
output "all_azure_accounts" {
  description = "List of all Azure accounts"
  value       = data.kion_azure_account.all_accounts.list
}

output "dev_account_names" {
  description = "Names of development Azure accounts"
  value       = [for account in data.kion_azure_account.dev_accounts.list : account.name]
}

output "filtered_account_details" {
  description = "Details of filtered Azure accounts"
  value = {
    for account in data.kion_azure_account.filtered_accounts.list :
    account.name => {
      id                   = account.id
      subscription_uuid    = account.subscription_uuid
      account_alias       = account.account_alias
      payer_id           = account.payer_id
      project_id         = account.project_id
      skip_access_checking = account.skip_access_checking
      created_at         = account.created_at
    }
  }
}

output "project_account_count" {
  description = "Number of Azure accounts in the specified project"
  value       = length(data.kion_azure_account.project_accounts.list)
}

output "account_statistics" {
  description = "Statistics about Azure accounts"
  value = {
    total_accounts     = length(data.kion_azure_account.all_accounts.list)
    dev_accounts      = length(data.kion_azure_account.dev_accounts.list)
    aliased_accounts  = length(data.kion_azure_account.aliased_accounts.list)
    recent_accounts   = length(data.kion_azure_account.recent_accounts.list)
  }
}