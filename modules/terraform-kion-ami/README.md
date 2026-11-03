# terraform-kion-ami

Terraform module for `kion_ami`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "ami" {
  source = "..."

  account_id = 1
  aws_ami_id = "example"
  name       = "example"
  region     = "example"
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
| [kion_ami.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/ami) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| account\_id | AWS account application ID where the AMI is stored. | `number` | n/a | yes |
| aws\_ami\_id | Image ID of the AMI from AWS. | `string` | n/a | yes |
| name | Name of the AMI in the application. | `string` | n/a | yes |
| region | AWS region where the AMI exists. | `string` | n/a | yes |
| description | Description for the AMI in the application. | `string` | `null` | no |
| expiration\_alert\_number | TThe amount of time before the expiration alert is shown | `number` | `null` | no |
| expiration\_alert\_unit | The unit for the expiration alert lead time (e.g. "days"). | `string` | `null` | no |
| expiration\_notify | Will notify the owners that the shared AMI has expired | `bool` | `null` | no |
| expiration\_warning\_number | The amount of time before the expiration warning is sent | `number` | `null` | no |
| expiration\_warning\_unit | The unit for the expiration warning lead time (e.g. "days"). | `string` | `null` | no |
| expires\_at | The expiration timestamp for the AMI, as an RFC3339 string. | `string` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the ami. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the ami. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |
| sync\_deprecation | Will sync the expiration date from the system into the AMI in AWS. | `bool` | `null` | no |
| sync\_tags | Will sync the AWS tags from the source AMI into all the accounts where the AMI is shared. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_ami. |
<!-- END_TF_DOCS -->
