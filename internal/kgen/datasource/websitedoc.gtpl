---
subcategory: ""
layout: "kion"
page_title: "Kion: {{ .ProviderResourceName }}"
description: |-
  Provides details about a Kion {{ .HumanDataSourceName }}.
---

{{- if .IncludeComments }}
<!---
Documentation guidelines:
- Begin data source descriptions with "Provides details about..."
- Use simple language and avoid jargon
- Focus on brevity and clarity
- Use present tense and active voice
- Don't begin argument/attribute descriptions with "An", "The", "Defines", "Indicates", or "Specifies"
- Boolean arguments should begin with "Whether to"
- Use "example" instead of "test" in examples
--->
{{- end }}

# Data Source: {{ .ProviderResourceName }}

Provides details about a Kion {{ .HumanDataSourceName }}.

## Example Usage

### Basic Usage

```terraform
data "{{ .ProviderResourceName }}" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Brief description of the required argument.

The following arguments are optional:

* `optional_arg` - (Optional) Brief description of the optional argument.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - The ID of the {{ .HumanDataSourceName }}.
* `example_attribute` - Brief description of the attribute.
