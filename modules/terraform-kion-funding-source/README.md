# terraform-kion-funding-source

Terraform module for `kion_funding_source`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "funding_source" {
  source = "..."

  amount         = 1
  end_datecode   = "2026-01"
  name           = "example"
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
| [kion_funding_source.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/funding_source) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| amount | Amount of the funding source. | `number` | n/a | yes |
| end\_datecode | The month this funding source stops being usable - this is exclusive of the date returned (YYYY-MM). | `string` | n/a | yes |
| name | Name of the funding source in the application. | `string` | n/a | yes |
| start\_datecode | The month this funding source starts being usable (YYYY-MM). | `string` | n/a | yes |
| description | Description for the funding source in the application. | `string` | `null` | no |
| ou\_id | ID of the top level OU that will receive the funding from this funding source. | `number` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the funding source. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the funding source. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |
| permission\_scheme\_id | ID of the permission scheme applied to the funding source. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_funding\_source. |
<!-- END_TF_DOCS -->
