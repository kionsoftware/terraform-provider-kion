# Create an Azure ARM template.
resource "kion_azure_arm_template" "arm1" {
  name                     = "tf test"
  description              = "A test Azure ARM template created via our Terraform provider."
  resource_group_name      = "3797RGI"
  resource_group_region_id = 41

  # Valid values are either 1 ("incremental") or 2 ("complete")
  # deployment_mode = 1
  deployment_mode = 2
  owner_users { id = 1 }

  template = <<EOF
{
    "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
    "contentVersion": "1.0.0.0",
    "parameters": {},
    "variables": {},
    "resources": [],
    "outputs": {}
}
EOF
}

# Output the ID of the resource created.
output "arm_template_id" {
  value = kion_azure_arm_template.arm1.id
}
