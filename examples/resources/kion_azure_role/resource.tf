# Create an Azure Role Definition.
resource "kion_azure_role" "ar1" {
  name             = "Test Role"
  description      = "A test role created by our Terraform provider."
  role_permissions = <<EOF
{
    "actions": [
        "Microsoft.Authorization/roleDefinitions/read"
    ],
    "notActions": [],
    "dataActions": [],
    "notDataActions": []
}
EOF
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
}

# Output the ID of the resource created.
output "ar_id" {
  value = kion_azure_role.ar1.id
}
