# terraform-kion-project-line-item

Terraform module for `kion_project_line_item`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "project_line_item" {
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
| [kion_project_line_item.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/project_line_item) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| amount | Amount for the project line item in the application. | `number` | `null` | no |
| category\_id | Category ID of related Category of the Project Line Item in the application. | `number` | `null` | no |
| datecode | Datecode for the project line item in the application. | `number` | `null` | no |
| description | Description for the project line item in the application. | `string` | `null` | no |
| funding\_source\_id | Funding Source ID of related Funding Source of the Project Line Item in the application. | `number` | `null` | no |
| payer\_id | Payer ID of related Payer of the Project Line Item in the application. | `number` | `null` | no |
| project\_id | Project ID of the Project Line Item in the application. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_project\_line\_item. |
<!-- END_TF_DOCS -->
