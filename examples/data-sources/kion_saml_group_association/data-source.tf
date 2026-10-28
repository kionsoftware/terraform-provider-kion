# Query all group associations for a SAML IDMS
data "kion_saml_group_association" "all_associations" {
  idms_id = 25
}

# Query SAML associations for DevOps teams
data "kion_saml_group_association" "devops_associations" {
  idms_id = 25

  filter {
    name   = "assertion_regex"
    values = [".*devops.*"]
    regex  = true
  }
}

# Query associations by specific assertion name
data "kion_saml_group_association" "group_assertions" {
  idms_id = 25

  filter {
    name   = "assertion_name"
    values = ["groups"]
  }
}

# Query associations with multiple filters
data "kion_saml_group_association" "admin_associations" {
  idms_id = 25

  filter {
    name   = "assertion_name"
    values = ["memberOf"]
  }
  filter {
    name   = "assertion_regex"
    values = [".*admin.*"]
    regex  = true
  }
}

# Output SAML group association information
output "saml_association_summary" {
  value = {
    total_associations = length(data.kion_saml_group_association.all_associations.list)
    devops_count       = length(data.kion_saml_group_association.devops_associations.list)
    group_assertions = {
      count = length(data.kion_saml_group_association.group_assertions.list)
      names = [
        for assoc in data.kion_saml_group_association.group_assertions.list :
        assoc.assertion_name
      ]
    }
    admin_mappings = {
      count = length(data.kion_saml_group_association.admin_associations.list)
      details = [
        for assoc in data.kion_saml_group_association.admin_associations.list :
        {
          id             = assoc.id
          assertion_name = assoc.assertion_name
          regex          = assoc.assertion_regex
        }
      ]
    }
  }
  description = "Summary of SAML group associations including counts and details for different filters"
}
