# terraform-kion-account-cache

Terraform module for `kion_account_cache`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "account_cache" {
  source = "..."

  account_name = "example"
  payer_id     = 1
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
| [kion_account_cache.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/account_cache) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| account\_name | Name of the account in the application. | `string` | n/a | yes |
| payer\_id | ID of the payer where the application can find the billing data for the account. | `number` | n/a | yes |
| account\_alias | Alias of the account in the application. | `string` | `null` | no |
| account\_email | Root email of the account if created inside the application, otherwise it will be an empty string. Note: If an account placeholder email (e.g. `jdoe+%s@example.com`) has been set, account\_email MUST be empty. In this case, account\_email will be set to the account placeholder email formatted with the account\_name field (e.g. `jdoe+account_name@example.com`). Otherwise, account\_email is required. | `string` | `null` | no |
| account\_number | The account number. | `string` | `null` | no |
| account\_type\_id | AccountTypeID is used to specify the AWS Account Type being created. If not explicitly set, the partition that Kion is running in will be used. For AWS Commercial accounts, this value is 1. For AWS Government accounts, this value is 2. For AWS C2S accounts, this value is 4. For AWS SC2S accounts, this value is 5. | `number` | `null` | no |
| commercial\_account\_name | Name of the Commercial Account in the application. If none is provided, a default value of "{Name} - Commercial" will be used instead. If include\_linked\_account\_spend is set to true, this value will be ignored. | `string` | `null` | no |
| create\_govcloud | CreateGovcloud determines if this account should be created in govcloud. The default value is false. | `bool` | `null` | no |
| gov\_account\_name | Name of the GovCloud Account in the application. If no value is provided, "{Name} - GovCloud" will be used if create\_govcloud is true | `string` | `null` | no |
| include\_linked\_account\_spend | When true, the application will include linked account spend from linked govcloud or commercial accounts. If a Govcloud account is created with include linked account spend set to false, cloudtamer.io will create both a commercial and govcloud account and add them both to the cache. Default is false. | `bool` | `null` | no |
| linked\_role | Name of the AWS Organizations service role. Default as well as what AWS recommends is: OrganizationAccountAccessRole. | `string` | `null` | no |
| organizational\_unit | PayerOrganizationalUnit represents an organizational unit in an AWS payer's organization. | `object({ name = optional(string), org_unit_id = optional(string) })` | `null` | no |
| skip\_access\_checking | Whether access checking is skipped. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| car\_external\_id | car\_external\_id of the kion\_account\_cache. |
| created\_at | created\_at of the kion\_account\_cache. |
| id | id of the kion\_account\_cache. |
| linked\_account\_number | linked\_account\_number of the kion\_account\_cache. |
| service\_external\_id | service\_external\_id of the kion\_account\_cache. |
<!-- END_TF_DOCS -->
