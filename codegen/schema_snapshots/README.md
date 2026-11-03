# Provider schema snapshots

Authoritative `terraform providers schema -json` output for both provider major
versions, used by `kgen migrate-diff` and the state-upgrade codegen as ground
truth for the old→new attribute delta.

- `old.json`: the shipped SDKv2 provider (registry.terraform.io/kionsoftware/kion, v0 state).
- `new.json`: the Plugin Framework provider (this repo).

## Refresh

`new.json` tracks this repo. `make schema-snapshot` rewrites it from the
provider the tree builds, and the `schema-snapshot` CI job
(`make schema-snapshot-check`) fails when it is stale, so a schema change cannot
leave the checks that read it passing against an old file. After refreshing,
regenerate what is derived from it:

    go test ./internal/kgen/migrate/ -run TestRemovedAttrsAreCurrent -update

`old.json` is fixed: 0.3.36 is the final 0.3.x release, and this is its
schema. To rebuild it, point a `dev_overrides` CLI config at that binary (named
`terraform-provider-kion` in the override dir) and run
`terraform providers schema -json > old.json`.
