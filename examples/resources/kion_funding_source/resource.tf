# Create and attach a funding source
resource "kion_funding_source" "fs1" {
  name                 = "Test Funding"
  description          = "Sample funding source created via terraform"
  amount               = 10000
  start_datecode       = "2023-01"
  end_datecode         = "2023-12"
  ou_id                = 1
  permission_scheme_id = 4
  owner_users { id = 1 }
  owner_user_groups { id = 3 }

  #labels = {
  #  (kion_label.env_staging.key) = kion_label.env_staging.value
  #  "Owner" = "jdoe"
  #}
}

# Output the ID of the created funding source.
output "fs_id" {
  value = kion_funding_source.fs1.id
}
