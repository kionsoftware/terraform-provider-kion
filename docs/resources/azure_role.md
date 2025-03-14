---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kion_azure_role Resource - terraform-provider-kion"
subcategory: ""
description: |-
  
---

# kion_azure_role (Resource)



## Example Usage

```terraform
# Example 1: Read-only role for resource groups
resource "kion_azure_role" "resource_viewer" {
  name        = "Resource Group Viewer"
  description = "Allows read-only access to resource groups and their resources"

  role_permissions = <<EOF
{
    "assignableScopes": [
        "/"
    ],
    "description": "View resource groups and their contents",
    "permissions": [
        {
            "actions": [
                "Microsoft.Resources/subscriptions/resourceGroups/read",
                "Microsoft.Resources/subscriptions/resourceGroups/resources/read"
            ],
            "notActions": [],
            "dataActions": [],
            "notDataActions": []
        }
    ]
}
EOF

  owner_users { id = 1 }
}

# Example 2: Storage Account Administrator
resource "kion_azure_role" "storage_admin" {
  name        = "Storage Account Administrator"
  description = "Full access to manage storage accounts"

  role_permissions = <<EOF
{
    "assignableScopes": [
        "/"
    ],
    "description": "Manage storage accounts and their contents",
    "permissions": [
        {
            "actions": [
                "Microsoft.Storage/storageAccounts/*",
                "Microsoft.Resources/subscriptions/resourceGroups/read",
                "Microsoft.Resources/subscriptions/resourceGroups/resources/read"
            ],
            "notActions": [
                "Microsoft.Storage/storageAccounts/delete"
            ],
            "dataActions": [
                "Microsoft.Storage/storageAccounts/blobServices/containers/*",
                "Microsoft.Storage/storageAccounts/fileServices/shares/*"
            ],
            "notDataActions": []
        }
    ]
}
EOF

  owner_users { id = 1 }
  owner_user_groups { id = 2 }
}

# Example 3: Network Security Administrator
resource "kion_azure_role" "network_admin" {
  name        = "Network Security Administrator"
  description = "Manage network security groups and firewall rules"

  role_permissions = <<EOF
{
    "assignableScopes": [
        "/"
    ],
    "description": "Manage network security configurations",
    "permissions": [
        {
            "actions": [
                "Microsoft.Network/networkSecurityGroups/*",
                "Microsoft.Network/virtualNetworks/subnets/join/action",
                "Microsoft.Network/virtualNetworks/read",
                "Microsoft.Network/publicIPAddresses/*"
            ],
            "notActions": [],
            "dataActions": [],
            "notDataActions": []
        }
    ]
}
EOF

  owner_user_groups { id = 3 }
}

# Example 4: Web App Developer
resource "kion_azure_role" "webapp_developer" {
  name        = "Web App Developer"
  description = "Manage web apps and app service plans"

  role_permissions = <<EOF
{
    "assignableScopes": [
        "/"
    ],
    "description": "Deploy and manage web applications",
    "permissions": [
        {
            "actions": [
                "Microsoft.Web/sites/*",
                "Microsoft.Web/serverfarms/*",
                "Microsoft.Resources/deployments/*",
                "Microsoft.Insights/components/*"
            ],
            "notActions": [
                "Microsoft.Web/sites/delete",
                "Microsoft.Web/serverfarms/delete"
            ],
            "dataActions": [],
            "notDataActions": []
        }
    ]
}
EOF

  owner_users { id = 1 }
}

# Example 5: Database Contributor
resource "kion_azure_role" "db_contributor" {
  name        = "Database Contributor"
  description = "Manage SQL databases but not security-related policies"

  role_permissions = <<EOF
{
    "assignableScopes": [
        "/"
    ],
    "description": "Manage SQL databases with restricted permissions",
    "permissions": [
        {
            "actions": [
                "Microsoft.Sql/servers/databases/*",
                "Microsoft.Sql/servers/read",
                "Microsoft.Resources/subscriptions/resourceGroups/read"
            ],
            "notActions": [
                "Microsoft.Sql/servers/databases/delete",
                "Microsoft.Sql/servers/databases/threatProtectionSettings/*",
                "Microsoft.Sql/servers/databases/transparentDataEncryption/*",
                "Microsoft.Sql/servers/databases/securityAlertPolicies/*",
                "Microsoft.Sql/servers/databases/auditingSettings/*"
            ],
            "dataActions": [
                "Microsoft.Sql/servers/databases/connect/action",
                "Microsoft.Sql/servers/databases/read",
                "Microsoft.Sql/servers/databases/write"
            ],
            "notDataActions": []
        }
    ]
}
EOF

  owner_users { id = 1 }
  owner_user_groups { id = 4 }
}

# Output examples
output "viewer_role_id" {
  description = "ID of the resource viewer role"
  value       = kion_azure_role.resource_viewer.id
}

output "storage_role_id" {
  description = "ID of the storage administrator role"
  value       = kion_azure_role.storage_admin.id
}

output "network_role_managed" {
  description = "Whether the network admin role is Azure managed"
  value       = kion_azure_role.network_admin.azure_managed_policy
}

output "webapp_role_system_managed" {
  description = "Whether the web app developer role is system managed"
  value       = kion_azure_role.webapp_developer.system_managed_policy
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String)
- `role_permissions` (String)

### Optional

- `description` (String)
- `last_updated` (String)
- `owner_user_groups` (Block Set) Must provide at least the owner_user_groups field or the owner_users field. (see [below for nested schema](#nestedblock--owner_user_groups))
- `owner_users` (Block Set) Must provide at least the owner_user_groups field or the owner_users field. (see [below for nested schema](#nestedblock--owner_users))

### Read-Only

- `azure_managed_policy` (Boolean)
- `id` (String) The ID of this resource.
- `system_managed_policy` (Boolean)

<a id="nestedblock--owner_user_groups"></a>
### Nested Schema for `owner_user_groups`

Optional:

- `id` (Number)


<a id="nestedblock--owner_users"></a>
### Nested Schema for `owner_users`

Optional:

- `id` (Number)
