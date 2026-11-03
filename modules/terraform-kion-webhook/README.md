# terraform-kion-webhook

Terraform module for `kion_webhook`, generated from the provider schema by
`kgen module`. Do not edit by hand -- regenerate instead.

## Usage

```hcl
module "webhook" {
  source = "..."

  callout_url        = "https://example.com"
  name               = "example"
  timeout_in_seconds = 1
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
| [kion_webhook.this](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs/resources/webhook) | resource |

## Inputs

| Name | Description | Type | Default | Required |
| ---- | ----------- | ---- | ------- | :------: |
| callout\_url | URL the application will call when the webhook is triggered. | `string` | n/a | yes |
| name | Name of the webhook in the application. | `string` | n/a | yes |
| timeout\_in\_seconds | The number of seconds the application will wait before considering the webhook "timed out". | `number` | n/a | yes |
| description | Description for the webhook in the application. | `string` | `null` | no |
| owner\_user\_group\_ids | List of groups IDs who will own the webhook. Is required if no owner user IDs are listed. | `set(number)` | `null` | no |
| owner\_user\_ids | List of user IDs who will own the webhook. Is required if no owner group IDs are listed. | `set(number)` | `null` | no |
| request\_body | HTTP request payload body to use when the webhook is triggered. | `string` | `null` | no |
| request\_headers | HTTP headers to use when the webhook is triggered. | `string` | `null` | no |
| request\_method | HTTP request method to use when the webhook is triggered. | `string` | `null` | no |
| should\_send\_secure\_info | Will be true when the request to the webhook will include temporary AWS access keys. | `bool` | `null` | no |
| skip\_ssl | Will be true when the request to the webhook will skip certificate verification. | `bool` | `null` | no |
| use\_request\_headers | Will be true when intending to send the request headers specified above. | `bool` | `null` | no |

## Outputs

| Name | Description |
| ---- | ----------- |
| id | id of the kion\_webhook. |
<!-- END_TF_DOCS -->
