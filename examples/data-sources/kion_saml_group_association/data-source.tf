# Query all SAML group associations
data "kion_saml_group_association" "all_associations" {
}

# Query SAML associations for DevOps teams
data "kion_saml_group_association" "devops_associations" {
  filter {
    name   = "assertion_regex"
    values = [".*devops.*"]
  }
}

# Query associations by specific assertion name
data "kion_saml_group_association" "group_assertions" {
  filter {
    name   = "assertion_name"
    values = ["groups"]
  }
}

# Query associations with multiple filters
data "kion_saml_group_association" "admin_associations" {
  filter {
    name   = "assertion_name"
    values = ["memberOf"]
  }
  filter {
    name   = "assertion_regex"
    values = [".*admin.*"]
  }
}

# Output SAML group association information
output "saml_association_summary" {
  value = {
    total_associations = length(data.kion_saml_group_association.all_associations.saml_group_associations)
    devops_count      = length(data.kion_saml_group_association.devops_associations.saml_group_associations)
    group_assertions  = {
      count = length(data.kion_saml_group_association.group_assertions.saml_group_associations)
      names = [
        for assoc in data.kion_saml_group_association.group_assertions.saml_group_associations :
        assoc.assertion_name
      ]
    }
    admin_mappings = {
      count = length(data.kion_saml_group_association.admin_associations.saml_group_associations)
      details = [
        for assoc in data.kion_saml_group_association.admin_associations.saml_group_associations :
        {
          id             = assoc.id
          assertion_name = assoc.assertion_name
          regex         = assoc.assertion_regex
        }
      ]
    }
  }
  description = "Summary of SAML group associations including counts and details for different filters"
}