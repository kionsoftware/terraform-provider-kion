# terraform-kion-label

Terraform module for `kion_label`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "label" {
  source = "..."

  color = "#1a2b3c"
  key   = "example"
  value = "example"
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
| [kion_label.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/label) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| color | Hex value of the label color | `string` | n/a | yes |
| key | Key of the label (first part of the label) | `string` | n/a | yes |
| value | Value of the label (second part of the label following the key) | `string` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_label. |
<!-- END_TF_DOCS -->
