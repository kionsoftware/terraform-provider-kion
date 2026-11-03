# terraform-kion-app-api-key

Terraform module for `kion_app_api_key`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "app_api_key" {
  source = "..."

  name = "example"
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
| [kion_app_api_key.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/app_api_key) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the key in the application. | `string` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_app\_api\_key. |
<!-- END_TF_DOCS -->
