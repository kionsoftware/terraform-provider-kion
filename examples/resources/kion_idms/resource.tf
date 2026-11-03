resource "kion_idms" "example" {
  # Required
  idms_type_id        = 1
  name                = "example"
  password_expiration = 1
}
