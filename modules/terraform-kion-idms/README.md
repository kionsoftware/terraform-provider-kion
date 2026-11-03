# terraform-kion-idms

Terraform module for `kion_idms`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "idms" {
  source = "..."

  idms_type_id        = 1
  name                = "example"
  password_expiration = 1
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
| [kion_idms.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/idms) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| idms\_type\_id | ID of the IDMS type in the application. 1 is Internal Directory, 2 is Active Directory, 3 is SAML. | `number` | n/a | yes |
| name | Name of the IDMS in the application. | `string` | n/a | yes |
| password\_expiration | Number of days until a password reset dialog displays on login. | `number` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_idms. |
<!-- END_TF_DOCS -->
