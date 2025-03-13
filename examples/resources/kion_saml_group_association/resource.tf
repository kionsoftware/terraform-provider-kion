# Create SAML group associations for different team roles
# DevOps team association
resource "kion_saml_group_association" "devops_mapping" {
  assertion_name   = "groups"
  assertion_regex  = "^devops-team"
  idms_id         = 5  # SAML IDMS ID
  update_on_login = true
  user_group_id   = 10  # DevOps team group ID
}

# Security team association with multiple group patterns
resource "kion_saml_group_association" "security_mapping" {
  assertion_name   = "memberOf"
  assertion_regex  = "^(security-team|compliance-team)"
  idms_id         = 5
  update_on_login = true
  user_group_id   = 20  # Security team group ID
}

# Cloud platform team association
resource "kion_saml_group_association" "platform_mapping" {
  assertion_name   = "groups"
  assertion_regex  = "^cloud-platform-.*"  # Matches any cloud platform group
  idms_id         = 5
  update_on_login = true
  user_group_id   = 30  # Platform team group ID
}

# Administrator group association
resource "kion_saml_group_association" "admin_mapping" {
  assertion_name   = "memberOf"
  assertion_regex  = "^kion-administrators$"  # Exact match for admin group
  idms_id         = 5
  update_on_login = true
  user_group_id   = 1  # Administrators group ID
}

# Output association information
output "saml_associations" {
  value = {
    devops = {
      id              = kion_saml_group_association.devops_mapping.id
      idms_saml_id    = kion_saml_group_association.devops_mapping.idms_saml_id
      update_on_login = kion_saml_group_association.devops_mapping.should_update_on_login
    }
    security = {
      id              = kion_saml_group_association.security_mapping.id
      idms_saml_id    = kion_saml_group_association.security_mapping.idms_saml_id
      update_on_login = kion_saml_group_association.security_mapping.should_update_on_login
    }
    platform = {
      id              = kion_saml_group_association.platform_mapping.id
      idms_saml_id    = kion_saml_group_association.platform_mapping.idms_saml_id
      update_on_login = kion_saml_group_association.platform_mapping.should_update_on_login
    }
    admin = {
      id              = kion_saml_group_association.admin_mapping.id
      idms_saml_id    = kion_saml_group_association.admin_mapping.idms_saml_id
      update_on_login = kion_saml_group_association.admin_mapping.should_update_on_login
    }
  }
  description = "Details of SAML group associations"
}
