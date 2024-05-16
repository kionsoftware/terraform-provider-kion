# Create an OU.
resource "kion_ou" "ou1" {
  name                 = "sample-ou"
  description          = "Sample OU."
  parent_ou_id         = 0
  permission_scheme_id = 2
  owner_users { id = 1 }
  owner_user_groups { id = 1 }

  #labels = {
  #  (kion_label.env_staging.key) = kion_label.env_staging.value
  #  "Owner" = "jdoe"
  #}
}

# Output the ID of the resource created.
output "ou_id" {
  value = kion_ou.ou1.id
}
