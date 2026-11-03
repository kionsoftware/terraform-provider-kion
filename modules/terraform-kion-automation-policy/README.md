# terraform-kion-automation-policy

Terraform module for `kion_automation_policy`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "automation_policy" {
  source = "..."

  cloud_provider_policies = []
  name                    = "example"
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
| [kion_automation_policy.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/automation_policy) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| cloud\_provider\_policies | The provider-specific policies this automation policy runs. At least one is required. | `list(object({ apply_to_all_regions = optional(bool), cloud_provider_id = optional(number), policy = optional(string), regions = optional(list(string)), resources = optional(list(string)) }))` | n/a | yes |
| name | Name of the automation policy. Must be unique. | `string` | n/a | yes |
| cloud\_rule\_ids | IDs of the cloud rules the policy is attached to. | `set(number)` | `null` | no |
| description | Description of the automation policy. | `string` | `null` | no |
| enabled | Whether the policy is enabled. | `bool` | `null` | no |
| engine | Engine that runs the policy: 0 (Kion) or 1 (Cloud Custodian). Kion rejects creating a policy with 1, but existing ones can be imported. | `number` | `null` | no |
| owner\_user\_group\_ids | IDs of the user groups that own the policy. | `set(number)` | `null` | no |
| owner\_user\_ids | IDs of the users who own the policy. | `set(number)` | `null` | no |
| scheduled\_frequency | When the policy runs. Omit for a policy that only runs on demand. | `object({ days_of_month = optional(list(number)), days_of_week = optional(list(number)), end_date = optional(string), hour = optional(number), interval_hours = optional(number), minute = optional(number), quarterly_recurrence = optional(number), start_date = optional(string), time_zone_identifier = optional(string), type = optional(number), weekly_recurrence = optional(number) })` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_automation\_policy. |
<!-- END_TF_DOCS -->
