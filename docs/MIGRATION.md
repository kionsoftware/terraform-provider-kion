# Migrating to the new Kion provider

The new `kion` provider (Plugin Framework) ships at the **same registry address**
(`kionsoftware/kion`) as a new **major version**. Your existing Terraform state
migrates **automatically**. The provider carries state upgraders that transform
old-schema state to the new schema the first time you plan, the same mechanism the
AWS provider uses across major versions. There is no separate state-rewriting
tool to run, and nothing is changed until you review and apply.

## Steps

1. **Bump the provider version** in your configuration:

   ```hcl
   terraform {
     required_providers {
       kion = {
         source  = "kionsoftware/kion"
         version = ">= 1.0.0" # the new major version
       }
     }
   }
   ```

2. **Update your configuration** to the new attribute names (see "Config changes"
   below). `kmigrate` does this for you.

   Download it from the release matching the provider version you are moving to.
   It is a single binary with no runtime or dependencies, and it carries the
   rewrite rules inside it, so there is nothing else to fetch:

   ```sh
   # macOS on Apple silicon; substitute your platform
   # (darwin/linux/windows x amd64/arm64)
   V=1.1.1
   curl -fsSLO "https://github.com/kionsoftware/terraform-provider-kion/releases/download/v${V}/kmigrate_${V}_darwin_arm64.zip"
   unzip -j "kmigrate_${V}_darwin_arm64.zip" kmigrate && chmod +x kmigrate
   ```

   Then, from the directory holding your `.tf` files:

   ```sh
   ./kmigrate --check ./   # preview the .tf changes
   ./kmigrate ./           # apply them
   ```

   `kmigrate` rewrites `.tf` files in place, leaving comments, formatting and
   everything it does not recognise untouched. It reports any block it cannot
   finish. `kion_user` gains five required attributes that have no old-provider
   equivalent, so it names those and leaves them for you.

   Commit or diff the result before moving on.

3. **Initialize and plan:**

   ```sh
   terraform init -upgrade
   terraform plan
   ```

   The provider upgrades each resource’s state in memory as it reads it. Review
   the plan. A clean migration shows **no changes** (or only the benign one-time
   diffs noted below). **Nothing is modified until you apply.**

4. **Apply** once the plan looks right:

   ```sh
   terraform apply
   ```

## Config changes

The state upgraders fix *state*; your `.tf` *configuration* references some
attributes by their old names, which the new provider rejects. `kmigrate` rewrites
them. The main patterns:

- Ownership/membership blocks became id lists:
  `owner_users { id = 5 }` → `owner_user_ids = [5]`; likewise `owner_groups` →
  `owner_user_group_ids`, `users` → `user_ids`, `user_groups` → `user_group_ids`,
  `accounts` → `account_ids`.
- `cloud_rule` association blocks became id lists: `aws_iam_policies { id = 1 }` →
  `iam_policy_ids = [1]`, and similarly for `cft`s, GCP IAM roles, Azure role
  definitions, compliance standards, OUs, projects, SCPs, etc.
- Account `name` → `account_name` (`azure`/`gcp`/`custom` accounts).
- `project`'s `project_funding` / `budget` / `move_ou_settings` blocks became
  nested attributes: `project_funding { … }` → `project_funding = [{ … }]`.
- `azure_policy`'s `name`/`description`/`policy`/`parameters` were folded into a
  nested `azure_policy = { … }` object.
- `aws_cloudformation_template`'s `tags` went from a map to a list of objects:
  `tags = { env = "prod" }` → `tags = [{ tag_key = "env", tag_value = "prod" }]`.
- `ou_cloud_access_role` and `project_cloud_access_role`'s `aws_session_tags`
  (added in 0.3.36) made the same change:
  `aws_session_tags = { Team = "platform" }` →
  `aws_session_tags = [{ tag_key = "Team", tag_value = "platform" }]`.
- Obsolete attributes the new schema dropped (e.g. `last_updated`,
  `gcp_iam_role.system_managed_policy`) are removed.

All of the block conversions above also handle the `dynamic` form, which is how
most configurations generate a repeatable block. An attribute has no `dynamic`
equivalent, so the whole construct collapses into a for expression over the same
`for_each`:

```hcl
# before
dynamic "owner_users" {
  for_each = var.owner_user_ids
  content {
    id = owner_users.value
  }
}

# after
owner_user_ids = [for owner_users in var.owner_user_ids : owner_users]
```

A resource that mixes literal and dynamic blocks of the same name gets a
`concat(…)` of the two.

`kmigrate` reports every change it makes, and prints anything it could not
convert, a dynamic block missing the field the projection reads for instance,
as a follow-up rather than guessing. Run `terraform fmt` afterward.

The back-compat aliases `kion_aws_iam_policy` and `kion_aws_cloudformation_template`
keep working under the new provider. Both their state and config migrate
automatically (the alias inherits the primary resource's upgrader), so you need
not rename them.

## Benign one-time diffs

A few resources may show a small, safe diff on the first plan (state is corrected
on the next refresh from the Kion API):

- `kion_aws_account`: `id` changed from a string to a number; the value is
  identical.
- Some computed fields that the old provider stored but the new one derives
  (e.g. `last_updated`) are dropped and repopulated by Read.
- `kion_aws_cloudformation_template`: `regions` changed from a set to a list; if
  ordering shifts, it is cosmetic. Its `tags` attribute changed shape from a map
  to a list of objects (see the config changes above); `kmigrate` rewrites the
  config for you and the state upgrader explodes the stored map into the same
  object list. Tags arrive in sorted key order, the order the old state itself
  recorded, and the next refresh replaces it with the API's order, so any
  reordering you see on the first plan is cosmetic.
- `kion_ou_cloud_access_role` / `kion_project_cloud_access_role`:
  `aws_session_tags` is exploded the same way. The tag objects' `id` fields
  start null and the next refresh fills them.

## What migrates automatically

Every resource whose schema changed structurally (ownership/membership/id-list
restructures, `aws_account`’s id, single-nested-block changes) carries a
generated state upgrader (`internal/service/<name>/<name>_upgrade_gen.go`),
driven by `codegen/state_upgrades.yaml`. A CI drift test
(`internal/kgen/migrate`) fails the build if any resource gains a structural
change without a corresponding upgrader, so this coverage cannot silently
regress.

Resources whose only change is an added or removed **scalar** attribute need no
upgrader. Terraform reconciles those automatically (missing → null → Read
fills; extra → dropped).
