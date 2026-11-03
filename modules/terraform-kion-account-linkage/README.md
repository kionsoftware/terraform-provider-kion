# terraform-kion-account-linkage

Terraform module for `kion_account_linkage`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "account_linkage" {
  source = "..."

  azure_object_id      = "example"
  azure_principal_name = "example"
  payer_id             = 1
  user_id              = 1
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
| [kion_account_linkage.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/account_linkage) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| azure\_object\_id | Azure Object ID of the user (UUID format). | `string` | n/a | yes |
| azure\_principal\_name | Azure Principal Name (UPN) of the user. | `string` | n/a | yes |
| payer\_id | ID of the payer this linkage is associated with. | `number` | n/a | yes |
| user\_id | ID of the Kion user to link. | `number` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_account\_linkage. |
<!-- END_TF_DOCS -->
