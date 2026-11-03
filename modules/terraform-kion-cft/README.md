# terraform-kion-cft

Terraform module for `kion_cft`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "cft" {
  source = "..."

  name    = "example"
  policy  = "{}"
  regions = []
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
| [kion_cft.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/cft) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the Cloudformation template in the application. | `string` | n/a | yes |
| policy | Body of the CloudFormation template in JSON or YAML. | `string` | n/a | yes |
| regions | List of the AWS regions where the CloudFormation template applies. | `set(string)` | n/a | yes |
| description | Description of the Cloudformation template in the application. | `string` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the CloudFormation template. Is required if no user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the CloudFormation template. Is required if no group IDs are listed. | `set(number)` | `null` | no |
| region | DEPRECATED! USE THE regions FIELD. AWS region where the CloudFormation template applies. | `string` | `null` | no |
| sns\_arns | List of comma separated AWS SNS ARNs that will trigger once the CFT is done applying. | `string` | `null` | no |
| tags | AWS Stack Tags used in this role when accessing the AWS console. | `list(object({ tag_key = optional(string), tag_value = optional(string) }))` | `null` | no |
| template\_parameters | List of CloudFormation parameters in a JSON array. | `string` | `null` | no |
| termination\_protection | Sets the termination protection status for this CFT. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_cft. |
<!-- END_TF_DOCS -->
