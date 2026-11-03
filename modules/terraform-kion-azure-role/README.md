# terraform-kion-azure-role

Terraform module for `kion_azure_role`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "azure_role" {
  source = "..."

  name             = "example"
  role_permissions = "{}"
}
```

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
| ---- | ------- |
| terraform | >= 1.0 |
| kion | >= 1.0.0 |

## Providers

| Name | Version |
| ---- | ------- |
| kion | >= 1.0.0 |

## Modules

No modules.

## Resources

| Name | Type |
| ---- | ---- |
| [kion_azure_role.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/azure_role) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the Azure Role in the application. | `string` | n/a | yes |
| role\_permissions | Text of the permissions section of the created Azure Role Definition. | `string` | n/a | yes |
| car\_restricted\_user\_group\_ids | List of groups IDs who have been allowed to use the Azure Role on Cloud Access Roles in the system. | `set(number)` | `null` | no |
| car\_restricted\_user\_ids | List of user IDs who have been allowed to use the Azure Role on Cloud Access Roles in the system. | `set(number)` | `null` | no |
| description | Description for the Azure Role in the application. | `string` | `null` | no |
| owner\_user\_group\_ids | List of group IDs who will own the Azure Role. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the Azure Role. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_azure\_role. |
<!-- END_TF_DOCS -->
