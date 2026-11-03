# terraform-kion-user

Terraform module for `kion_user`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "user" {
  source = "..."

  email      = "user@example.com"
  first_name = "example"
  idms_id    = 1
  last_name  = "example"
  username   = "example"
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
| [kion_user.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/user) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| email | Email address of the user. | `string` | n/a | yes |
| first\_name | First name of the user. | `string` | n/a | yes |
| idms\_id | ID of the IDMS where the user exists. | `number` | n/a | yes |
| last\_name | Last name of the user. | `string` | n/a | yes |
| username | Username of the user. | `string` | n/a | yes |
| mfa | The ID of the MFA type. Options: 1 is Webauthn (Yubikey), 2 is TOTP (Google Auth). | `number` | `null` | no |
| phone | Phone number of the user. | `string` | `null` | no |
| user\_group\_ids | List of IDs of groups the user is in. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_user. |
<!-- END_TF_DOCS -->
