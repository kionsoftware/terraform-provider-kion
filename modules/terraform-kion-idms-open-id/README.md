# terraform-kion-idms-open-id

Terraform module for `kion_idms_open_id`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "idms_open_id" {
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
| [kion_idms_open_id.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/idms_open_id) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| access\_rules | AccessRules defines the console access defined for matching users in Kion | `list(object({ assertion_name = optional(string), assertion_regex = optional(string), cloudtamer_access_level_id = optional(number) }))` | `null` | no |
| authorization\_endpoint | AuthorizationEndpoint is the endpoint the OpenID provider determines to kick off the authentication process | `string` | `null` | no |
| client\_id | ClientID that links Kion to the OpenID Provider, determined when configuring OpenID | `string` | `null` | no |
| email\_claim | EmailClaim defines an optional mapping of an OpenID claim value to a user's email attribute in Kion | `string` | `null` | no |
| first\_name\_claim | FirstNameClaim defines an optional mapping of an OpenID claim value to a user's first name attribute in Kion | `string` | `null` | no |
| issuer | Issuer is the URL for the OpenID Provider | `string` | `null` | no |
| jwks\_uri | JwksURI describes the endpoint of an OpenID provider used to retrieve token security information for validation | `string` | `null` | no |
| last\_name\_claim | LastNameClaim defines an optional mapping of an OpenID claim value to a user's last name attribute in Kion | `string` | `null` | no |
| name | Name of the OpenID in the application. | `string` | `null` | no |
| phone\_claim | PhoneClaim defines an optional mapping of an OpenID claim value to a user's phone attribute in Kion | `string` | `null` | no |
| scopes | Scopes define the claim variables that an OpenID provider will include in an auth request. Specifics depend on the OpenID provider. | `set(string)` | `null` | no |
| username\_claim | UsernameClaim defines an optional mapping of an OpenID claim value to a user's username attribute in Kion. Required to uniquely identify users | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_idms\_open\_id. |
<!-- END_TF_DOCS -->
