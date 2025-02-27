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
