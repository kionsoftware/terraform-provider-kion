# terraform-kion-budget

Terraform module for `kion_budget`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "budget" {
  source = "..."

  end_datecode   = "2026-01"
  start_datecode = "2026-01"
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
| [kion_budget.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/budget) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| end\_datecode | Year and month the budget ends. This is an exclusive date. | `string` | n/a | yes |
| start\_datecode | Year and month the budget starts. | `string` | n/a | yes |
| amount | Total amount for the budget. This is required if data is not specified. Budget entries are created between start\_datecode and end\_datecode (exclusive) with the amount evenly distributed across the months. | `number` | `null` | no |
| funding\_source\_ids | Optional funding source IDs to use when data is not specified. This value is ignored is data is specified. If specified, the amount is distributed evenly across months and funding sources. Funding sources will be processed in order from first to last. | `set(number)` | `null` | no |
| ou\_id | ID of OU this budget is attached to. Required for OU thresholds. | `number` | `null` | no |
| project\_id | ID of project this budget is attached to. Required for project budgets. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_budget. |
<!-- END_TF_DOCS -->
