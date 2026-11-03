# terraform-kion-billing-source-azure

Terraform module for `kion_billing_source_azure`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "billing_source_azure" {
  source = "..."

  account_type_id    = 1
  billing_start_date = "2026-01"
  name               = "example"
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
| [kion_billing_source_azure.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/billing_source_azure) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| account\_type\_id | Kion account type of the billing source: 3 Azure CSP, 6 Azure EA, 16 Azure MCA, or their Government and Secret variants. | `number` | n/a | yes |
| billing\_start\_date | First month Kion reads billing data for, as YYYY-MM. | `string` | n/a | yes |
| name | Name of the billing source. | `string` | n/a | yes |
| account\_creation | Whether accounts can be created under this billing source. | `bool` | `null` | no |
| app\_id | Application (client) ID of the Azure service principal. | `string` | `null` | no |
| azure\_csp\_id | ID of the parent Azure CSP record | `number` | `null` | no |
| client\_secret | Client secret of the Azure service principal. Kion never returns it, so it is not read back: a change made outside Terraform is not detected. | `string` | `null` | no |
| customer\_id | The Azure customer or enrollment id. | `string` | `null` | no |
| domain | Azure AD tenant domain. | `string` | `null` | no |
| focus\_storage\_container | Storage container holding the FOCUS export. | `string` | `null` | no |
| focus\_storage\_prefix | Path prefix of the FOCUS export in the container. | `string` | `null` | no |
| focus\_storage\_primary\_endpoint | Primary blob endpoint of the FOCUS export storage account. | `string` | `null` | no |
| primary\_management\_group\_azure\_id | Azure id of the management group new subscriptions are placed under. | `string` | `null` | no |
| resource\_group\_creation | Whether Kion may create resource groups under this billing source. | `bool` | `null` | no |
| skip\_validation | Create or update without first testing the Azure credentials and storage access. Kion still stores the values as given. | `bool` | `null` | no |
| storage\_container | Storage container holding the billing export. | `string` | `null` | no |
| storage\_prefix | Path prefix of the billing export in the container. | `string` | `null` | no |
| storage\_primary\_endpoint | Primary blob endpoint of the billing export storage account. | `string` | `null` | no |
| subscription\_creation | Whether Kion may create subscriptions under this billing source. | `bool` | `null` | no |
| tenant\_level\_charge\_attribution\_account\_id | ID of the account tenant-level charges are attributed to. | `number` | `null` | no |
| use\_focus\_reports | Whether Kion reads the FOCUS export. | `bool` | `null` | no |
| use\_proprietary\_reports | Whether Kion reads the Azure cost export. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_billing\_source\_azure. |
<!-- END_TF_DOCS -->
