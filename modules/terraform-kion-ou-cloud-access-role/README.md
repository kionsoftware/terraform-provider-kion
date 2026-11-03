# terraform-kion-ou-cloud-access-role

Terraform module for `kion_ou_cloud_access_role`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "ou_cloud_access_role" {
  source = "..."

  name  = "example"
  ou_id = 1
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
| [kion_ou_cloud_access_role.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/ou_cloud_access_role) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| name | Name of the cloud access role in the application. | `string` | n/a | yes |
| ou\_id | ID of the OU where the cloud access role is attached. | `number` | n/a | yes |
| aws\_create\_instance\_profile | If true, an IAM instance profile is created for this role. Applies only to non-User CAR types (Custom Trust, Account, Service). Will default to false if not set. | `bool` | `null` | no |
| aws\_iam\_path | Text of the IAM Path in AWS to be stored in AWS. | `string` | `null` | no |
| aws\_iam\_permissions\_boundary | ID of the AWS IAM policy to be used as a permissions boundary for this role. | `number` | `null` | no |
| aws\_iam\_policies | IDs of the AWS IAM policies attached to this role. | `set(number)` | `null` | no |
| aws\_iam\_role\_name | AWS IAM role name corresponding to the cloud access role. | `string` | `null` | no |
| aws\_iam\_role\_trust\_policy | AWS IAM role trust policy JSON. Required when cloud\_access\_role\_type\_id = 2 (Custom Trust). Rejected for other types. | `string` | `null` | no |
| aws\_partition | AWS partition a non-User role syncs to. Defaults to aws when omitted. | `string` | `null` | no |
| aws\_session\_tags | AWS Session Tags used in this role when accessing the AWS console. | `list(object({ cloud_access_role_id = optional(number), id = optional(number), ou_cloud_access_role_id = optional(number), tag_key = optional(string), tag_value = optional(string) }))` | `null` | no |
| aws\_trusted\_account\_numbers | AWS account IDs this role trusts. Required when cloud\_access\_role\_type\_id = 3 (Account). Currently supports only one 12-digit AWS account number. | `list(string)` | `null` | no |
| aws\_trusted\_services | AWS service principals this role trusts. Required when cloud\_access\_role\_type\_id = 4 (Service). | `list(string)` | `null` | no |
| azure\_role\_definitions | IDs of the Azure Role Definitions attached to this role. | `set(number)` | `null` | no |
| cloud\_access\_role\_type\_id | Type of the cloud access role. 1 User (default), 2 Custom Trust, 3 Account, 4 Service. Non-User types are AWS only. Changing this forces a new role: the update body carries no role type, so the API would accept the change and ignore it. | `number` | `null` | no |
| gcp\_iam\_roles | IDs of the GCP IAM roles attached to this role. | `set(number)` | `null` | no |
| long\_term\_access\_keys | If long term access is true, users of this cloud access role can generate aws long-term access keys. Will default to false if not set. | `bool` | `null` | no |
| short\_term\_access\_keys | If short term access is true, users of this cloud access role can generate short-term access keys. Will default to false if not set. | `bool` | `null` | no |
| user\_group\_ids | IDs of the user groups allowed to use this role to access the AWS console. | `set(number)` | `null` | no |
| user\_ids | IDs of the users allowed to use this role to access the AWS console. | `set(number)` | `null` | no |
| web\_access | If web access is true, users of this cloud access role can log into the console Will default to false if not set. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_ou\_cloud\_access\_role. |
<!-- END_TF_DOCS -->
