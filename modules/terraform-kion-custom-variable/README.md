# terraform-kion-custom-variable

Terraform module for `kion_custom_variable`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "custom_variable" {
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
| [kion_custom_variable.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/custom_variable) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| default\_value\_list | Default value when the custom variable type is list. | `list(string)` | `null` | no |
| default\_value\_map | Default value when the custom variable type is map. | `map(string)` | `null` | no |
| default\_value\_string | Default value when the custom variable type is string. | `string` | `null` | no |
| description | Description is the description of the custom variable. The description can be up to 1024 characters. | `string` | `null` | no |
| key\_validation\_message | KeyValidationMessage is the message displayed when the custom variable key(s) do not match the regular expression. The key validation message can be up to 255 characters. | `string` | `null` | no |
| key\_validation\_regex | KeyValidationRegex is the regular expression used to validate the custom variable key(s) within the map custom variable type. The key validation regex can be up to 255 characters. | `string` | `null` | no |
| name | Name is the name of the custom variable. The name can be up to 255 characters. | `string` | `null` | no |
| owner\_user\_group\_ids | OwnerUGroupIDs is the list of user group IDs who will own the custom variable. | `set(number)` | `null` | no |
| owner\_user\_ids | OwnerUserIDs is the list of user IDs who will own the custom variable. | `set(number)` | `null` | no |
| type | Type is the type of the custom variable. The supported types are string, list, and map. | `string` | `null` | no |
| value\_validation\_message | ValueValidationMessage is the message displayed when the custom variable value(s) do not match the regular expression. The validation message can be up to 255 characters. | `string` | `null` | no |
| value\_validation\_regex | ValueValidationRegex is the regular expression used to validate the custom variable value(s). The validation regex can be up to 255 characters. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_custom\_variable. |
<!-- END_TF_DOCS -->
