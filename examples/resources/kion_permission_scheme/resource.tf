resource "kion_permission_scheme" "example" {
  # Required
  name = "example"
  type = "example"

  # Optional
  # roles = [{
  #   permission_id = 1
  #   role_id       = 1
  # }]
}
