# terraform-kion-billing-source

Terraform module for `kion_billing_source`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "billing_source" {
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
| [kion_billing_source.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/billing_source) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| aws\_connection | n/a | `object({ account_number = optional(string), bucket_access_role = optional(string), storage_bucket = optional(string), storage_prefix = optional(string), storage_region = optional(string) })` | `null` | no |
| azure\_connection | CustomBillingSourceAzureConnection defines the Azure storage connection for a custom billing source. Provide either credentialed\_billing\_source\_id (reuse an existing Azure billing source's tenant credentials) or the tenant credential fields, not both. | `object({ credentialed_billing_source_id = optional(number), storage_container = optional(string), storage_prefix = optional(string), storage_primary_endpoint = optional(string), tenant_app_id = optional(string), tenant_client_secret = optional(string), tenant_cloud_partition_id = optional(number), tenant_domain = optional(string) })` | `null` | no |
| billing\_start\_date | Start date of billing source (YYYY-MM). | `string` | `null` | no |
| name | Name of the billing source. | `string` | `null` | no |
| skip\_validation | When true, will skip validating the connection to defined bucket. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_billing\_source. |
<!-- END_TF_DOCS -->
