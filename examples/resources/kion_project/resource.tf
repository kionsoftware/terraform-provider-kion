resource "kion_project" "p1" {
  ou_id                = 1
  name                 = "Tech Project I"
  description          = "This is a sample project."
  permission_scheme_id = 3
  owner_user_ids { id = 1 }
  project_funding {
    amount            = 1000
    funding_order     = 1
    funding_source_id = 1
    start_datecode    = "2021-01"
    end_datecode      = "2022-01"
  }

  #labels = {
  #  (kion_label.env_staging.key) = kion_label.env_staging.value
  #  "Owner" = "jdoe"
  #}
}

# Output the ID of the resource created.
output "project_id" {
  value = kion_project.p1.id
}
