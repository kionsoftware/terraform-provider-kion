# Get all compliance standards
data "kion_compliance_standard" "all" {
}

# Filter compliance standards by name
data "kion_compliance_standard" "by_name" {
  filter {
    name   = "name"
    values = ["AWS Security Standard"]
  }
}

# Filter by description with regex
data "kion_compliance_standard" "security_standards" {
  filter {
    name   = "description"
    values = [".*security.*"]
    regex  = true
  }
}

# Filter by creation date
data "kion_compliance_standard" "recent_standards" {
  filter {
    name   = "created_at"
    values = ["2024-01-01"]  # Standards created after this date
    regex  = true
  }
}

# Filter by multiple criteria
data "kion_compliance_standard" "managed_standards" {
  filter {
    name   = "ct_managed"
    values = ["true"]
  }
  filter {
    name   = "created_by_user_id"
    values = ["1"]
  }
}

# Example outputs
output "all_standards" {
  value = data.kion_compliance_standard.all.list
}

output "aws_standard_details" {
  value = data.kion_compliance_standard.by_name.list[0]
}

output "security_standard_names" {
  value = [for std in data.kion_compliance_standard.security_standards.list : std.name]
}

output "recent_standard_count" {
  value = length(data.kion_compliance_standard.recent_standards.list)
}

output "managed_standards" {
  value = [
    for std in data.kion_compliance_standard.managed_standards.list : {
      name        = std.name
      description = std.description
      created_at  = std.created_at
    }
  ]
}