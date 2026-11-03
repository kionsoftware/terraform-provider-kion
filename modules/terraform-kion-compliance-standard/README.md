# terraform-kion-compliance-standard

Terraform module for `kion_compliance_standard`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "compliance_standard" {
  source = "..."

  created_by_user_id = 1
  name               = "example"
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
| [kion_compliance_standard.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/compliance_standard) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| created\_by\_user\_id | CreatedByUserID refers to the User in the application who created the Compliance Standard | `number` | n/a | yes |
| name | Name of the Compliance Standard in the application. | `string` | n/a | yes |
| cloud\_rule\_id | Cloud Rule ID the Compliance Standard should be added to | `number` | `null` | no |
| compliance\_check\_ids | List of compliance checks associated with the compliance standard | `set(number)` | `null` | no |
| description | Description for the Compliance Standard in the application. | `string` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the Compliance Standard. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the Compliance Standard. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| created\_at | created\_at of the kion\_compliance\_standard. |
| ct\_managed | ct\_managed of the kion\_compliance\_standard. |
| id | id of the kion\_compliance\_standard. |
<!-- END_TF_DOCS -->
