---
page_title: "Importing an existing Kion install"
subcategory: ""
description: |-
  Generate Terraform import blocks for everything in a live Kion install with kion-import.
---

# Importing an existing Kion install

`kion-import` reads a live Kion install and writes Terraform `import` blocks.

It deliberately does not write configuration or state. Every resource implements
`ImportState`, so `terraform plan -generate-config-out` produces the
configuration and `terraform apply` writes the state, both through the provider's
own Read. What Terraform cannot do is discover what exists in Kion, and that is
the only job this tool has.

It needs **Terraform 1.5 or later**, the first release with `import` blocks and
`terraform plan -generate-config-out`.

## Get it

`kion-import` ships in the release, as a single binary with no runtime or
dependencies. It carries the import manifest inside it, so there is nothing else
to fetch — and so a binary can never be paired with a manifest from a different
commit.

```sh
# Set V to the provider release you are using. macOS on Apple silicon
# shown; substitute your platform (darwin/linux/windows x amd64/arm64).
V=X.Y.Z
curl -fsSLO "https://github.com/kionsoftware/terraform-provider-kion/releases/download/v${V}/kion-import_${V}_darwin_arm64.zip"
unzip -j "kion-import_${V}_darwin_arm64.zip" kion-import && chmod +x kion-import
```

Releases up to and including v1.1.0 do not carry it; from a checkout of those,
build it with `go build -o bin/kion-import ./cmd/kion-import` (or `make
build-tools`, which builds it alongside the other dev tools).

```sh
export KION_URL=https://kion.example.com
export KION_APIKEY=…

kion-import --list-types                       # what can be imported
kion-import --probe                            # what exists, writing nothing
kion-import --out imports.tf                   # generate

terraform init
terraform plan -generate-config-out=generated.tf

kion-import rewrite-refs                       # literals -> references
terraform apply
```

To manage the result through the modules instead of the resources, add one
step after `rewrite-refs`:

```sh
make import-modules-check IMPORT_IN=generated.tf IMPORT_OUT=modules.tf
```

See *Rewrite into module calls*.

## Rewrite the literal ids into references

`terraform plan -generate-config-out` writes every foreign key as a bare
integer, because Terraform has no way to know the value points at anything:

```hcl
resource "kion_project" "kion_project_9" {
  ou_id = 11
}
```

That is accurate for the install it came from, and it is a poor configuration.
It does not express its own dependency graph, so `terraform destroy` followed by
a re-apply will not order correctly and recreating an OU will not cascade. Point
it at a **different** install and it is worse than poor: id 11 either does not
exist or, silently, belongs to something else entirely.

```sh
kion-import rewrite-refs --imports imports.tf --generated generated.tf
```

```hcl
resource "kion_project" "kion_project_9" {
  ou_id = kion_ou.kion_ou_11.id
}
```

Terraform now resolves the value from the id it actually created. **This is why
moving a configuration between installs needs no id map** — the dependency graph
is the id map, and it brings correct ordering and parallelism with it.

Both files are required. `generated.tf` contains no ids (they are Computed, so
`-generate-config-out` omits them), so the id-to-label index comes from the
import blocks, where `to` and `id` sit together.

Use `--dry-run` to see what would change, and `--out` to write elsewhere instead
of rewriting in place. Running it twice is a no-op.

### What is left alone, and why

