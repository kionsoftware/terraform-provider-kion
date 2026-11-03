# terraform-kion-service-catalog

Terraform module for `kion_service_catalog`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "service_catalog" {
  source = "..."

  account_id   = 1
  name         = "example"
  portfolio_id = "example"
  region       = "example"
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
| [kion_service_catalog.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/service_catalog) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| account\_id | ID of account where the service catalog portfolio exists. | `number` | n/a | yes |
| name | Name of the service catalog portfolio in the application. | `string` | n/a | yes |
| portfolio\_id | ID of the service catalog portfolio in AWS. | `string` | n/a | yes |
| region | AWS region where the service catalog portfolio exists. | `string` | n/a | yes |
| description | Description for the service catalog portfolio in the application. | `string` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the portfolio. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the portfolio. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |
| tag\_option | Boolean that enables or disables tag option sharing on service catalog portfolios. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_service\_catalog. |
<!-- END_TF_DOCS -->
