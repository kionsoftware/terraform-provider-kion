# terraform-kion-billing-source-aws

Terraform module for `kion_billing_source_aws`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "billing_source_aws" {
  source = "..."

  account_type_id    = 1
  aws_account_number = "example"
  billing_start_date = "2026-01"
  linked_role        = "example"
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
| [kion_billing_source_aws.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/billing_source_aws) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| account\_type\_id | The Account Type ID is the corresponding billing source account's type. 1 - AWS Commercial 2 - Govcloud 4 - AWS C2S 5 - AWS SC2S | `number` | n/a | yes |
| aws\_account\_number | AWS Account Number of the master billing account | `string` | n/a | yes |
| billing\_start\_date | Start date of billing source (YYYY-MM). | `string` | n/a | yes |
| linked\_role | Enter the name of an existing IAM role that has full administrator permissions. If you do not have a custom role, we suggest using Amazon's default OrganizationAccountAccessRole. The role you enter here will be prefilled as the linked role when creating or importing new accounts under this billing source. | `string` | n/a | yes |
| name | Name of the billing source. | `string` | n/a | yes |
| account\_creation | When true, cloudtamer is able to automatically create accounts in this billing source. | `bool` | `null` | no |
| billing\_bucket\_account\_number | AWS Account Number of the s3 bucket holding the billing reports | `string` | `null` | no |
| billing\_region | Region of the s3 bucket holding billing reports (both CUR and DBR reports) | `string` | `null` | no |
| billing\_report\_type | Specify the available billing report types for this Billing Source. Available options are: "none" - Do not use any proprietary billing report "cur" - Use the AWS Cost and Usage Report "dbrrt" - Use the AWS Detailed Billing Report with Resources and Tags If omitted, "cur" is assumed. | `string` | `null` | no |
| bucket\_access\_role | Alternate role for accessing the billing buckets (optional). | `string` | `null` | no |
| cur\_bucket | Name of the bucket containing the cost and usage reports. Required if not using only DBR | `string` | `null` | no |
| cur\_bucket\_region | Region of the bucket containing the cost and usage reports. Required if not using only DBR | `string` | `null` | no |
| cur\_name | Name of the cost and usage report. Required if not using only DBR | `string` | `null` | no |
| cur\_prefix | Report prefix for the cost and usage reports. Required if not using only DBR | `string` | `null` | no |
| focus\_billing\_bucket\_account\_number | AWS Account Number of the s3 bucket holding the FOCUS reports | `string` | `null` | no |
| focus\_billing\_report\_bucket | Name of the bucket containing the FOCUS reports | `string` | `null` | no |
| focus\_billing\_report\_bucket\_region | Region of the bucket containing the FOCUS reports | `string` | `null` | no |
| focus\_billing\_report\_name | Name of the FOCUS billing report | `string` | `null` | no |
| focus\_billing\_report\_prefix | Prefix for the FOCUS billing reports | `string` | `null` | no |
| focus\_bucket\_access\_role | Alternate role for accessing the focus billing buckets (optional). | `string` | `null` | no |
| key\_id | The AWS Access Key used to access the billing s3 bucket | `string` | `null` | no |
| key\_secret | The AWS Secret Access Key used to access the billing s3 bucket | `string` | `null` | no |
| mr\_bucket | Name of the bucket containing the monthly reports (detailed billing reports). Only required when Only CUR is false | `string` | `null` | no |
| only\_dbr | DEPRECATED: Use billing\_report\_type to specify your billing report type Only use the Detailed Billing Report and Detailed Billing Report With Resources And Tags for financial reports. | `bool` | `null` | no |
| skip\_validation | When true, will skip validating the connection to the billing source | `bool` | `null` | no |
| use\_focus\_reports | Use FOCUS Reports - If true, cloudtamer will use FOCUS reports for this billing source | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_billing\_source\_aws. |
<!-- END_TF_DOCS -->
