resource "kion_user_group" "ug1" {
  name        = "sample-user-group2"
  description = "This is a sample user group."
  idms_id     = 1
  owner_groups { id = 1 }
  owner_users { id = 1 }
  users { id = 1 }
}

# Output the ID of the resource created.
output "user_group_id" {
  value = kion_user_group.ug1.id
}
