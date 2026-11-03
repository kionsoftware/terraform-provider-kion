# terraform-kion-compliance-program

Terraform module for `kion_compliance_program`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "compliance_program" {
  source = "..."

  name                       = "example"
  compliance_program_version = "example"
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
| [kion_compliance_program.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/compliance_program) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| compliance\_program\_version | Version of the Compliance Program in the application. | `string` | n/a | yes |
| name | Name of the Compliance Program in the application. This is the proper name of the program meant for formal page/card/document titles in the UI. | `string` | n/a | yes |
| description | Description for the Compliance Program in the application. | `string` | `null` | no |
| grouping\_type | GroupType of the Compliance Program in the application. | `string` | `null` | no |
| terse\_name | TerseName of the Compliance Program in the application. This is a concise name meant for easy identification and small UI spaces like program icons. | `string` | `null` | no |
| verbose\_name | VerboseName of the Compliance Program in the application. This is a more detailed program name meant to suppliment the Name or TerseName in specific contexts. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_compliance\_program. |
<!-- END_TF_DOCS -->
