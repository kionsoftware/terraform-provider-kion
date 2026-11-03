# terraform-kion-billing-source-gcp

Terraform module for `kion_billing_source_gcp`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "billing_source_gcp" {
  source = "..."

  account_type_id            = 1
  gcp_billing_account_create = { big_query_export = {}, billing_start_date = "2026-01", gcp_id = "example", name = "example", service_account_id = 1 }
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
| [kion_billing_source_gcp.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/billing_source_gcp) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| account\_type\_id | The Account Type ID is the corresponding billing source account's type. 15 - Google Cloud | `number` | n/a | yes |
| gcp\_billing\_account\_create | GCPBillingAccountWithStart contains fields describing a billing account in GCP | `object({ big_query_export = object({ dataset_name = optional(string), focus_view_name = optional(string), gcp_project_id = optional(string), table_format = optional(string), table_name = optional(string) }), billing_account_attribution_account_id = optional(number), billing_start_date = string, gcp_id = string, is_reseller = optional(bool), name = string, service_account_id = number, use_focus = optional(bool), use_proprietary = optional(bool) })` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_billing\_source\_gcp. |
<!-- END_TF_DOCS -->
