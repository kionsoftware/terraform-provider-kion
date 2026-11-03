# terraform-kion-azure-policy

Terraform module for `kion_azure_policy`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "azure_policy" {
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
| [kion_azure_policy.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/azure_policy) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| azure\_policy | AzurePolicyDefinitionCreate represents an create Azure Policy Definition, complete with the policy itself and its parameters. | `object({ description = optional(string), name = string, parameters = optional(string), policy = string })` | `null` | no |
| owner\_user\_groups | List of user group IDs that will be owners of the azure policy. | `set(number)` | `null` | no |
| owner\_users | List of user IDs that will be owners of the azure policy. | `set(number)` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_azure\_policy. |
<!-- END_TF_DOCS -->
