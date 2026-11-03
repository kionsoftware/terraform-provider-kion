# terraform-kion-billing-source-anthropic

Terraform module for `kion_billing_source_anthropic`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "billing_source_anthropic" {
  source = "..."

  admin_api_key      = "example"
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
| [kion_billing_source_anthropic.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/billing_source_anthropic) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| admin\_api\_key | Anthropic Admin API key (sk-ant-admin...). Kion uses it to read the organization's usage and validates it with Anthropic on create. Kion never returns it, so it is not read back. | `string` | n/a | yes |
| billing\_start\_date | First month Kion reads billing data for, as YYYY-MM. | `string` | n/a | yes |
| name | Name of the billing source. | `string` | n/a | yes |
| analytics\_api\_key | Anthropic Enterprise Analytics API key, provisioned at claude.ai with the read:analytics scope. Kion never returns it, so it is not read back. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_billing\_source\_anthropic. |
| organization\_id | organization\_id of the kion\_billing\_source\_anthropic. |
<!-- END_TF_DOCS -->
