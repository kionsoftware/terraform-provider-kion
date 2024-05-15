# Create a cloud rule.
resource "kion_cloud_rule" "cr1" {
  name        = "sample-resource"
  description = "Sample cloud rule."
  aws_iam_policies { id = 1 }
  owner_users { id = 1 }
  owner_user_groups { id = 1 }

  #labels = {
  #  (kion_label.env_staging.key) = kion_label.env_staging.value
  #  "Owner" = "jdoe"
  #}
}

# Output the ID of the resource created.
output "rule_id" {
  value = kion_cloud_rule.cr1.id
}
