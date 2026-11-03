# terraform-kion-idms-open-id-access-rule

Terraform module for `kion_idms_open_id_access_rule`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "idms_open_id_access_rule" {
  source = "..."

  open_id_id = 1
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
| [kion_idms_open_id_access_rule.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/idms_open_id_access_rule) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| open\_id\_id | The ID of the OpenID IDMS the access rule belongs to. | `number` | n/a | yes |
| assertion\_name | AssertionName name of the assertion in OpenID. | `string` | `null` | no |
| assertion\_regex | AssertionRegex is the regular expression used to determine a match. | `string` | `null` | no |
| cloudtamer\_access\_level\_id | CloudtamerAccessLevelID is the ID that represents access level. 1 represents full access, 2 represents no cloud account access and 3 represents no access. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_idms\_open\_id\_access\_rule. |
<!-- END_TF_DOCS -->
