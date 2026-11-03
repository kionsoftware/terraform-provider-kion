# terraform-kion-app-config

Terraform module for `kion_app_config`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "app_config" {
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
| [kion_app_config.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/app_config) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| all\_users\_see\_ou\_names | Indicates whether all users can see the names of OU's in the organization chart. | `bool` | `null` | no |
| allocation\_mode | Indicates if allocation mode is enabled in the application. | `bool` | `null` | no |
| allow\_custom\_permission\_schemes | Indicates whether custom permission schemes are allowed or not. | `bool` | `null` | no |
| app\_api\_key\_creation\_enabled | Indicates if App API Key creation is enabled. | `bool` | `null` | no |
| app\_api\_key\_lifespan | Indicates the lifespan of App API Keys in days. | `number` | `null` | no |
| app\_api\_key\_limit | Indicates the max amount of App API Keys per user. | `number` | `null` | no |
| aws\_access\_key\_creation\_enabled | Indicates whether AWS access keys creation is enabled. | `bool` | `null` | no |
| budget\_mode | Indicates if budget mode is enabled in the application. | `bool` | `null` | no |
| cloud\_rule\_group\_ownership\_only | Indicates if cloud rules are restricted to User Group ownership only. Setting this to true will remove all users from cloud rules. This cannot be undone. | `bool` | `null` | no |
| cost\_savings\_allow\_terminate | Indicates whether resource termination is allowed in-app. | `bool` | `null` | no |
| cost\_savings\_enabled | Indicates whether Cost Savings is enabled or not. | `bool` | `null` | no |
| cost\_savings\_post\_token\_life\_hours | Post token life (hours) for Cloud Custodian webhook actions to execute. | `number` | `null` | no |
| default\_org\_chart\_view | Defines the default organization chart view. | `string` | `null` | no |
| enforce\_funding | Indicates whether spend plans or budgets must be created on all projects. | `bool` | `null` | no |
| enforce\_funding\_sources | Indicates whether every project should have a funding source. | `bool` | `null` | no |
| event\_driven\_enabled | Indicates whether event driven is enabled or not. | `bool` | `null` | no |
| reserved\_instances\_enabled | Indicates whether reserved instances are enabled or not. | `bool` | `null` | no |
| resource\_inventory\_enabled | Indicates whether resource inventory is enabled or not. | `bool` | `null` | no |
| smtp\_enabled | Indicates whether SMTP is enabled or not. | `bool` | `null` | no |
| smtp\_from | The SMTP from address. | `string` | `null` | no |
| smtp\_host | The SMTP host. | `string` | `null` | no |
| smtp\_password | The SMTP password. | `string` | `null` | no |
| smtp\_port | The SMTP port. | `number` | `null` | no |
| smtp\_skip\_verify | Indicates if the app should skip SMTP verification. | `bool` | `null` | no |
| smtp\_username | The SMTP username. | `string` | `null` | no |
| supported\_aws\_regions | The list of supported AWS regions. | `set(string)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_app\_config. |
<!-- END_TF_DOCS -->
