# terraform-kion-app-role

Terraform module for `kion_app_role`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "app_role" {
  source = "..."

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
| [kion_app_role.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/app_role) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the app role. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| disabled | disabled of the kion\_app\_role. |
| id | id of the kion\_app\_role. |
| system\_managed | system\_managed of the kion\_app\_role. |
<!-- END_TF_DOCS -->
