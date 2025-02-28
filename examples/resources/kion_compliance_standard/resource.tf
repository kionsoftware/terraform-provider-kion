# Create a complete compliance standard with all available options
resource "kion_compliance_standard" "complete_example" {
  name               = "Complete Compliance Standard Example"
  description        = "A comprehensive example of a compliance standard with all supported options."
  created_by_user_id = 1

  # Associated compliance checks
  compliance_checks {
    id = 1
  }
  compliance_checks {
    id = 2
  }

  # Ownership
  owner_users {
    id = 1
  }

  owner_user_groups {
    id = 2
  }
}

# Create a simple compliance standard for AWS security
resource "kion_compliance_standard" "aws_security" {
  name               = "AWS Security Standard"
  description        = "Basic AWS security compliance standard"
  created_by_user_id = 1

  compliance_checks {
    id = 3  # AWS S3 Encryption Check
  }
  compliance_checks {
    id = 4  # AWS KMS Key Rotation Check
  }

  owner_users {
    id = 1
  }
}

# Create a compliance standard for regulatory compliance
resource "kion_compliance_standard" "regulatory" {
  name               = "Regulatory Compliance"
  description        = "Standard for meeting regulatory requirements"
  created_by_user_id = 1

  compliance_checks {
    id = 5  # Data Privacy Check
  }
  compliance_checks {
    id = 6  # Access Control Check
  }
  compliance_checks {
    id = 7  # Audit Logging Check
  }

  owner_user_groups {
    id = 3  # Compliance Team
  }
}

# Output examples
output "complete_standard_id" {
  value = kion_compliance_standard.complete_example.id
}

output "aws_standard_id" {
  value = kion_compliance_standard.aws_security.id
}

output "regulatory_standard_id" {
  value = kion_compliance_standard.regulatory.id
}

output "creation_time" {
  value = kion_compliance_standard.complete_example.created_at
}
