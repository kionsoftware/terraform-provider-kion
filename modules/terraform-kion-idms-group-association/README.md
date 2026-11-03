# terraform-kion-idms-group-association

Terraform module for `kion_idms_group_association`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "idms_group_association" {
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
| [kion_idms_group_association.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/idms_group_association) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| assertion\_name | AssertionName of the assertion in saml. | `string` | `null` | no |
| assertion\_regex | Regular expression used to determine a match. | `string` | `null` | no |
| idms\_id | ID of the idms the group association will apply to. | `number` | `null` | no |
| update\_on\_login | If the group associations should be updated every time a user logs in. | `bool` | `null` | no |
| user\_group\_id | ID of the user group this assertion will map to. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_idms\_group\_association. |
<!-- END_TF_DOCS -->
