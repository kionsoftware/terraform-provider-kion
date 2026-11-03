# terraform-kion-ou-note

Terraform module for `kion_ou_note`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "ou_note" {
  source = "..."

  create_user_id = 1
  name           = "example"
  ou_id          = 1
  text           = "example"
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
| [kion_ou_note.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/ou_note) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| create\_user\_id | ID of the user creating the note. | `number` | n/a | yes |
| name | Name of the OU note. | `string` | n/a | yes |
| ou\_id | ID of the OU this note belongs to. | `number` | n/a | yes |
| text | Text content of the OU note. | `string` | n/a | yes |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_ou\_note. |
<!-- END_TF_DOCS -->
