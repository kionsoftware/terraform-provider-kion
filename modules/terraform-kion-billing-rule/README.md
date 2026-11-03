# terraform-kion-billing-rule

Terraform module for `kion_billing_rule`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "billing_rule" {
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
| [kion_billing_rule.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/billing_rule) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| billing\_source\_ids | List of Billing Source IDs attached to the billing rule. | `set(number)` | `null` | no |
| description | Description of the billing rule. The description can be up to 1024 characters long. | `string` | `null` | no |
| end\_month | End month of the billing rule. This must be in the format YYYYMM where 202501 represents January 2025. The end month is exclusive, e.g. if you want the rule to end in January 2025, send a value of 202502. Omit if the end month should be "Never Ends". If included, the value must be after the start month. | `number` | `null` | no |
| name | Name of the billing rule. The name can be up to 255 characters long. | `string` | `null` | no |
| rule\_type | Type of the billing rule. The allowed values are: 1 - Rate Conversion 2 - Markup 3 - Discount | `number` | `null` | no |
| rule\_value | The Value of the Billing Rule. This is handled differently based on the Type of Billing Rules. For instance: For a 5% markup, the value would be 5.0 For a 10% discount, the value would be 10.0 For a 1.214 currency conversion, the value would be 1.214 | `number` | `null` | no |
| start\_month | Start month of the billing rule. This must be in the format YYYYMM where 202501 represents January 2025. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_billing\_rule. |
<!-- END_TF_DOCS -->
