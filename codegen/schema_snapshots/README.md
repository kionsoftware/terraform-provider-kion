# Provider schema snapshots

Authoritative `terraform providers schema -json` output for both provider major
versions, used by `kgen migrate-diff` and the state-upgrade codegen as ground
truth for the old→new attribute delta.

- `old.json`: the shipped SDKv2 provider (registry.terraform.io/kionsoftware/kion, v0 state).
- `new.json`: the Plugin Framework provider (this repo).

## Refresh

Build both provider binaries, then for each, with a `dev_overrides` CLI config
pointing `kionsoftware/kion` at the binary's dir:

    terraform providers schema -json > <old|new>.json

(named `terraform-provider-kion` in the override dir).
