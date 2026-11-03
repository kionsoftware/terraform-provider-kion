# terraform-kion-category

Terraform module for `kion_category`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "category" {
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
| [kion_category.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/category) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Category name. | `string` | n/a | yes |
| description | Category description. | `string` | `null` | no |
| payer\_id | Payer ID of related Payer of the category in the application. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_category. |
<!-- END_TF_DOCS -->
