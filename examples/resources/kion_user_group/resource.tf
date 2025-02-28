# Create DevOps team user group
resource "kion_user_group" "devops_team" {
  name        = "DevOps Team"
  description = "DevOps and Platform Engineering team members"
  idms_id     = 1  # Internal IDMS

  # Group owners
  owner_users {
    id = 10  # DevOps Manager
  }
  owner_users {
    id = 11  # Platform Lead
  }

  owner_groups {
    id = 5  # IT Management
  }

  # Group members
  users {
    id = 15  # DevOps Engineer
  }
  users {
    id = 16  # Platform Engineer
  }
  users {
    id = 17  # SRE
  }
}

# Create Security team user group
resource "kion_user_group" "security_team" {
  name        = "Security Team"
  description = "Security and compliance team members"
  idms_id     = 1

  # Group owners
  owner_users {
    id = 20  # Security Manager
  }

  owner_groups {
    id = 6  # Security Management
  }

  # Group members
  users {
    id = 25  # Security Engineer
  }
  users {
    id = 26  # Security Analyst
  }
}

# Create Cloud Operations team user group
resource "kion_user_group" "cloud_ops" {
  name        = "Cloud Operations"
  description = "Cloud infrastructure operations team"
  idms_id     = 1

  # Group owners
  owner_users {
    id = 30  # Operations Manager
  }
  owner_users {
    id = 31  # Cloud Architect
  }

  owner_groups {
    id = 7  # Operations Management
  }

  # Group members
  users {
    id = 35  # Cloud Engineer
  }
  users {
    id = 36  # Systems Engineer
  }
  users {
    id = 37  # Network Engineer
  }
}

# Output group information
output "team_group_ids" {
  value = {
    devops = {
      id         = kion_user_group.devops_team.id
      created_at = kion_user_group.devops_team.created_at
      enabled    = kion_user_group.devops_team.enabled
    }
    security = {
      id         = kion_user_group.security_team.id
      created_at = kion_user_group.security_team.created_at
      enabled    = kion_user_group.security_team.enabled
    }
    cloud_ops = {
      id         = kion_user_group.cloud_ops.id
      created_at = kion_user_group.cloud_ops.created_at
      enabled    = kion_user_group.cloud_ops.enabled
    }
  }
  description = "Details of created team groups"
}
