# terraform-kion-gcp-service-account

Terraform module for `kion_gcp_service_account`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "gcp_service_account" {
  source = "..."

  email                     = "user@example.com"
  enable_federation_support = false
  gcp_project_id            = "example"
  name                      = "example"
  unique_id                 = "example"
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
| [kion_gcp_service_account.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/gcp_service_account) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| email | Email is the google-designated email id for the service account | `string` | n/a | yes |
| enable\_federation\_support | EnableFederationSupport represents if the service account will be used for federation | `bool` | n/a | yes |
| gcp\_project\_id | GCPProjectID is the Google Cloud Project that the service account was created in | `string` | n/a | yes |
| name | Name is the kebab-case name for the service account | `string` | n/a | yes |
| unique\_id | UniqueID is the client id used to authenticate in tandem with a key | `string` | n/a | yes |
| description | Description is a brief description of the service account | `string` | `null` | no |
| disabled | Disabled is true if the service account is disabled with GCP | `bool` | `null` | no |
| display\_name | DisplayName is a human-friendly name for the service account | `string` | `null` | no |
| oauth\_client\_id | OAuthClientID is the client id for an oauth client in the service account's project | `string` | `null` | no |
| oauth\_client\_secret | OAuthClientSecret is the client secret for an oauth client in the service account's project | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_gcp\_service\_account. |
<!-- END_TF_DOCS -->
