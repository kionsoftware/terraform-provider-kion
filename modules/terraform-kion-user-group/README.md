# terraform-kion-user-group

Terraform module for `kion_user_group`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "user_group" {
  source = "..."

  idms_id = 1
  name    = "example"
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
| [kion_user_group.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/user_group) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| idms\_id | ID of the IDMS where the user group is located. | `number` | n/a | yes |
| name | Name of the user group in the application. | `string` | n/a | yes |
| add\_self\_as\_viewer | This option will add the user group as a viewer of itself after it is created. | `bool` | `null` | no |
| description | Description of the user group in the application. | `string` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the group. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the group. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |
| user\_ids | IDs of the users in the user group. | `set(number)` | `null` | no |
| viewer\_user\_group\_ids | List of groups IDs who will own the group. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| viewer\_user\_ids | List of user IDs who will own the group. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| created\_at | created\_at of the kion\_user\_group. |
| enabled | enabled of the kion\_user\_group. |
| id | id of the kion\_user\_group. |
<!-- END_TF_DOCS -->
