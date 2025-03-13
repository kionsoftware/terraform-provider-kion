# Kion Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/kionsoftware/kion/latest)
[![Documentation](https://img.shields.io/badge/documentation-blue.svg)](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs)

The Kion Terraform Provider enables you to manage your Kion resources using HashiCorp's [Terraform](https://www.terraform.io) infrastructure as code tool. This provider supports creating, updating, reading, and deleting resources through Terraform, as well as querying existing resources using filters.

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Authentication](#authentication)
- [Examples](#examples)
- [Resource State Management](#resource-state-management)
- [Contributing](#contributing)

## Quick Start

```hcl
# Configure the Kion Provider
terraform {
  required_providers {
    kion = {
      source  = "kionsoftware/kion"
      # Kion recommends pinning the provider to a specific version, though usually you want to use the latest one.
      # version = "x.x.x"
    }
  }
}

provider "kion" {
  # If these are commented out, they will be loaded from environment variables.  To load them from the environment variables, be sure you're prefixing the exported environment variable correctly with TF_VAR
  # kion_url    = "https://kion.example.com"
  # kion_apikey = "key here"
}

# Create an IAM policy
resource "kion_aws_iam_policy" "example" {
  name        = "example-policy"
  description = "Example IAM Policy"
  policy      = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "*"
      Resource  = "*"
    }]
  })
}
```

## Installation

1. Install [Terraform](https://developer.hashicorp.com/terraform/downloads) (1.0+)
2. Create a new directory for your Terraform configuration
3. Create a `providers.tf` file with the following content:

```hcl
terraform {
  required_providers {
    kion = {
      source    = "kionsoftware/kion"
      # Recommended: Pin to a specific version
      # version = "x.x.x"
    }
  }
}

provider "kion" {
  # If these are commented out, they will be loaded from environment variables.  To load them from the environment variables, be sure you're prefixing the exported environment variable correctly with TF_VAR
  # kion_url    = "https://kion.example.com"
  # kion_apikey = "key here"
}
```

4. Initialize Terraform:

```bash
terraform init
```

## Authentication

You can authenticate with Kion using either environment variables or provider configuration:

### Environment Variables

```bash
export TF_VAR_kion_url="https://kion.example.com"
export TF_VAR_kion_apikey="your-api-key"
```

### Provider Configuration

```hcl
provider "kion" {
  kion_url    = "https://kion.example.com"
  kion_apikey = "your-api-key"
}
```

## Examples

For a complete list of available resources and data sources, please refer to our [provider documentation](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs).

## Resource State Management

The provider supports importing existing Kion resources into Terraform state. This allows you to manage existing resources without recreating them.

To import a resource:

1. Create the resource configuration in your Terraform files
2. Import the resource state:

```bash
terraform import [resource_type].[resource_name] [resource_id]
```

Example:

```bash
terraform import kion_aws_cloudformation_template.audit_logging 20
```

For a complete list of available data sources, please refer to our [provider documentation](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs).

## Contributing

For repository maintainers pushing to the Terraform Registry:

1. Update the version in the Makefile
2. Create and push a new tag:

```bash
git tag -a vX.Y.Z -m "Description of changes"
git push origin vX.Y.Z
```
