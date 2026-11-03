# terraform-kion-gcp-iam-role

Terraform module for `kion_gcp_iam_role`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "gcp_iam_role" {
  source = "..."

  name = "example"
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
| [kion_gcp_iam_role.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/gcp_iam_role) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the GCP Role in the application. | `string` | n/a | yes |
| car\_restricted\_user\_group\_ids | List of groups IDs who have been allowed to use the GCP Role on Cloud Access Roles in the system. | `set(number)` | `null` | no |
| car\_restricted\_user\_ids | List of user IDs who have been allowed to use the GCP Role on Cloud Access Roles in the system. | `set(number)` | `null` | no |
| description | Description for the GCP Role in the application. | `string` | `null` | no |
| gcp\_role\_launch\_stage | GCPRoleLaunchStage is the launch stage for a gcp role | `number` | `null` | no |
| owner\_user\_group\_ids | List of group IDs who will own the GCP Role. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the GCP Role. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |
| role\_denials | List of GCP Permissions to deny when applying this role. Wildcards are supported. | `set(string)` | `null` | no |
| role\_permissions | List of GCP Permissions to assign the role | `set(string)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_gcp\_iam\_role. |
<!-- END_TF_DOCS -->
