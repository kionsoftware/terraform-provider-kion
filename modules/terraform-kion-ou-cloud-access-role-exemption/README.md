# terraform-kion-ou-cloud-access-role-exemption

Terraform module for `kion_ou_cloud_access_role_exemption`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "ou_cloud_access_role_exemption" {
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
| [kion_ou_cloud_access_role_exemption.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/ou_cloud_access_role_exemption) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| ou\_cloud\_access\_role\_id | ID of the ou cloud access role in the application being exempted from. | `number` | `null` | no |
| ou\_id | ID of the ou in the application. | `number` | `null` | no |
| reason | Reason the Cloud Access Role is being exempted. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_ou\_cloud\_access\_role\_exemption. |
<!-- END_TF_DOCS -->
