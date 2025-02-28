# Create a complete compliance check with all available options
resource "kion_compliance_check" "complete_example" {
  name                     = "Complete Compliance Check Example"
  description             = "A comprehensive example of a compliance check with all supported options."
  cloud_provider_id        = 1
  compliance_check_type_id = 1

  # Check configuration
  body = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "*"
        Resource = "*"
      }
    ]
  })

  # Azure-specific configuration
  azure_policy_id = 1

  # Frequency settings
  frequency_minutes  = 60
  frequency_type_id = 1

  # Region settings
  is_all_regions = false
  regions        = ["us-east-1", "us-west-2"]

  # Additional settings
  is_auto_archived    = false
  severity_type_id    = 1
  created_by_user_id  = 1

  # Ownership
  owner_users {
    id = 1
  }

  owner_user_groups {
    id = 2
  }
}

# Create an AWS-specific compliance check
resource "kion_compliance_check" "aws_example" {
  name                     = "AWS S3 Bucket Encryption"
  description             = "Checks if S3 buckets have encryption enabled"
  cloud_provider_id        = 1  # AWS
  compliance_check_type_id = 1

  body = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Deny"
        Action   = "s3:CreateBucket"
        Resource = "*"
        Condition = {
          "StringNotEquals": {
            "s3:x-amz-server-side-encryption": "AES256"
          }
        }
      }
    ]
  })

  frequency_minutes = 30
  is_all_regions   = true
  owner_users { id = 1 }
}

# Create an Azure-specific compliance check
resource "kion_compliance_check" "azure_example" {
  name                     = "Azure Storage Account Encryption"
  description             = "Ensures Azure Storage accounts are encrypted"
  cloud_provider_id        = 2  # Azure
  compliance_check_type_id = 2
  azure_policy_id         = 1

  frequency_minutes = 60
  is_all_regions   = true
  owner_user_groups { id = 1 }
}

# Output examples
output "complete_check_id" {
  value = kion_compliance_check.complete_example.id
}

output "aws_check_id" {
  value = kion_compliance_check.aws_example.id
}

output "azure_check_id" {
  value = kion_compliance_check.azure_example.id
}

output "last_scan_id" {
  value = kion_compliance_check.complete_example.last_scan_id
}
