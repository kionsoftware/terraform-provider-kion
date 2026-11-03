# terraform-kion-project

Terraform module for `kion_project`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "project" {
  source = "..."

  name  = "example"
  ou_id = 1
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
| [kion_project.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/project) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the Project in the application. | `string` | n/a | yes |
| ou\_id | ID of the OU containing the project. | `number` | n/a | yes |
| archived | Whether the project is archived. | `bool` | `null` | no |
| auto\_pay | True means the application can use the spend plan to process payments from the account. Should be true unless using a custom module. | `bool` | `null` | no |
| budget | The project budget. | `set(object({ amount = optional(number), data = optional(set(object({ amount = number, datecode = string, funding_source_id = optional(number), priority = optional(number) }))), end_datecode = string, funding_source_ids = optional(set(number)), start_datecode = string }))` | `null` | no |
| default\_aws\_region | Default AWS region that the project will use for assuming into accounts. | `string` | `null` | no |
| description | Description for the project in the application. | `string` | `null` | no |
| labels | The labels applied to the project. | `map(string)` | `null` | no |
| move\_ou\_settings | How to handle the project's cloud rules, financial history and budgets when `ou_id` changes. Ignored unless the project moves. | `set(object({ cloud_rule_setting = optional(string), financial_setting = optional(string), spend_plan_setting = optional(string) }))` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the project. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the project. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |
| permission\_scheme\_id | ID of the permission scheme applied to the project. | `number` | `null` | no |
| project\_funding | n/a | `list(object({ amount = optional(number), end_datecode = optional(string), funding_order = optional(number), funding_source_id = optional(number), start_datecode = optional(string) }))` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_project. |
<!-- END_TF_DOCS -->
