# Retrieve an existing project note by its ID
data "kion_project_note" "documentation" {
  id         = "4" # The ID of the note to retrieve
  project_id = 11  # The project ID that contains the note
}

# Example of using the data source attributes
output "note_details" {
  value = {
    name             = data.kion_project_note.documentation.name
    text             = data.kion_project_note.documentation.text
    create_user_name = data.kion_project_note.documentation.create_user_name
  }
}

# Find all notes for a specific project
data "kion_project_note" "project_notes" {
  project_id = 10  # Development project
}

# Search notes by name pattern
data "kion_project_note" "runbook_notes" {
  query = "Runbook"

  filter {
    name   = "name"
    values = [".*Runbook.*", ".*Procedure.*"]
    regex  = true
  }
}

# Find notes by specific creator
data "kion_project_note" "devops_notes" {
  filter {
    name   = "create_user_id"
    values = ["15"]  # DevOps Lead
  }
}

# Find notes with specific content
data "kion_project_note" "security_notes" {
  filter {
    name   = "text"
    values = [".*security.*", ".*compliance.*"]
    regex  = true
  }
}

# Find recently updated notes
data "kion_project_note" "recent_notes" {
  filter {
    name   = "last_update_user_id"
    values = ["15", "16", "17"]  # Recent updaters
  }
}

# Output note information
output "project_note_summary" {
  value = {
    for note in data.kion_project_note.project_notes.list :
    note.name => {
      id              = note.id
      creator         = note.create_user_name
      last_updater    = note.last_update_user_name
    }
  }
  description = "Summary of all notes in the project"
}

output "runbook_details" {
  value = {
    for note in data.kion_project_note.runbook_notes.list :
    note.name => {
      id          = note.id
      creator     = note.create_user_name
      text_sample = substr(note.text, 0, 100)  # First 100 characters
    }
  }
  description = "Details of runbook notes"
}

output "devops_documentation" {
  value = {
    for note in data.kion_project_note.devops_notes.list :
    note.name => {
      id       = note.id
      project  = note.project_id
      text     = note.text
    }
  }
  description = "Documentation created by DevOps team"
}

output "security_documentation" {
  value = {
    for note in data.kion_project_note.security_notes.list :
    note.name => {
      id              = note.id
      creator         = note.create_user_name
      last_update_id  = note.last_update_user_id
    }
  }
  description = "Security-related documentation"
}

output "recent_updates" {
  value = {
    for note in data.kion_project_note.recent_notes.list :
    note.name => {
      id           = note.id
      last_updater = note.last_update_user_name
    }
  }
  description = "Recently updated notes"
}
