# terraform-kion-gcp-account

Terraform module for `kion_gcp_account`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "gcp_account" {
  source = "..."

  account_name   = "example"
  payer_id       = 1
  project_id     = 1
  start_datecode = "2026-01"
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
| [kion_gcp_account.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/gcp_account) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| account\_name | Name of the account in the application. | `string` | n/a | yes |
| payer\_id | ID of the payer (aka Google Cloud Billing Account) | `number` | n/a | yes |
| project\_id | ID of the project the account belongs to. Changing it moves the account in place through POST /v3/account/{id}/move rather than recreating it: Kion gives the account a new id, the resource adopts it, and the account's financial history stays with the old project (the move's preserve setting). The new project's cloud rules and cloud access roles apply after the move. Requires permission to manage accounts and projects. | `number` | n/a | yes |
| start\_datecode | Date when the GCP org will begin submitting payments against a funding source (YYYY-MM) | `string` | n/a | yes |
| account\_alias | Alias of the account in the application. | `string` | `null` | no |
| account\_type\_id | The AccountTypeID is the corresponding account's type. Will default to 15 for GCP if not given. | `number` | `null` | no |
| google\_cloud\_project\_id | Google Cloud Project ID. Can be found in the project id field here: https://console.cloud.google.com/iam-admin/settings | `string` | `null` | no |
| skip\_access\_checking | When true, the application does not perform periodic access validation. Default is false. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_gcp\_account. |
<!-- END_TF_DOCS -->
