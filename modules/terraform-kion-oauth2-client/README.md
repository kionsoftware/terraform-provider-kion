# terraform-kion-oauth2-client

Terraform module for `kion_oauth2_client`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "oauth2_client" {
  source = "..."

  grant_types = []
  name        = "example"
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
| [kion_oauth2_client.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/oauth2_client) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| grant\_types | Allowed grant types. | `list(string)` | n/a | yes |
| name | Name of the OAuth client. | `string` | n/a | yes |
| allowed\_scopes | Kion product scopes this client may request. Supported values are "mcp" and "api". An empty list grants no Kion product access. | `list(string)` | `null` | no |
| description | Description of the OAuth client. | `string` | `null` | no |
| is\_public | Whether the client is public (no secret). Changing it replaces the client, which issues a new secret. | `bool` | `null` | no |
| owner\_user\_id | Owner of the client. Omitted from configuration, the current owner is kept; Kion makes a client with no owner a global application. | `number` | `null` | no |
| redirect\_uris | Allowed redirect URIs. | `list(string)` | `null` | no |
| service\_account\_user\_id | Service account user ID for client\_credentials grant. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_oauth2\_client. |
| is\_dynamically\_registered | is\_dynamically\_registered of the kion\_oauth2\_client. |
| secret | secret of the kion\_oauth2\_client. |
<!-- END_TF_DOCS -->
