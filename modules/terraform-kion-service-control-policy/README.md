# terraform-kion-service-control-policy

Terraform module for `kion_service_control_policy`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "service_control_policy" {
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
| [kion_service_control_policy.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/service_control_policy) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the Service Control policy in the application. | `string` | n/a | yes |
| policy | Text of the Service Control policy in AWS to be stored in AWS. | `string` | n/a | yes |
| description | Description for the Service Control policy in the application. | `string` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the service\_control policy. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the service\_control policy. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_service\_control\_policy. |
<!-- END_TF_DOCS -->
