# terraform-kion-global-permission-mapping

Terraform module for `kion_global_permission_mapping`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "global_permission_mapping" {
  source = "..."

  app_role_id = 1
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
| [kion_global_permission_mapping.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/global_permission_mapping) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| app\_role\_id | The ID of the app role. | `number` | n/a | yes |
| user\_groups\_ids | The IDs of the user groups in the mapping. | `set(number)` | `null` | no |
| user\_ids | The IDs of the users in the mapping. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_global\_permission\_mapping. |
<!-- END_TF_DOCS -->
