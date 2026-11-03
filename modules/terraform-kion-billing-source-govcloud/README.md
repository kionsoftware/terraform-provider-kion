# terraform-kion-billing-source-govcloud

Terraform module for `kion_billing_source_govcloud`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "billing_source_govcloud" {
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
| [kion_billing_source_govcloud.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/billing_source_govcloud) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| account\_creation\_enabled | Setting for whether account creation in govcloud is enabled or not in this billing source. | `bool` | `null` | no |
| aws\_account\_number | AWS account number. | `string` | `null` | no |
| car\_external\_id | The external ID used when assuming the cloud access role for this billing source. | `string` | `null` | no |
| name | Name of the payer in the application. | `string` | `null` | no |
| payer\_id | The ID of the Payer that this govcloud info is attached to. | `number` | `null` | no |
| service\_external\_id | The external ID used for automated internal actions using the service role for this billing source. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_billing\_source\_govcloud. |
<!-- END_TF_DOCS -->
