# terraform-kion-cloud-rule

Terraform module for `kion_cloud_rule`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "cloud_rule" {
  source = "..."

  name = "example"
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
| [kion_cloud_rule.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/cloud_rule) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the Cloud Rule in the application. | `string` | n/a | yes |
| automation\_policy\_ids | List of Automation Policies attached to the Cloud Rule | `set(number)` | `null` | no |
| azure\_arm\_template\_definition\_ids | List of Azure ARM template definition IDs to attach to the Cloud Rule. | `set(number)` | `null` | no |
| azure\_policy\_definition\_ids | List of Azure Policy IDs to attach to the Cloud Rule. | `set(number)` | `null` | no |
| azure\_role\_definition\_ids | List of Azure Role Definition IDs to attach to the Cloud Rule. | `set(number)` | `null` | no |
| cft\_ids | List of CloudFormation template IDs to attach to the Cloud Rule. | `set(number)` | `null` | no |
| compliance\_standard\_ids | List of Compliance Standards attached to the Cloud Rule | `set(number)` | `null` | no |
| concurrent\_cft\_sync | Whether to run CFTs concurrently or not. If true, the application will deploy all templates at once in any order. (Faster) If false, the application will deploy each template in order and wait for completion before advancing to the next. (Slower) | `bool` | `null` | no |
| description | Description of the Cloud Rule in more detail. | `string` | `null` | no |
| gcp\_iam\_role\_ids | List of Google Cloud IAM role IDs to attach to the Cloud Rule. | `set(number)` | `null` | no |
| iam\_policy\_ids | List of IAM Policy IDs to attach to the Cloud Rule. | `set(number)` | `null` | no |
| internal\_ami\_ids | List of AMIs to attach to the Cloud Rule. | `set(number)` | `null` | no |
| internal\_portfolio\_ids | List of Service Catalog Portfolio IDs attached to the Cloud Rule. | `set(number)` | `null` | no |
| labels | The labels applied to the cloud rule. | `map(string)` | `null` | no |
| ou\_ids | List of OUs where the Cloud Rule will be applied. | `set(number)` | `null` | no |
| owner\_user\_group\_ids | List of groups that own the Cloud Rule. | `set(number)` | `null` | no |
| owner\_user\_ids | List of users that own the Cloud Rule. | `set(number)` | `null` | no |
| post\_webhook\_id | ID of a post-rule webhook to attach to the Cloud Rule. | `number` | `null` | no |
| pre\_webhook\_id | ID of a pre-rule webhook to attach to the Cloud Rule. | `number` | `null` | no |
| project\_ids | List of projects where the Cloud Rule will be applied. | `set(number)` | `null` | no |
| service\_control\_policy\_ids | List of Service Control Policies attached to the Cloud Rule | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| built\_in | built\_in of the kion\_cloud\_rule. |
| id | id of the kion\_cloud\_rule. |
<!-- END_TF_DOCS -->
