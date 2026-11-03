# terraform-kion-custom-variable-override

Terraform module for `kion_custom_variable_override`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "custom_variable_override" {
  source = "..."

  custom_variable_id = "example"
  entity_id          = "example"
  entity_type        = "example"
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
| [kion_custom_variable_override.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/custom_variable_override) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| custom\_variable\_id | The ID of the custom variable. | `string` | n/a | yes |
| entity\_id | The ID of the entity. | `string` | n/a | yes |
| entity\_type | The type of the entity. | `string` | n/a | yes |
| value\_list | The value of the custom variable override as a list. | `list(string)` | `null` | no |
| value\_map | The value of the custom variable override as a map. | `map(string)` | `null` | no |
| value\_string | The value of the custom variable override as a string. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_custom\_variable\_override. |
<!-- END_TF_DOCS -->
