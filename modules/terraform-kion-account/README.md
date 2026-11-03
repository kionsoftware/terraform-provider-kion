# terraform-kion-account

Terraform module for `kion_account`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "account" {
  source = "..."

  name     = "example"
  payer_id = 1
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
| [kion_account.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/account) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | The name of the AWS account within Kion. | `string` | n/a | yes |
| payer\_id | The ID of the billing source containing billing data for this account. | `number` | n/a | yes |
| account\_alias | Account alias is an optional short unique name that helps identify the account within Kion. | `string` | `null` | no |
| account\_number | The account number of the AWS account. If account\_number is provided, the existing account will be imported into Kion. If account\_number is omitted, a new account will be created. | `string` | `null` | no |
| account\_type\_id | An ID representing the account type within Kion. | `number` | `null` | no |
| aws\_organizational\_unit | Where to place this account within AWS Organization when creating an account. | `object({ name = optional(string), org_unit_id = optional(string) })` | `null` | no |
| commercial\_account\_name | The name used when creating new commercial account. | `string` | `null` | no |
| create\_govcloud | True to create an AWS GovCloud account. | `bool` | `null` | no |
| email | The root email address to associate with a new account. Required when creating a new account unless an account placeholder email has been set. | `string` | `null` | no |
| gov\_account\_name | The name used when creating new GovCloud account. | `string` | `null` | no |
| include\_linked\_account\_spend | True to associate spend from a linked GovCloud account with this account. | `bool` | `null` | no |
| labels | A map of labels to assign to the account. The labels must already exist in Kion. | `map(string)` | `null` | no |
| linked\_role | The AWS organization service role. | `string` | `null` | no |
| move\_project\_settings | Parameters used when moving an account between Kion projects. These settings are ignored unless moving an account. | `object({ financials = optional(string), move_datecode = optional(number) })` | `null` | no |
| project\_id | The ID of the Kion project to place this account within. If empty, the account will be placed within the account cache. | `number` | `null` | no |
| skip\_access\_checking | True to skip periodic access checking on the account. | `bool` | `null` | no |
| start\_datecode | Date when the AWS account will starting submitting payments against a funding source (YYYY-MM). Required if placing an account within a project. | `string` | `null` | no |
| use\_org\_account\_info | True to keep the account name and email address in Kion in sync with the account name and email address as set in AWS Organization. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| car\_external\_id | car\_external\_id of the kion\_account. |
| created\_at | created\_at of the kion\_account. |
| id | id of the kion\_account. |
| linked\_account\_number | linked\_account\_number of the kion\_account. |
| location | location of the kion\_account. |
| service\_external\_id | service\_external\_id of the kion\_account. |
<!-- END_TF_DOCS -->
