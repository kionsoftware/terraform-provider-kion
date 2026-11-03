# terraform-kion-billing-source-oci

Terraform module for `kion_billing_source_oci`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "billing_source_oci" {
  source = "..."

  billing_start_date = "2026-01"
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
| [kion_billing_source_oci.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/billing_source_oci) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| billing\_start\_date | Start date of billing source (YYYY-MM). | `string` | n/a | yes |
| account\_type\_id | The Account Type ID is the corresponding billing source account's type. 26 - OCI Commercial 27 - OCI Government 28 - OCI Federal | `number` | `null` | no |
| fingerprint | The api access fingerprint | `string` | `null` | no |
| is\_parent\_tenancy | Indicates whether the billing source is a parent OCI tenancy. | `bool` | `null` | no |
| name | Name of the billing source. | `string` | `null` | no |
| private\_key | The private key used for API Access | `string` | `null` | no |
| region | The OCI default api region | `string` | `null` | no |
| skip\_validation | When true, will skip validating the connection to the billing source | `bool` | `null` | no |
| tenancy\_ocid | The OCID of the tenancy | `string` | `null` | no |
| use\_focus\_reports | Use FOCUS Reports - If true, Kion will use FOCUS reports for this billing source | `bool` | `null` | no |
| use\_proprietary\_reports | Use Proprietary Reports - If true, Kion will use proprietary reports for this billing source | `bool` | `null` | no |
| user\_ocid | The OCID of the api user | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_billing\_source\_oci. |
<!-- END_TF_DOCS -->
