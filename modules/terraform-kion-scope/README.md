# terraform-kion-scope

Terraform module for `kion_scope`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "scope" {
  source = "..."

  name = "example"
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
| [kion_scope.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/scope) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the role in the application. | `string` | n/a | yes |
| alias | Alias of the scope in the application. Limited to 16 characters. | `string` | `null` | no |
| criteria | The criteria for the scope, as a JSON-encoded string. | `string` | `null` | no |
| description | Description of the scope in the application. | `string` | `null` | no |
| end\_datecode | End datecode for the scope (YYYYMM format), if applicable. | `number` | `null` | no |
| project\_id | ID of the project associated with this scope. | `number` | `null` | no |
| start\_datecode | Start datecode for the scope (YYYYMM format). | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_scope. |
<!-- END_TF_DOCS -->
