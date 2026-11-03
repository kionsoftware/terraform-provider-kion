# terraform-kion-azure-arm-template

Terraform module for `kion_azure_arm_template`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "azure_arm_template" {
  source = "..."

  deployment_mode          = 1
  name                     = "example"
  resource_group_name      = "example"
  resource_group_region_id = 1
  template                 = "example"
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
| [kion_azure_arm_template.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/azure_arm_template) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| deployment\_mode | Deployment mode for the ARM template. Valid values are either 1 ("incremental") or 2 ("complete") | `number` | n/a | yes |
| name | The name of the ARM template definition in cloudtamer.io | `string` | n/a | yes |
| resource\_group\_name | Name of the resource group where these resources should be deployed | `string` | n/a | yes |
| resource\_group\_region\_id | Database ID of the Azure region where the ARM template should be deployed | `number` | n/a | yes |
| template | Contents of the ARM template to be deployed. | `string` | n/a | yes |
| description | A short description of the ARM template | `string` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the Azure ARM template. Is required if no user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the Azure ARM template. Is required if no group IDs are listed. | `set(number)` | `null` | no |
| template\_parameters | Parameters to fill for the template. Should be the contents of the "properties" attribute on the traditional payload. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_azure\_arm\_template. |
<!-- END_TF_DOCS -->
