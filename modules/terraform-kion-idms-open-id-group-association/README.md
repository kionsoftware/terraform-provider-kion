# terraform-kion-idms-open-id-group-association

Terraform module for `kion_idms_open_id_group_association`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "idms_open_id_group_association" {
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
| [kion_idms_open_id_group_association.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/idms_open_id_group_association) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| assertion\_name | AssertionName name of the assertion in OpenID. | `string` | `null` | no |
| assertion\_regex | AssertionRegex is the regular expression used to determine a match. | `string` | `null` | no |
| open\_id\_id | The ID of the OpenID IDMS the group association belongs to. | `number` | `null` | no |
| update\_on\_login | ShouldUpdateOnLogin denotes if the group associations should be updated every time a user logs in. | `bool` | `null` | no |
| user\_group\_id | UgroupID is the ID of the user group this assertion will map to. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_idms\_open\_id\_group\_association. |
<!-- END_TF_DOCS -->
