# terraform-kion-compliance-control

Terraform module for `kion_compliance_control`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "compliance_control" {
  source = "..."

  program_id = 1
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
| [kion_compliance_control.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/compliance_control) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| program\_id | ID of the compliance program the control belongs to (used to scope delete). | `number` | n/a | yes |
| arm\_template\_definition\_ids | ARMTemplateDefinitionIDs is a set of ARM template ids applied to the control. | `set(number)` | `null` | no |
| aws\_cloudformation\_policy\_ids | AWSCloudformationPolicyIDs is a set of AWS cloudformation policy ids applied to the control. | `set(number)` | `null` | no |
| azure\_policy\_definition\_ids | AzurePolicyDefinitionIDs is a set of Azure policy ids applied to the control. | `set(number)` | `null` | no |
| cloud\_provider\_policy\_ids | CloudProviderPolicyIDs is a set of cloud provider policy ids applied to the control. | `set(number)` | `null` | no |
| compliance\_check\_ids | ComplianceCheckIDs is a set of compliance check ids applied to the control. | `set(number)` | `null` | no |
| compliance\_family\_id | Description for the compliance control in the application. | `number` | `null` | no |
| compliance\_levels | ComplianceLevels is a set of levels to which the control belongs. | `set(number)` | `null` | no |
| control\_number | ControlNumber of the compliance control in the application. | `number` | `null` | no |
| description | Description for the compliance control in the application. | `string` | `null` | no |
| name | Name of the compliance control in the application. | `string` | `null` | no |
| severity | Severity for the compliance control in the application. | `string` | `null` | no |
| title | Title for the compliance control in the application. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_compliance\_control. |
<!-- END_TF_DOCS -->