**Attributes that are not references.** Not every `*_id` points at a Kion
record. `portfolio_id` is the portfolio's id *in AWS*; `account_type_id` is an
enum; `car_external_id` is an external identifier. Rewriting one would aim a
Terraform reference at a foreign id space. The full classification, with a
reason for every entry, is [`codegen/references.yaml`](https://github.com/kionsoftware/terraform-provider-kion/blob/main/codegen/references.yaml).

**`payer_id` resolves across billing source types.** Every billing source type
is keyed on the payer id, so `payer_id = 5` becomes a reference to whichever of
`kion_billing_source`, `_aws`, `_gcp` or `_oci` was imported with id 5 (the
`families:` table in `codegen/references.yaml`). A payer of a type the provider
has no resource for, such as an Azure CSP payer, stays a literal and is
reported.

**`parent_ou_id = 0`.** That means "top-level OU", not "the OU whose id is 0". A
reference there would invent a parent.

**Referents this configuration does not manage.** These are reported, never
silently kept:

```
7 reference(s) left as literals -- their referent is not managed here:
  kion_ou_cloud_access_role                3
  kion_user                                4
```

That is the boundary of your configuration. Three ways to close it:

1. **Import them too** — add the type to your `--include` and re-run. Usually
   the right answer for Kion-managed resources.
2. **A data source**, when the record already exists on the target and you want
   to bind by name rather than manage it:
   ```hcl
   data "kion_user" "alice" { filter { ... } }
   ```
3. **A variable**, when the value differs per environment.

Users, IDMS and stock permission schemes are the common cases — they come from
an identity provider or ship with the install, so they are rarely yours to
manage.

**One side of a relation that is settable from both.** A user names its groups
(`user_group_ids`) and a group names its users (`user_ids`). Both are genuine
foreign keys, but rewriting both directions closes a loop, and Terraform refuses
to plan a configuration containing one. Only the first direction offered becomes
a reference; the other keeps its literal, and is reported separately:

```
103 reference(s) left as literals -- a reference would be a dependency cycle:
  kion_user.user_group_ids                 45
  kion_user_group.user_ids                 57
  kion_user_group.viewer_user_ids          1
```

Nothing is missing here and there is nothing to import: the relation is already
expressed, from the other end. The refusal is per id, so a list keeps every
reference that was not part of a loop, and which direction wins is fixed by
block order in the file — the same edge loses on every run.

### Cloud access role exemptions

`kion_ou_cloud_access_role_exemption` and
`kion_project_cloud_access_role_exemption` are read from the parent's cloud
access role listing (`GET /v1/ou/{id}/ou-cloud-access-role`, field
`ou_exemptions`; `GET /v1/project/{id}/ou-cloud-access-role`, field
`project_exemptions`). Each record carries `id`, the parent id,
`ou_cloud_access_role_id` and `reason`, so `reason` is stored and imports with
its value.

The API fills those exemption lists only for a caller holding **Browse All Cloud
Access Roles** and **Manage OU** (**Manage Project** for a project) on the
parent. With less, the list comes back empty: `kion-import` finds no exemptions
under that parent, and a managed exemption refreshes as deleted.

Earlier releases read `/v1/{ou,project}/{id}/cloud-access-role-exemption`
instead. That route lists a different, cloud-rule-linked exemption kind, so an
exemption created through these resources was never found there
([#80](https://github.com/bshutterkion/terraform-provider-kion/issues/80)).

## Rewrite into module calls

`terraform plan -generate-config-out` writes bare `resource` blocks. To manage
the imported install through the modules under `modules/terraform-kion-*`
instead:

```sh
make import-modules-check IMPORT_IN=generated.tf IMPORT_OUT=modules.tf
```

That rewrites and then proves Terraform accepts the result; see *Check the
rewrite* below for why the second half matters. To rewrite without checking,
`make import-modules` or the underlying `kgen import-modules --in … --out …`.

Each `resource "kion_*" "<label>"` becomes `module "<label>"` with
`source = "./modules/terraform-kion-<name>"`. A block whose type has no module
keeps its type, labels, attributes and position, as does every non-resource
block — except for references into converted resources, which are retargeted;
see below.

The attribute-to-variable mapping comes from `modules/module_manifest.json`,
written by `make modules`, not from `variables.tf`. That matters because a
module variable is not always named after the provider attribute: `kgen module`
renames anything that collides with a `module` block meta-argument, so
`kion_cloud_rule`'s `source` becomes `cloud_rule_source` and
`kion_compliance_program`'s `version` becomes `compliance_program_version`.
Getting those wrong does not fail cleanly — Terraform reads a stray
`source = "aws"` as the module's own source address and goes looking for a
module there.

An attribute with no module variable (`last_updated` is the common case) is
dropped and reported rather than guessed at:

```
  dropped: kion_ou.example: dropped "last_updated" -- no module variable for this attribute
Wrote modules.tf: 3 block(s) rewritten, 1 left untouched, 1 attribute(s) dropped
```

### References are retargeted, so run it after `rewrite-refs`

`rewrite-refs` turns literal foreign keys into references, which is the shape a
configuration should end up in. Those references are retargeted at the module:

```hcl
ou_id = kion_ou.kion_ou_11.id      # before
ou_id = module.kion_ou_11.id       # after
```

That is mechanical rather than a guess: `rewrite-refs` emits exactly
`<type>.<label>.id`, every generated module exposes an `id` output, and the
block's label is preserved, so `module.<label>.id` is the same value by
construction.

Blocks the rewrite does not convert are retargeted too. A `locals`, an
`output`, or another provider's resource holding `kion_ou.kion_ou_11.id` would
otherwise be left naming a resource that no longer exists, and Terraform
refuses to load the file. That reference is the one edit made to a block this
rewrite does not otherwise own.

### Check the rewrite

The rewrite can tell you that every attribute mapped to a module variable. It
cannot tell you the result is a configuration Terraform will accept, because it
never looks at the module — only at the manifest's names. Two things get past
it:

- **A Required module variable the source resource never set.** Nothing was
  dropped and no warning fires; the module simply demands an argument that is
  not there.
- **A type the module declares differently** from the provider attribute the
  value came from.

Both surface only under `terraform validate`, which is why the checked target
exists:

```sh
make import-modules-check IMPORT_IN=generated.tf IMPORT_OUT=modules.tf
```

It rewrites into a scratch directory, serves the working-tree provider from a
filesystem mirror at the version `modules/` pins, then runs `terraform init`
and `terraform validate` against the real modules. Only on success is
`IMPORT_OUT` written:

```
Wrote .import-check/main.tf: 3 block(s) rewritten, 0 left untouched, 0 attribute(s) dropped
Validating the rewritten configuration...
Success! The configuration is valid.
✓ modules.tf rewritten and accepted by terraform validate
```

On failure it exits non-zero, prints the errors, and leaves the rewritten file
at `.import-check/main.tf` so you can see what Terraform objected to:

```
### terraform validate failed
Error: Missing required argument
  on main.tf line 1, in module "kion_ou_11":
The argument "permission_scheme_id" is required, but no definition was found.
The rewritten config is at .import-check/main.tf
```

It is plan-only and needs no credentials — the provider is never contacted.
The mirror exists because `terraform init` skips `dev_overrides`, and
validating a module call requires init. It also needs no published provider:
the check uses the provider built from your working tree.

`make import-modules` does the rewrite alone, without the check.

### Colliding labels are qualified

A resource label is unique per type; a module label is unique across the whole
configuration. `kion_cloud_rule.soc_2` and `kion_compliance_program.soc_2` are
both legal, and both occur -- a cloud rule and the compliance program it
enforces routinely share a name. On one 588-block import, 29 labels collided.

A colliding label is qualified with its resource type:

```hcl
module "cloud_rule_soc_2"         { source = "./modules/terraform-kion-cloud-rule" }
module "compliance_program_soc_2" { source = "./modules/terraform-kion-compliance-program" }
```

Labels that do not collide keep their name, and references follow the qualified
form.

### What is not retargeted

A traversal into any attribute other than `id` — `kion_ou.parent.name`, say. A
module exposes outputs, not attributes, and there may be no output of that
name, so retargeting would be a guess that plans against the wrong thing. Those
are reported and the command exits non-zero, having still written the output
for you to hand-edit:

```
  dangling: module "child": name reaches into kion_ou.parent, which this rewrite turns into a module (only .id can be retargeted, to module.<label>.id)
```

`rewrite-refs` never produces that shape, so it should not arise from the
standard flow.

## Choose what to import

**Importing everything is rarely what you want.** Kion ships large policy and
compliance catalogs. On a real install those dwarf everything a team manages in
Terraform: one production-scale install enumerates 53,513 records, of which
19,425 are Azure policy definitions and 16,265 are compliance checks.

Narrow the run with flags, a file, or both.

```sh
# just these
kion-import --include kion_ou,kion_project,kion_label --out imports.tf

# everything except the catalogs
kion-import --exclude kion_azure_policy,kion_compliance_check --out imports.tf
```

For anything you will run more than once, keep the selection in a file next to
the configuration, where it can carry comments:

```yaml
# kion-import.yaml
exclude:
  - kion_azure_policy      # Kion's shipped catalog, ~19k records
  - kion_compliance_check  # ditto, ~16k
  - kion_compliance_control
```

```sh
kion-import --selection kion-import.yaml --out imports.tf
```

`--include` and `--exclude` add to the file rather than replacing it, so the file
holds the standing policy and a flag covers a one-off. `--include`, when present,
is the entire set; `--exclude` is applied after it.

A type name that matches nothing is an error, not a silent no-op:

```
Error: unknown resource type(s): kion_labels
run with --list-types to see the 68 the manifest knows
```

## Aliases

Three resources are served under two names, for configurations written against
the previous provider:

| legacy name | current name |
|---|---|
| `kion_aws_cloudformation_template` | `kion_cft` |
| `kion_aws_iam_policy` | `kion_iam_policy` |
| `kion_cached_account` | `kion_account_cache` |

Both names reach one implementation over one endpoint, so importing both would
read the same objects twice and put two Terraform resources in charge of one Kion
record. `kion-import` enumerates only the current name; `--list-types` shows the
legacy ones marked `alias of`.

## Reading the report

Every type the tool knows produces exactly one line, so nothing is silently
absent:

```
kion_label                                    ok             225
kion_account_linkage                          empty            0
kion_dashboard                                error            0  GET /beta/dashboard: 405 Method Not Allowed
kion_aws_resource_tag                         unsupported      0  crud_archetypes.yaml declares kind: no_read

Coverage: 68 resource types, 53513 records
  ok: 59, empty: 2, error: 3, unsupported: 4
```

| status | meaning |
|---|---|
| `ok` | read, with a record count |
| `empty` | read, nothing there |
| `error` | the read failed; the reason is the API's |
| `unsupported` | no way to enumerate it, or a second name for another type |

Errors and unsupported types appear in the generated file as comments too, so the
output records what it could not reach rather than quietly omitting it.

If every read fails -- a rejected API key, the wrong `--url` or `--api-prefix` --
the run exits non-zero and writes nothing, `--probe` included.

A read that gets a 401, 429, 502, 503 or 504 is retried twice, after 1s and then
3s, because Kion occasionally rejects a valid key. Once a 401 outlasts its
retries, later 401s are not retried, so a bad key still fails quickly. A 500 is
not retried.

## Applying to another install

Imported configuration recreates its records when you apply it to an install
that does not already hold them. Some records cannot be recreated as they are.
The report lists them under `Will not apply to another install as-is`, and the
generated file puts a `# not portable:` comment above each one's import block:

| record | why it fails elsewhere | what to do |
|---|---|---|
| `kion_permission_scheme` with `system_managed` | every install has its own built-in schemes | reference the target's scheme (ids 1 to 4) |
| `kion_user_group` "Administrators" | every install has one: "User Group name already in use" | reference the target's group |
| `kion_custom_variable` named `system-…` | the prefix is reserved for Kion | drop it |
| `kion_funding_source` with no `ou_id` | Kion refuses to create one without an OU | set `ou_id` |

The check uses only the payload already read, so it cannot see everything. A
foreign key to a user or IDMS that exists only on the source install also fails;
`rewrite-refs` leaves those as literals and reports them.

## Other flags

| flag | |
|---|---|
| `--api-prefix` | defaults to `/api`. Use `--api-prefix ""` for an app serving the API at the root |
| `--skip-ssl-validation` | skip TLS verification |
| `--provider-version` | version constraint written into the generated config. Defaults to `>= ` the release that built the binary (`kion-import --version`), which the `>= 1.0.0` in `modules/*/versions.tf` overlaps |
| `--manifest` | override the embedded manifest. Maintainers only; see `codegen/README.md` |

## Differences from the previous importer

The Python importer in the SDKv2 provider generated `.tf` files itself and walked
the OU and project tree. This one emits import blocks and enumerates each
resource's collection endpoint, so counts are not comparable between the two: the
old tool reached a policy only if something it walked referenced it, while this
one lists the whole collection.

It also has no equivalent of the old `--sync` (pushing repository state back into
Kion) or the `--clone-*` flags. Those are a different job from importing.

## The generated file is self-contained

`imports.tf` carries its own `terraform { required_providers { … } }` block, so
put it in an empty directory. Dropping it beside a configuration that already
declares its providers fails with:

```
Error: Duplicate required providers configuration
A module may have only one required providers configuration.
```

## Parent-scoped import ids

Resources whose read finds the record under a parent take a compound id, and
`kion-import` emits it for you:

```hcl
import {
  to = kion_project_enforcement.example
  id = "42/167"     # project_id/id
}
```

`kion_scope_criteria` is the same shape (`scope_id/criteria_id`). Importing one
of these with a bare id leaves the parent unset, and the read looks for the
record under parent `0`.
