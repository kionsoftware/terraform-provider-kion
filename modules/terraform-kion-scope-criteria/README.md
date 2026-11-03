# terraform-kion-scope-criteria

Terraform module for `kion_scope_criteria`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "scope_criteria" {
  source = "..."

  criteria    = "{}"
  scope_id    = 1
  start_month = 1
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
| [kion_scope_criteria.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/scope_criteria) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| criteria | The criteria for the scope, as a JSON-encoded string. | `string` | n/a | yes |
| scope\_id | The ID of the parent scope. | `number` | n/a | yes |
| start\_month | Start month for the criteria period (YYYYMM format). | `number` | n/a | yes |
| end\_month | End month for the criteria period (YYYYMM format). | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| criteria\_id | criteria\_id of the kion\_scope\_criteria. |
| id | id of the kion\_scope\_criteria. |
<!-- END_TF_DOCS -->
