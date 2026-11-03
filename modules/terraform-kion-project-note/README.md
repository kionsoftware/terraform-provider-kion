# terraform-kion-project-note

Terraform module for `kion_project_note`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "project_note" {
  source = "..."

  create_user_id = 1
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
| [kion_project_note.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/project_note) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| create\_user\_id | ID of the user who created the note. | `number` | n/a | yes |
| name | Name of the note. | `string` | `null` | no |
| project\_id | ID of the project. | `number` | `null` | no |
| text | Body text of the note. | `string` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| create\_user\_name | create\_user\_name of the kion\_project\_note. |
| created\_at | created\_at of the kion\_project\_note. |
| id | id of the kion\_project\_note. |
| last\_update\_user\_id | last\_update\_user\_id of the kion\_project\_note. |
| last\_update\_user\_name | last\_update\_user\_name of the kion\_project\_note. |
| updated\_at | updated\_at of the kion\_project\_note. |
<!-- END_TF_DOCS -->
