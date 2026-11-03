# terraform-kion-compliance-check

Terraform module for `kion_compliance_check`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "compliance_check" {
  source = "..."

  cloud_provider_id        = 1
  compliance_check_type_id = 1
  name                     = "example"
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
| [kion_compliance_check.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/compliance_check) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| cloud\_provider\_id | CloudProviderID for the Compliance Check in the application. Must be "2" for Azure Policy checks. 1 - AWS. 2 - Microsoft. | `number` | n/a | yes |
| compliance\_check\_type\_id | ComplianceCheckTypeID for the Compliance Check in the application. 1 - External. These checks are not triggered by cloudtamer.io, but rather can display findings from checks run outside of cloudtamer.io. 2 - Cloud Custodian. These checks are run by cloudtamer.io using Cloud Custodian. 3 - Azure Policy. These checks are scraped from Azure's policy reporting engine and apply Azure Policies to accounts. | `number` | n/a | yes |
| name | Name of the Compliance Check in the application. | `string` | n/a | yes |
| azure\_policy\_id | AzurePolicyID refers to the ID of the Azure Policy in cloudtamer.io which will be deployed and scraped to determine compliance status of Azure resources. Should only be provided for Azure Policy checks. | `number` | `null` | no |
| body | TODO: add a valid example here Body of the Compliance Check defining what actions will be run. | `string` | `null` | no |
| compliance\_control\_ids | List of Compliance Control IDs to which this Compliance Check is linked. | `set(number)` | `null` | no |
| compliance\_standard\_id | ComplianceStandardID is an optional field for which compliance standard the created check should be associated with. | `number` | `null` | no |
| created\_by\_user\_id | CreatedByUserID refers to the User in the application who created the Compliance Check. Will be the requesting User's ID if not specified. | `number` | `null` | no |
| description | Description for the Compliance Check in the application. | `string` | `null` | no |
| frequency\_minutes | FrequencyMinutes defines how often the check will be run. Is required if check type is Cloud Custodian or Azure Policy. | `number` | `null` | no |
| frequency\_type\_id | FrequencyTypeID refers to the duration type of the frequency it will be checked. Is required if check type is Cloud Custodian or Azure Policy. 2 - minutes 3 - hours 4 - days | `number` | `null` | no |
| is\_all\_regions | IsAllRegions determines if the check should be applied to all regions applied on the system | `bool` | `null` | no |
| is\_auto\_archived | IsAutoArchived defines whether existing findings should be archived before new findings are reported. | `bool` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the Compliance Check. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the Compliance Check. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |
| regions | List of the AWS regions where the compliance check applies. Required when check type id is cloud custodian. | `set(string)` | `null` | no |
| severity\_type\_id | SeverityTypeID for the severity level of the compliance check 1 - Informational. 2 - Low. 3 - Medium (Default). 4 - High. | `number` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| created\_at | created\_at of the kion\_compliance\_check. |
| ct\_managed | ct\_managed of the kion\_compliance\_check. |
| id | id of the kion\_compliance\_check. |
| last\_scan\_id | last\_scan\_id of the kion\_compliance\_check. |
<!-- END_TF_DOCS -->
