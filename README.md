# Kion Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/kionsoftware/kion/latest)
[![Documentation](https://img.shields.io/badge/documentation-blue.svg)](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs)

The Kion Terraform Provider lets you manage [Kion](https://kion.io) with
HashiCorp's [Terraform](https://www.terraform.io): organizational units,
projects, cloud access, compliance, funding and more. You can create, update,
read and delete resources, and look up existing ones with data sources.

## Table of Contents

- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [Importing an existing install](#importing-an-existing-install)
- [Upgrading from 0.3.x](#upgrading-from-03x)
- [Kion version compatibility](#kion-version-compatibility)
- [Documentation](#documentation)
- [Contributing](#contributing)

## Quick Start

Requires [Terraform](https://developer.hashicorp.com/terraform/install) 1.4 or later.

```hcl
terraform {
  required_providers {
    kion = {
      source = "kionsoftware/kion"
      # Pinning a version is recommended:
      # version = "~> 1.8"
    }
  }
}

provider "kion" {
  # Or set KION_API_URL and KION_API_KEY in the environment.
  # api_url = "https://kion.example.com"
  # api_key = var.kion_api_key
}

resource "kion_label" "example" {
  key   = "team"
  value = "platform"
  color = "#2E86C1"
}
```

Then run `terraform init`.

## Authentication

The provider authenticates with a Kion app API key, created in Kion under your
user's App API Keys. Set it in the environment:

```bash
export KION_API_URL="https://kion.example.com"
export KION_API_KEY="app_1_XXXXXXXXXXXX"
```

or in the provider block, with `api_url` and `api_key`. Set
`KION_SKIP_SSL_VALIDATION=true` only for an install with a self-signed
certificate. The [provider documentation](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs)
lists every setting.

## Importing an existing install

To bring resources that already exist in Kion under Terraform, `kion-import`
(attached to every release) reads your install and writes Terraform `import`
blocks, and `terraform plan -generate-config-out` writes the matching
configuration. See the [importing guide](docs/guides/importing.md).

A single resource can also be imported by id:

```bash
terraform import kion_label.example 20
```

## Upgrading from 0.3.x

Version 1.0 is a new major version at the same registry address. **Your state
migrates automatically** the first time you plan. Your configuration needs some
attribute renames, which `kmigrate` (attached to every release) makes for you.
0.3.36 is the final 0.3.x release. See the
[upgrade guide](docs/guides/upgrading-from-0.3.md).

## Kion version compatibility

The provider's version is its own and does not track Kion's. It supports the
current Kion release and the three before it. Against an older install, a
resource the install's API can't serve fails with a clear error naming the
version it needs.

## Documentation

- [Provider documentation](https://registry.terraform.io/providers/kionsoftware/kion/latest/docs): every resource and data source, with examples
- [Importing an existing install](docs/guides/importing.md)
- [Upgrading from 0.3.x](docs/guides/upgrading-from-0.3.md)
- [Terraform modules](modules/): one per resource
- [Changelog](CHANGELOG.md)

## Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md). How the provider is built and
generated is in [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md), and the release
process is in [docs/RELEASING.md](docs/RELEASING.md).

## License

Mozilla Public License 2.0. See [LICENSE](LICENSE).
