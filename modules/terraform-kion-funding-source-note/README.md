# terraform-kion-funding-source-note

Terraform module for `kion_funding_source_note`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "funding_source_note" {
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
| [kion_funding_source_note.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/funding_source_note) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| funding\_source\_id | ID of the funding source the note belongs to. | `number` | `null` | no |
| name | Name of the note. | `string` | `null` | no |
| text | Body text of the note. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_funding\_source\_note. |
<!-- END_TF_DOCS -->
