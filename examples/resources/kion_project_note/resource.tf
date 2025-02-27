# Create a new project note
resource "kion_project_note" "documentation" {
  name           = "Project Configuration"
  project_id     = 11
  create_user_id = 1
  text           = <<-EOT
    # Project Overview
    This project is configured for development and testing purposes.

    ## Important Contacts
    - Project Owner: John Doe
    - Technical Lead: Jane Smith

    ## Key Dates
    - Project Start: 2024-01-01
    - Next Review: 2024-06-01

    ## Configuration Details
    - Environment: Development
    - Region: us-east-1
    - VPC ID: vpc-12345678
  EOT
}

# Example of referencing the note's attributes
output "note_info" {
  value = {
    id               = kion_project_note.documentation.id
    name             = kion_project_note.documentation.name
    create_user_name = kion_project_note.documentation.create_user_name
    last_update_user = kion_project_note.documentation.last_update_user_name
  }
}
