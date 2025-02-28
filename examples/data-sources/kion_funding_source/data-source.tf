# Get all funding sources
data "kion_funding_source" "all" {
}

# Filter funding sources by name
data "kion_funding_source" "by_name" {
  filter {
    name   = "name"
    values = ["Annual Cloud Budget 2024"]
  }
}

# Filter by OU ID
data "kion_funding_source" "by_ou" {
  filter {
    name   = "ou_id"
    values = ["1"]
  }
}

# Filter by date range
data "kion_funding_source" "current_quarter" {
  filter {
    name   = "start_datecode"
    values = ["2024-01"]
  }
  filter {
    name   = "end_datecode"
    values = ["2024-03"]
  }
}

# Filter by multiple criteria including amount
data "kion_funding_source" "large_budgets" {
  filter {
    name   = "amount"
    values = ["100000"]  # Budgets over 100k
    regex  = true
  }
  filter {
    name   = "permission_scheme_id"
    values = ["1"]
  }
}

# Example outputs
output "all_funding_sources" {
  value = data.kion_funding_source.all.list
}

output "annual_budget_details" {
  value = data.kion_funding_source.by_name.list[0]
}

output "ou_funding_sources" {
  value = [
    for fs in data.kion_funding_source.by_ou.list : {
      name   = fs.name
      amount = fs.amount
    }
  ]
}

output "q1_budgets" {
  value = [
    for fs in data.kion_funding_source.current_quarter.list : {
      name        = fs.name
      amount      = fs.amount
      start_date  = fs.start_datecode
      end_date    = fs.end_datecode
    }
  ]
}

output "large_budget_summary" {
  value = [
    for fs in data.kion_funding_source.large_budgets.list : {
      name           = fs.name
      amount         = fs.amount
      ou_id          = fs.ou_id
      owner_users    = fs.owner_users
      owner_groups   = fs.owner_user_groups
    }
  ]
}