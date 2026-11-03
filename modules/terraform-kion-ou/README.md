# terraform-kion-ou

Terraform module for `kion_ou`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "ou" {
  source = "..."

  name                 = "example"
  parent_ou_id         = 1
  permission_scheme_id = 1
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
| [kion_ou.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/ou) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the OU in the application. | `string` | n/a | yes |
| parent\_ou\_id | ID of the OU containing this OU. A parent OU ID of 0 will create a top-level OU. Changing it moves the OU in place, with its child OUs and projects, rather than recreating it: inherited cloud rules and permissions are then those of the new ancestry. The move goes through Kion's internal POST /v2/ou/{id}/move route, which the public API does not publish, and requires global admin. | `number` | n/a | yes |
| permission\_scheme\_id | ID of the permission scheme applied to the ou. | `number` | n/a | yes |
| description | Description for the OU in the application. | `string` | `null` | no |
| labels | The labels applied to the OU. | `map(string)` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the ou. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the ou. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| created\_at | created\_at of the kion\_ou. |
| id | id of the kion\_ou. |
<!-- END_TF_DOCS -->
