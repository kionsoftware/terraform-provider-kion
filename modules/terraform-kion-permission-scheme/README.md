# terraform-kion-permission-scheme

Terraform module for `kion_permission_scheme`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "permission_scheme" {
  source = "..."

  name = "example"
  type = "example"
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
| [kion_permission_scheme.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/permission_scheme) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the permission scheme in the application. | `string` | n/a | yes |
| type | Specifies the type of policy. Options are: global, ou, project, or funding\_source. | `string` | n/a | yes |
| roles | n/a | `list(object({ permission_id = number, role_id = number }))` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_permission\_scheme. |
<!-- END_TF_DOCS -->
