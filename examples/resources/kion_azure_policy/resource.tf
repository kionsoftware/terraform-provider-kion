# Example 1: Require tag on resources
resource "kion_azure_policy" "require_tags" {
  name        = "Require Environment Tag"
  description = "Requires resources to have an environment tag"

  policy = <<EOF
{
    "if": {
        "allOf": [
            {
                "field": "type",
                "equals": "Microsoft.Resources/subscriptions/resourceGroups"
            },
            {
                "field": "tags['environment']",
                "exists": "false"
            }
        ]
    },
    "then": {
        "effect": "deny"
    }
}
EOF

  parameters = <<EOF
{
    "tagName": {
        "type": "String",
        "metadata": {
            "displayName": "Tag Name",
            "description": "Name of the tag to enforce"
        },
        "defaultValue": "environment"
    }
}
EOF

  owner_users { id = 1 }
}

# Example 2: Allowed VM SKUs
resource "kion_azure_policy" "allowed_vms" {
  name        = "Allowed VM SKUs"
  description = "Restricts VM deployments to specific SKUs"

  policy = <<EOF
{
    "if": {
        "allOf": [
            {
                "field": "type",
                "equals": "Microsoft.Compute/virtualMachines"
            },
            {
                "not": {
                    "field": "Microsoft.Compute/virtualMachines/sku.name",
                    "in": "[parameters('allowedSkus')]"
                }
            }
        ]
    },
    "then": {
        "effect": "deny"
    }
}
EOF

  parameters = <<EOF
{
    "allowedSkus": {
        "type": "Array",
        "metadata": {
            "displayName": "Allowed VM SKUs",
            "description": "List of allowed VM SKUs"
        },
        "defaultValue": [
            "Standard_D2s_v3",
            "Standard_D4s_v3",
            "Standard_D8s_v3"
        ]
    }
}
EOF

  owner_user_groups { id = 2 }
}

# Example 3: Enforce Storage Account Encryption
resource "kion_azure_policy" "storage_encryption" {
  name        = "Storage Encryption Requirements"
  description = "Enforces encryption settings on storage accounts"

  policy = <<EOF
{
    "if": {
        "allOf": [
            {
                "field": "type",
                "equals": "Microsoft.Storage/storageAccounts"
            },
            {
                "not": {
                    "allOf": [
                        {
                            "field": "Microsoft.Storage/storageAccounts/supportsHttpsTrafficOnly",
                            "equals": "true"
                        },
                        {
                            "field": "Microsoft.Storage/storageAccounts/minimumTlsVersion",
                            "equals": "TLS1_2"
                        },
                        {
                            "field": "Microsoft.Storage/storageAccounts/encryption.services.blob.enabled",
                            "equals": "true"
                        }
                    ]
                }
            }
        ]
    },
    "then": {
        "effect": "deny"
    }
}
EOF

  owner_users { id = 1 }
  owner_user_groups { id = 3 }
}

# Example 4: Network Security Group Rules
resource "kion_azure_policy" "nsg_rules" {
  name        = "NSG Security Requirements"
  description = "Enforces security rules on Network Security Groups"

  policy = <<EOF
{
    "if": {
        "allOf": [
            {
                "field": "type",
                "equals": "Microsoft.Network/networkSecurityGroups/securityRules"
            },
            {
                "anyOf": [
                    {
                        "allOf": [
                            {
                                "field": "Microsoft.Network/networkSecurityGroups/securityRules/access",
                                "equals": "Allow"
                            },
                            {
                                "field": "Microsoft.Network/networkSecurityGroups/securityRules/direction",
                                "equals": "Inbound"
                            },
                            {
                                "field": "Microsoft.Network/networkSecurityGroups/securityRules/sourceAddressPrefix",
                                "equals": "*"
                            }
                        ]
                    },
                    {
                        "field": "Microsoft.Network/networkSecurityGroups/securityRules/destinationPortRange",
                        "in": "[parameters('restrictedPorts')]"
                    }
                ]
            }
        ]
    },
    "then": {
        "effect": "deny"
    }
}
EOF

  parameters = <<EOF
{
    "restrictedPorts": {
        "type": "Array",
        "metadata": {
            "displayName": "Restricted Ports",
            "description": "Ports that should not be exposed"
        },
        "defaultValue": [
            "22",
            "3389",
            "161",
            "162"
        ]
    }
}
EOF

  owner_users { id = 1 }
}

# Output examples
output "tag_policy_id" {
  description = "ID of the tag requirement policy"
  value       = kion_azure_policy.require_tags.id
}

output "vm_policy_id" {
  description = "ID of the VM SKU policy"
  value       = kion_azure_policy.allowed_vms.id
}

output "storage_policy_managed" {
  description = "Whether the storage policy is CT managed"
  value       = kion_azure_policy.storage_encryption.ct_managed
}

output "nsg_policy_definition" {
  description = "Azure managed policy definition ID for NSG rules"
  value       = kion_azure_policy.nsg_rules.azure_managed_policy_def_id
}