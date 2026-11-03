---
subcategory: ""
layout: "kion"
page_title: "Kion: {{ .ProviderResourceName }}"
description: |-
  Manages a Kion {{ .HumanResourceName }}.
---

{{- if .IncludeComments }}
<!---
Documentation guidelines:
- Begin resource descriptions with "Manages..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->
{{- end }}

# Resource: {{ .ProviderResourceName }}

Manages a Kion {{ .HumanResourceName }}.

## Example Usage

### Basic Usage

```terraform
resource "{{ .ProviderResourceName }}" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the {{ .HumanResourceName }}.
* `example_attribute` - Brief description of the attribute.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import {{ .HumanResourceName }} using its ID. For example:

```terraform
import {
  to = {{ .ProviderResourceName }}.example
  id = "12345"
}
```

Using `terraform import`, import {{ .HumanResourceName }} using the ID. For example:

```console
% terraform import {{ .ProviderResourceName }}.example 12345
```
