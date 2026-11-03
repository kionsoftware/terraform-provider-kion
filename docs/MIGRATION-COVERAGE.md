# Migration coverage map

The complete old (SDKv2) → new (Plugin Framework) migration surface for the 33
shared resource types, and exactly how each is handled. This is enforced by the
tests in `internal/kgen/migrate` (the classifications here are not prose; they
correspond to guards that fail the build if reality drifts) and exercised
end-to-end by the harness in `../terraform-coverage-test/migrate`.

## How migration works (recap)

- **State** migrates automatically in-provider: resources with a schema change
  carry a generated `UpgradeState` upgrader (`SchemaVersion 0→1`) that runs on
  `terraform plan` after the version bump. Restructured attributes are set to
  null and the provider's next Read repopulates them from the Kion API, so the
  state path is fail-safe.
- **Config** (`.tf`) is rewritten by `kmigrate`: attribute renames, block→id-list
  projections, obsolete-attribute drops, block→nested-attribute conversions, and
  attribute→nested-object folds.

## Coverage categories

### Fully automated: state upgrader + kmigrate

Ownership/membership blocks → id lists, association blocks → id lists, and the
attendant renames/drops. State upgrades transparently; `kmigrate` rewrites config.

`user_group`, `ou`, `funding_source`, `service_control_policy`, `azure_role`,
`gcp_iam_role`, `azure_arm_template`, `compliance_check`, `compliance_standard`,
`ou_cloud_access_role`, `project_cloud_access_role`, `cloud_rule`, `azure_policy`
(owners + policy-body fold), `project` (owners + `project_funding`/`budget`/
`move_ou_settings` block→attr), `custom_variable`, `webhook` (set→list),
`aws_account` (id string→number, block→single unwrap).

### Config drops (obsolete attributes)

Old settable attributes the new schema no longer accepts. `kmigrate` strips
them (they would otherwise fail `terraform plan`). Chiefly `last_updated`
everywhere, plus `gcp_iam_role.system_managed_policy`, `funding_source.labels`,
`gcp_account.create_mode`, `custom_account.skip_access_checking`, and the
`app_config` toggles. Enforced by `TestConfigDropsAreSettableAndRemoved` and
`TestCompleteness_settableDropsAreDocumented`.

### Read-only drops (attributes that became computed)

The other way an attribute stops being writable: the new schema still declares
it, but only as computed, so config that sets it fails with *Invalid
Configuration for Read-Only Attribute*. `kmigrate` strips these too, via
`ReadOnlyDrops`, a sibling table to `ConfigDrops` with the opposite guard
(present-and-computed rather than absent). Currently one attribute:
`project_note.create_user_id`, which the new resource takes from the
authenticated caller instead of the config. Every resource's SDKv2-injected
top-level `id` is excluded by design. Enforced both ways round by
`TestReadOnlyDropsArePresentAndComputed`.

### Cloud accounts: migrate by import (Layer 3)

`aws_account`, `azure_account`, `gcp_account`, `custom_account` link real cloud
accounts. `azure/gcp/custom` also renamed `name → account_name` (handled by the
state upgrader). Because they can't be freely recreated, the harness migrates
their *state* via `terraform import`
(`../terraform-coverage-test/migrate/import`) rather than a state upgrade. Their
*config* is still rewritten by `kmigrate` like any other resource's. The
`name → account_name` rename plus the obsolete-attribute drops above.

### Alias resources: fully automated (state + config)

`kion_aws_iam_policy` and `kion_aws_cloudformation_template` are back-compat
aliases the new provider kept, renamed to `kion_iam_policy` and `kion_cft`. Their
alias resource struct **embeds** the primary (`awsIamPolicyResource` embeds
`iam_policyResource`), so:

- **State**: the upgrader is generated on the primary (`iam_policy` / `cft`
  package) via a `state_upgrades.yaml` entry keyed by the new primary type with
  `old_type:` set to the alias. Go embedding promotes `UpgradeState` to the
  alias, so old-name state (schema version 0) upgrades. The shared schema version
  is bumped to 1. `TestAliasInheritsUpgradeState` proves this fires on the alias
  receiver at runtime (owners project, scalars pass through).
- **Config**: `kmigrate` matches the alias's old type name (via `old_type`) and
  rewrites the still-in-use `resource "kion_aws_iam_policy"` blocks.

Enforced by `TestAliasResources_haveGeneratedUpgraders`.

### Benign: no upgrader needed

Resources whose only changes are added optional attributes or dropped
computed-only fields; Terraform reconciles these natively. e.g. `project_note`,
`project_enforcement`, `label`, `user`, `app_config` (beyond its dropped toggles).

## Enforcement

| Guard | What it prevents |
|---|---|
| `TestDrift_everyStructuralChangeIsCovered` | a structural change with no upgrader |
| `TestUpgrades_targetsAndSourcesExist` | a projection into a non-existent attr (the azure_policy bug) |
| `TestCompleteness_ownershipBlocksAreProjected` | an unprojected ownership block (the aws_iam_policy miss) |
| `TestCompleteness_settableDropsAreDocumented` | a new, unreviewed config break |
| `TestConfigDropsAreSettableAndRemoved` | kmigrate dropping a still-valid attribute |
| `TestReadOnlyDropsArePresentAndComputed` | an attribute that became read-only left in config |
| `TestAliasResources_haveGeneratedUpgraders` | an alias losing its (inherited) state upgrader |
| `TestAliasInheritsUpgradeState` | the alias not actually inheriting/running the upgrader |
| `TestUpgradeState_decodes` (×22) | an upgrader producing state that won't decode |
| migratehelper unit tests | a transform primitive returning wrong values |
