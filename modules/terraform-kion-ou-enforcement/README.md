# terraform-kion-ou-enforcement

Terraform module for `kion_ou_enforcement`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "ou_enforcement" {
  source = "..."

  ou_id     = 1
  threshold = 1
  timeframe = "example"
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
| [kion_ou_enforcement.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/ou_enforcement) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| ou\_id | The ID of the OU the enforcement belongs to. | `number` | n/a | yes |
| threshold | Threshold value. Either a dollar amount or percent based on threshold type. | `number` | n/a | yes |
| timeframe | Timeframe of the enforcement. Valid values are "lifetime", "month", "year" | `string` | n/a | yes |
| cloud\_rule\_id | Optional ID of cloud rule that is attached to the enforcement. Use endpoint /v3/cloud-rule to get a list of valid cloud rules and IDs. | `number` | `null` | no |
| description | Optional, user-provided description of the enforcement. | `string` | `null` | no |
| enabled | Whether the enforcement is enabled. | `bool` | `null` | no |
| overburn | Flag that specifies if enforcement will place project in a overburn state when triggered. Options are: true, false. | `bool` | `null` | no |
| service\_id | Option ID of service to set enforcement against. Use endpoint /v3/cloud-provider/service to get a list of valid services and IDs. | `number` | `null` | no |
| threshold\_type | Type of the threshold value. Valid values are "dollar", "percent". | `string` | `null` | no |
| trigger\_planned\_amount\_type | For OU percent enforcements, controls whether percentage is calculated against an aggregate budget or active threshold when applicable. | `string` | `null` | no |
| user\_group\_ids | List of user group IDs that will receive notifications from the enforcement. Is required if no user IDs are listed. | `set(number)` | `null` | no |
| user\_ids | List of user IDs that will receive notifications from the enforcement. Is required if no user group IDs are listed. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_ou\_enforcement. |
| triggered | triggered of the kion\_ou\_enforcement. |
<!-- END_TF_DOCS -->
