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

# Example 1: Basic Storage Account Template
resource "kion_azure_arm_template" "storage_example" {
  name                     = "Basic Storage Account"
  description              = "Creates a basic Azure storage account with secure configuration"
  resource_group_name      = "storage-rg"
  resource_group_region_id = 41  # Azure East US
  deployment_mode          = 1   # Incremental deployment

  owner_users { id = 1 }

  template = <<EOF
{
    "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
    "contentVersion": "1.0.0.0",
    "parameters": {
        "storageAccountName": {
            "type": "string",
            "defaultValue": "storageaccount1"
        },
        "storageSKU": {
            "type": "string",
            "defaultValue": "Standard_LRS",
            "allowedValues": [
                "Standard_LRS",
                "Standard_GRS",
                "Standard_ZRS"
            ]
        }
    },
    "resources": [
        {
            "type": "Microsoft.Storage/storageAccounts",
            "apiVersion": "2021-04-01",
            "name": "[parameters('storageAccountName')]",
            "location": "[resourceGroup().location]",
            "sku": {
                "name": "[parameters('storageSKU')]"
            },
            "kind": "StorageV2",
            "properties": {
                "supportsHttpsTrafficOnly": true,
                "minimumTlsVersion": "TLS1_2",
                "allowBlobPublicAccess": false
            }
        }
    ]
}
EOF

  template_parameters = <<EOF
{
    "storageAccountName": {
        "value": "securestorage001"
    },
    "storageSKU": {
        "value": "Standard_GRS"
    }
}
EOF
}

# Example 2: Complete Virtual Network Template
resource "kion_azure_arm_template" "vnet_example" {
  name                     = "Production VNet"
  description              = "Creates a complete virtual network with subnets and security rules"
  resource_group_name      = "network-rg"
  resource_group_region_id = 41
  deployment_mode          = 2  # Complete deployment

  owner_user_groups { id = 2 }
  owner_users { id = 1 }

  template = <<EOF
{
    "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
    "contentVersion": "1.0.0.0",
    "parameters": {
        "vnetName": {
            "type": "string",
            "defaultValue": "MainVNet"
        },
        "vnetAddressPrefix": {
            "type": "string",
            "defaultValue": "10.0.0.0/16"
        },
        "subnet1Prefix": {
            "type": "string",
            "defaultValue": "10.0.1.0/24"
        },
        "subnet2Prefix": {
            "type": "string",
            "defaultValue": "10.0.2.0/24"
        }
    },
    "resources": [
        {
            "type": "Microsoft.Network/virtualNetworks",
            "apiVersion": "2021-02-01",
            "name": "[parameters('vnetName')]",
            "location": "[resourceGroup().location]",
            "properties": {
                "addressSpace": {
                    "addressPrefixes": [
                        "[parameters('vnetAddressPrefix')]"
                    ]
                },
                "subnets": [
                    {
                        "name": "subnet1",
                        "properties": {
                            "addressPrefix": "[parameters('subnet1Prefix')]",
                            "networkSecurityGroup": {
                                "properties": {
                                    "securityRules": [
                                        {
                                            "name": "DenyAllInbound",
                                            "properties": {
                                                "priority": 4096,
                                                "direction": "Inbound",
                                                "access": "Deny",
                                                "protocol": "*",
                                                "sourceAddressPrefix": "*",
                                                "sourcePortRange": "*",
                                                "destinationAddressPrefix": "*",
                                                "destinationPortRange": "*"
                                            }
                                        }
                                    ]
                                }
                            }
                        }
                    },
                    {
                        "name": "subnet2",
                        "properties": {
                            "addressPrefix": "[parameters('subnet2Prefix')]"
                        }
                    }
                ]
            }
        }
    ]
}
EOF

  template_parameters = <<EOF
{
    "vnetName": {
        "value": "prod-vnet-001"
    },
    "vnetAddressPrefix": {
        "value": "172.16.0.0/16"
    },
    "subnet1Prefix": {
        "value": "172.16.1.0/24"
    },
    "subnet2Prefix": {
        "value": "172.16.2.0/24"
    }
}
EOF
}

# Example 3: Web App with Database Template
resource "kion_azure_arm_template" "webapp_example" {
  name                     = "Web App with SQL"
  description              = "Deploys a web app with SQL database and application insights"
  resource_group_name      = "webapp-rg"
  resource_group_region_id = 41
  deployment_mode          = 1

  owner_users { id = 1 }

  template = <<EOF
{
    "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
    "contentVersion": "1.0.0.0",
    "parameters": {
        "webAppName": {
            "type": "string",
            "defaultValue": "webapp1"
        },
        "sqlServerName": {
            "type": "string",
            "defaultValue": "sqlserver1"
        },
        "databaseName": {
            "type": "string",
            "defaultValue": "database1"
        },
        "administratorLogin": {
            "type": "string"
        },
        "administratorLoginPassword": {
            "type": "securestring"
        }
    },
    "resources": [
        {
            "type": "Microsoft.Web/serverfarms",
            "apiVersion": "2021-02-01",
            "name": "[concat(parameters('webAppName'), '-plan')]",
            "location": "[resourceGroup().location]",
            "sku": {
                "name": "S1",
                "tier": "Standard"
            }
        },
        {
            "type": "Microsoft.Web/sites",
            "apiVersion": "2021-02-01",
            "name": "[parameters('webAppName')]",
            "location": "[resourceGroup().location]",
            "dependsOn": [
                "[concat(parameters('webAppName'), '-plan')]"
            ],
            "properties": {
                "serverFarmId": "[resourceId('Microsoft.Web/serverfarms', concat(parameters('webAppName'), '-plan'))]",
                "httpsOnly": true,
                "siteConfig": {
                    "minTlsVersion": "1.2"
                }
            }
        },
        {
            "type": "Microsoft.Sql/servers",
            "apiVersion": "2021-02-01-preview",
            "name": "[parameters('sqlServerName')]",
            "location": "[resourceGroup().location]",
            "properties": {
                "administratorLogin": "[parameters('administratorLogin')]",
                "administratorLoginPassword": "[parameters('administratorLoginPassword')]",
                "version": "12.0"
            }
        },
        {
            "type": "Microsoft.Sql/servers/databases",
            "apiVersion": "2021-02-01-preview",
            "name": "[concat(parameters('sqlServerName'), '/', parameters('databaseName'))]",
            "location": "[resourceGroup().location]",
            "dependsOn": [
                "[parameters('sqlServerName')]"
            ],
            "sku": {
                "name": "Basic",
                "tier": "Basic"
            }
        }
    ]
}
EOF

  template_parameters = <<EOF
{
    "webAppName": {
        "value": "mywebapp-prod-001"
    },
    "sqlServerName": {
        "value": "mysqlserver-prod-001"
    },
    "databaseName": {
        "value": "mydb-prod"
    },
    "administratorLogin": {
        "value": "dbadmin"
    }
}
EOF
}

# Output examples
output "storage_template_id" {
  description = "ID of the storage account template"
  value       = kion_azure_arm_template.storage_example.id
}

output "vnet_template_id" {
  description = "ID of the virtual network template"
  value       = kion_azure_arm_template.vnet_example.id
}

output "webapp_template_version" {
  description = "Version of the web app template"
  value       = kion_azure_arm_template.webapp_example.version
}

output "template_managed_status" {
  description = "Whether the storage template is CT managed"
  value       = kion_azure_arm_template.storage_example.ct_managed
}
