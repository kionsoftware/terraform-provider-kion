# Create an association between a User Group and a SAML IDMS.
resource "kion_saml_group_association" "sa1" {
  assertion_name  = "memberOf"
  assertion_regex = "^kion-admins"
  idms_id         = 5
  update_on_login = true
  user_group_id   = 1
}

# Output the ID of the resource created.
output "saml_group_assocation_id" {
  value = kion_saml_group_association.sa1.id
}
