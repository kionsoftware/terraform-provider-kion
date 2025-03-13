# Get all compliance checks
data "kion_compliance_check" "all" {
}

# Filter compliance checks by name
data "kion_compliance_check" "by_name" {
  filter {
    name   = "name"
    values = ["S3 Bucket Encryption"]
  }
}

# Filter by cloud provider
data "kion_compliance_check" "aws_checks" {
  filter {
    name   = "cloud_provider_id"
    values = ["1"]  # AWS
  }
}

# Filter by multiple criteria
data "kion_compliance_check" "critical_aws_checks" {
  filter {
    name   = "cloud_provider_id"
    values = ["1"]  # AWS
  }
  filter {
    name   = "severity_type_id"
    values = ["1"]  # Critical
  }
}

# Filter by region and auto-archive settings
data "kion_compliance_check" "regional_checks" {
  filter {
    name   = "is_all_regions"
    values = ["false"]
  }
  filter {
    name   = "is_auto_archived"
    values = ["false"]
  }
}

# Example outputs
output "all_checks" {
  value = data.kion_compliance_check.all.list
}

output "encryption_check_id" {
  value = data.kion_compliance_check.by_name.list[0].id
}

output "aws_check_count" {
  value = length(data.kion_compliance_check.aws_checks.list)
}

output "critical_aws_checks" {
  value = [for check in data.kion_compliance_check.critical_aws_checks.list : check.name]
}

output "regional_check_details" {
  value = [
    for check in data.kion_compliance_check.regional_checks.list : {
      name    = check.name
      regions = check.regions
    }
  ]
}