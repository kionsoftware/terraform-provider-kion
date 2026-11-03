# terraform-kion-aws-iam-policy

Terraform module for `kion_aws_iam_policy`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "aws_iam_policy" {
  source = "..."

  name   = "example"
  policy = "{}"
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
| [kion_aws_iam_policy.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/aws_iam_policy) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the IAM policy in the application. | `string` | n/a | yes |
| policy | Text of the IAM policy in AWS to be stored in AWS. | `string` | n/a | yes |
| aws\_iam\_path | Text of the IAM Path in AWS to be stored in AWS. | `string` | `null` | no |
| car\_restricted | True if the policy has been restricted for use by a subset of Users/UGroups on a Cloud Access Role. | `bool` | `null` | no |
| car\_restricted\_user\_group\_ids | List of groups IDs who have been allowed to use the iam policy on Cloud Access Roles in the system. | `set(number)` | `null` | no |
| car\_restricted\_user\_ids | List of user IDs who have been allowed to use the iam policy on Cloud Access Roles in the system. | `set(number)` | `null` | no |
| description | Description for the IAM policy in the application. | `string` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the iam policy. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the iam policy. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| aws\_managed\_policy | aws\_managed\_policy of the kion\_aws\_iam\_policy. |
| id | id of the kion\_aws\_iam\_policy. |
| path\_suffix | path\_suffix of the kion\_aws\_iam\_policy. |
| system\_managed\_policy | system\_managed\_policy of the kion\_aws\_iam\_policy. |
<!-- END_TF_DOCS -->
