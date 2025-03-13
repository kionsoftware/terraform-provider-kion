# Find all environment labels
data "kion_label" "environment_labels" {
  filter {
    name   = "key"
    values = ["Environment"]
  }
}

# Find specific environment label using exact match
data "kion_label" "production_label" {
  filter {
    name   = "key"
    values = ["Environment"]
  }
  filter {
    name   = "value"
    values = ["Production"]
  }
}

# Find all cost center labels using regex
data "kion_label" "it_cost_centers" {
  filter {
    name   = "key"
    values = ["CostCenter"]
  }
  filter {
    name   = "value"
    values = ["IT-.*"]  # Matches any IT cost center
    regex  = true
  }
}

# Find team labels with multiple possible values
data "kion_label" "tech_teams" {
  filter {
    name   = "key"
    values = ["Team"]
  }
  filter {
    name   = "value"
    values = ["Platform", "Security", "DevOps"]
  }
}

# Output the discovered labels
output "all_environment_labels" {
  value = {
    for label in data.kion_label.environment_labels.list :
    label.value => {
      id    = label.id
      color = label.color
    }
  }
  description = "Map of all environment labels"
}

output "production_label_details" {
  value = {
    for label in data.kion_label.production_label.list :
    "production" => {
      id    = label.id
      color = label.color
      key   = label.key
      value = label.value
    }
  }
  description = "Details of the production environment label"
}

output "it_cost_center_labels" {
  value = {
    for label in data.kion_label.it_cost_centers.list :
    label.value => {
      id    = label.id
      color = label.color
    }
  }
  description = "Map of IT cost center labels"
}

output "tech_team_labels" {
  value = {
    for label in data.kion_label.tech_teams.list :
    label.value => {
      id    = label.id
      color = label.color
    }
  }
  description = "Map of technical team labels"
}
