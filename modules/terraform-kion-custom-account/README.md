# terraform-kion-custom-account

Terraform module for `kion_custom_account`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "custom_account" {
  source = "..."

  account_name   = "example"
  account_number = "example"
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
| [kion_custom_account.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/custom_account) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| account\_name | The name of the account as it will appear in the application. | `string` | n/a | yes |
| account\_number | The account number or identifier for the custom account. | `string` | n/a | yes |
| payer\_id | The ID of the payer to link the account to. | `number` | n/a | yes |
| project\_id | ID of the project the account belongs to. Changing it moves the account in place through POST /v3/account/{id}/move rather than recreating it: Kion gives the account a new id, the resource adopts it, and the account's financial history stays with the old project (the move's preserve setting). The new project's cloud rules and cloud access roles apply after the move. Requires permission to manage accounts and projects. | `number` | n/a | yes |
| start\_datecode | The start date that the account's spend accrues against the linked project. | `string` | n/a | yes |
| account\_alias | Alias of the account in the application. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_custom\_account. |
<!-- END_TF_DOCS -->
