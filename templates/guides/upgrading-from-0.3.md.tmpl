---
page_title: "Upgrading from 0.3.x"
subcategory: ""
description: |-
  Move a configuration and its state from the 0.3.x Kion provider to the new major version.
---

# Upgrading from 0.3.x

The new `kion` provider (Plugin Framework) ships at the **same registry address**
(`kionsoftware/kion`) as a new **major version**. Your existing Terraform state
migrates **automatically**. The provider carries state upgraders that transform
old-schema state to the new schema the first time you plan, the same mechanism the
AWS provider uses across major versions. There is no separate state-rewriting
tool to run, and nothing is changed until you review and apply.

**0.3.36 is the final 0.3.x release.** The 0.3.x line gets no further fixes,
security fixes included, and has no maintenance branch. If you need a fix,
upgrade to the new major version.

## Requirements

- **Terraform 1.4 or later.** The new provider speaks plugin protocol 6, which
  Terraform 0.x cannot load, and before 1.4 Terraform plans the computed
  fields of a nested list (a cloud access role's session tag ids, for example)
  as unknown on every run, so those resources never show `No changes`.
- Upgrade from **0.3.36**, the final 0.3.x release, if you can. The state
  upgraders are written against it. An attribute that a later 0.3.x added and your
  version lacks simply starts empty.

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
   # Set V to the provider release you are moving to. macOS on Apple silicon
   # shown; substitute your platform (darwin/linux/windows x amd64/arm64).
   V=X.Y.Z
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
   finish, such as an attribute the new schema requires that it has no value
   for (see "Breaking changes" below), and leaves those for you.

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
  (added in 0.3.35) made the same change:
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
as a follow-up rather than guessing. It also lists every reference, in an
output, a local or anywhere else, to an attribute the new schema dropped, such
as `kion_azure_role.this.last_updated` or `data.kion_account.a.id`, naming the
replacement when a resource attribute was renamed. Run `terraform fmt` afterward.

The back-compat aliases `kion_aws_iam_policy` and `kion_aws_cloudformation_template`
keep working under the new provider. Both their state and config migrate
automatically (the alias inherits the primary resource's upgrader), so you need
not rename them.

## Breaking changes

These need a decision or a value from you. `kmigrate` reports each one it can
see in a resource block, and `terraform plan` fails on any you miss.

- **`kion_user` is a full resource.** The old one could only adopt an existing
  user. The new one requires `email`, `first_name`, `idms_id`, `last_name` and
  `username`.
- **`kion_azure_account`, `kion_custom_account` and `kion_gcp_account` require
  `project_id` and `start_datecode`.** The old provider placed an account with no
  `project_id` in the account cache. The new resources always manage an
  account in a project, and an account in the cache is a `kion_account_cache`.
  **Don't add `project_id` to a cached account**, because that would move it into
  the project. Move it to `kion_account_cache` instead:

  ```hcl
  # Terraform 1.7+. On older versions, run
  # `terraform state rm kion_custom_account.example` instead.
  removed {
    from = kion_custom_account.example
    lifecycle {
      destroy = false
    }
  }

  import {
    to = kion_account_cache.example
    id = "123" # the account cache ID
  }

  resource "kion_account_cache" "example" {
    account_name = "example"
    payer_id     = 1
  }
  ```

- **`data "kion_aws_iam_policy"`** no longer takes `query`, `policy_type`,
  `page` or `page_size`. Match with a `filter` block instead, for example
  `filter { name = "name", values = ["ReadOnly"] }`. `kmigrate` does not rewrite
  data sources.
- **`data "kion_project_enforcement"`** requires `project_id`. The old data
  source sent its request with the project placeholder unfilled, so it never
  returned results. Set `project_id` to the project to read.

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
- `kion_project_cloud_access_role`: `apply_to_all_accounts = true -> false`.
  Kion's create stores `true` whatever it is sent, and 0.3.x never corrected
  it, so the old state recorded `true`. The first apply sets the value you
  configured, and the new provider applies it on create too. This changes the
  role's account scope in Kion, so check the configured value before applying.
- Empty collections that the old provider stored as null, such as
  `regions = []` on `kion_compliance_check` or `owner_user_group_ids = []` on
  `kion_webhook` and `kion_custom_variable`, show `+ ... = []` once.

## What migrates automatically

Every resource whose schema changed structurally (ownership/membership/id-list
restructures, `aws_account`’s id, single-nested-block changes) carries a
generated state upgrader, and a CI test fails the build if any resource gains a
structural change without one. The [migration coverage
record](https://github.com/kionsoftware/terraform-provider-kion/blob/main/docs/MIGRATION-COVERAGE.md)
lists every resource and how it migrates.

Resources whose only change is an added or removed **scalar** attribute need no
upgrader. Terraform reconciles those automatically (missing → null → Read
fills; extra → dropped).
