# Create a compliance standard.
resource "kion_compliance_standard" "s1" {
  name               = "sample-resource"
  created_by_user_id = 1
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
}

# Output the ID of the resource created.
output "standard_id" {
  value = kion_compliance_standard.s1.id
}
