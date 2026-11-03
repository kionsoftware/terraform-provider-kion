# terraform-kion-compliance-level

Terraform module for `kion_compliance_level`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "compliance_level" {
  source = "..."

  compliance_program_id = 1
  name                  = "example"
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
| [kion_compliance_level.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/compliance_level) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| compliance\_program\_id | ComplianceProgramID of the compliance level. | `number` | n/a | yes |
| name | Name of the compliance level in the application. | `string` | n/a | yes |
| description | Description for the compliance level in the application. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_compliance\_level. |
<!-- END_TF_DOCS -->
