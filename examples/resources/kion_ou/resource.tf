# Create an OU with multiple owners and labels
resource "kion_ou" "production_ou" {
  name                 = "production"
  description          = "Production environment organizational unit"
  parent_ou_id         = 1  # Root OU
  permission_scheme_id = 2  # Example permission scheme

  # Multiple owner users
  owner_users {
    id = 10  # DevOps Lead
  }
  owner_users {
    id = 11  # Security Lead
  }

  # Multiple owner groups
  owner_user_groups {
    id = 5  # DevOps Team
  }
  owner_user_groups {
    id = 6  # Security Team
  }

  # Labels for organization and tracking
  labels = {
    "Environment" = "Production"
    "CostCenter" = "IT-1234"
    "Team"       = "Platform"
  }
}

# Output important values
output "production_ou_id" {
  value       = kion_ou.production_ou.id
  description = "The ID of the production OU"
}

output "production_ou_created_at" {
  value       = kion_ou.production_ou.created_at
  description = "Timestamp when the OU was created"
}
