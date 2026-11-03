# terraform-kion-aws-resource-tag

Terraform module for `kion_aws_resource_tag`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "aws_resource_tag" {
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
| [kion_aws_resource_tag.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/aws_resource_tag) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| resource\_key | ResourceKey is the resource tag's key. | `string` | `null` | no |
| resource\_value | ResourceValue is the resource tag's value. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_aws\_resource\_tag. |
<!-- END_TF_DOCS -->
