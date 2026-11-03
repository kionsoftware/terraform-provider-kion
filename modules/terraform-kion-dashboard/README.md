# terraform-kion-dashboard

Terraform module for `kion_dashboard`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "dashboard" {
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
| [kion_dashboard.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/dashboard) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| config | JSON configuration (cards | `string` | `null` | no |
| description | Description of the dashboard. | `string` | `null` | no |
| is\_default | Whether this is the default dashboard. | `bool` | `null` | no |
| name | Name of the dashboard. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| created\_at | created\_at of the kion\_dashboard. |
| created\_by\_user\_id | created\_by\_user\_id of the kion\_dashboard. |
| id | id of the kion\_dashboard. |
| updated\_at | updated\_at of the kion\_dashboard. |
<!-- END_TF_DOCS -->
