# Code generation

Everything the provider ships is generated from `spec/openapi3.json`. This file
is the guide to the inputs in this directory, the invariants that hold, and the
traps that have cost real time.

## Two paths, one spec

```
spec/openapi3.json
  |
  |-- tfplugingen ------------> internal/service/<n>/<n>_schema_gen.go   resource schema
  |
  `-- ogen --> kion-sdk-go --> kgen crud --> <n>.go                      resource CRUD
                                             <n>_data_source.go          the whole data source
```

The SDK is ogen-generated from the same spec, so both paths are spec-derived.
They are not two sources of truth.

**Resources** get their schema from tfplugingen and their CRUD from `kgen crud`.
**Data sources** are generated whole by `kgen crud`, schema included.

tfplugingen also *can* emit a data-source schema, but it maps the HTTP response
literally, so the result is the list envelope (`count`, `page`, `query`,
`sort_method`, `sort_order`, nested `data`) rather than the record a practitioner
selects. `kgen crud` unwraps the envelope and adds the `filter` blocks. Those
tfplugingen data-source schemas are therefore written only for packages that read
one, which is `gcp_regions` alone; see `dataSourceSchemaConsumers` in
`internal/kgen/schemas/schemas.go`. 48 of them used to be emitted, referenced by
nothing, and one drifted for months before anyone noticed.

## The inputs

| File | Generated? | Purpose |
|---|---|---|
| `generator_config.yaml` | **yes**, by `kconfig gen --write` | Per-resource CRUD paths driving schema generation |
| `config_overrides.yaml` | no | Paths and ignores `kconfig` cannot derive. Merged over the derivation |
| `schema_overrides.yaml` | no | What the spec cannot express: descriptions, plan modifiers, retypes, `sensitive`, attribute removal |
| `spec_additions.yaml` | no | Endpoints and property types the public spec omits. Merged into a working copy of the spec |
| `renames.yaml` | no | Attribute renames, applied before schema overrides |
| `crud_archetypes.yaml` | no | Which CRUD shape each resource uses |
| `private_endpoints.yaml` | no | Endpoints absent from the public spec, served over raw HTTP |
| `unbound_attributes.yaml` | **yes**, by `make bind-audit` | Ratchet over attributes no request sends. See "Nothing may be dropped in silence" |
| `memberships.yaml`, `state_upgrades.yaml`, `test_values.yaml`, `version_support.yaml` | mixed | See the header comment in each |

**Never hand-edit `generator_config.yaml`.** It is regenerated from the service
packages plus `config_overrides.yaml`. Anything you add by hand is lost on the
next `kconfig gen --write`, silently. If derivation cannot produce what you need,
put it in `config_overrides.yaml`, which is where the five entries below live.

## Invariants

`make codegen-check` enforces both. Run it after touching anything here.

- `kconfig check` reports drift between `generator_config.yaml` and the code. It
  must say `config in sync with code`.
- `kconfig gen --write` is idempotent, and produces no diff against the committed
  file. If regenerating changes it, someone hand-edited it.

Neither runs in `make ci`, because both need `spec/openapi3.json` and it is
gitignored. CI and a plain clone cannot check this, so it is on you.

If either breaks, derivation and the committed file have diverged and the file is
no longer generated in any meaningful sense. Both were broken before this was
written: five entries existed only by hand, and regenerating dropped `count` and
`page` from `billing_source`'s ignores, which fails schema validation because
`count` is a reserved Terraform root attribute name.

## Nothing may be dropped in silence

Request bodies are assembled by matching SDK fields to model fields by
json/tfsdk name. Until recently, a field that matched nothing was skipped
without a word — in either direction. That is invisible to `build`, `vet`,
`lint` and the unit tests, because the result is a struct literal that compiles
and type-checks. It is also the shape of two shipped bugs:

- `custom_variable` split one polymorphic `default_value` into three typed
  attributes, matched none of them, and sent a create body with no default value
  in it (#61).
- `permission_scheme`'s update body nests `app_policy`/`permission_roles` while
  the model is flat, so nothing matched and the provider sent `{}` (#64).

`make bind-check` (`internal/kgen/bindaudit`, run by `ci-test`) fails when a
settable attribute is read by neither the resource's `Create` nor its `Update`.
It reads the **generated output**, not the spec, so bespoke and hand-written
bodies are audited by the same rule and it runs without `spec/openapi3.json`.
Pre-existing debt lives in `unbound_attributes.yaml`, a ratchet: adding an entry
is a reviewed act and an entry that stops reproducing fails the test too.

Three derivations exist so that shapes like the two above map instead of being
dropped, and none needs a declaration of its own:

- **Object wrap** — a body field that is a struct the model does not mirror,
  whose sub-fields *are* the model's top-level attributes (`app_policy`). The
  inverse of the existing `FlatSubs`.
- **Implode** — a body array that a declared `read_shape` explode already
  describes; writing regroups the flat model rows the read exploded. Inverting
  the read shape means the two cannot drift apart.
- **Polymorphic raw value** — a `jx.Raw` body field the schema re-expresses as
  `<name>_string` / `<name>_list` / `<name>_map`. The generator emits a builder
  that takes exactly one of them, and a flattener that takes the value back
  apart on read so it survives refresh and import.

A private update that **overwrites every column** from its body (the payer PUT
behind `kion_billing_source_aws`/`_gcp`) cannot be built from the model alone:
whatever the model does not carry would be blanked. Declare a `write_shape` in
`private_endpoints.yaml` beside the `read_shape`; the generated update reads the
record, lays the listed plan values over it (`flex.MergeJSON`) and sends it back.
A null plan value leaves the read value in place, and an enum value the map does
not list fails the apply rather than being dropped.

`kgen crud` also checks every `read_shape` kind against the model attribute it
flattens into. A mismatch compiles and then fails every read in the Value
constructor; that is how `kion_billing_source_gcp`'s `billing_start_date`,
declared `int` against a string attribute, broke its every refresh.

Each baseline entry ends in one of four decisions:

- **Bind it** through the endpoint that does change it: owners, associations and
  slice members in `memberships.yaml` (parent_list resources included), or a
  `moves` entry in `crud_archetypes.yaml` for a parent change that has its own
  move endpoint (`kion_ou`, the azure/gcp/custom accounts). A completed move is
  written to state at once, so a later failure cannot leave state naming the old
  parent. An attribute only the update body carries is applied right after create
  with `create_via_update` (parent_list) or the resource's own template.
- **`RequiresReplace`** (`schema_overrides.yaml`) when the API fixes it at create
  and recreating is what the change means. This removes it from the update side.
- **Remove it** when it is not an input at all, such as a list endpoint's filter
  query param leaking into a resource schema.
- **Keep it with a reason** when nothing can change it and replacement would
  destroy real data (an account, OU, project or funding source). The reason
  names the body and, where one exists, the resource that owns the relation.

New entries arrive as `TODO`; a `TODO` is debt, not a decision.

## Derivation guesses, and sometimes guesses wrong

`codegen/import_manifest.json` takes its list paths from the `data_sources`
section, so a wrong entry there is not cosmetic: it sends `kion-import` at a path
that cannot be listed. Two were caught this way when the config was first made
reproducible.

- `saml_group_association` derived `/v3/idms/group-association/{id}`, whose
  collection is POST only. The records hang off the parent, so the read is pinned
  to `/v3/idms/{id}/group-association`.
- `custom_variable_override` is polymorphic across account, OU and project, so
  derivation resolved whichever GET the data source reached first, `/v3/ou`, and
  called it the collection. There is no single list, so the entry is suppressed
  with `skip: true`. It is still importable: `kion-import` reads all three
  `/v3/{entity}/{id}/custom-variable` collections from `multiParentOverrides` in
  `internal/kgen/importmanifest/generate.go`, not from this file.

**`match` restricts the parents a parent-scoped read is made under.** It sits on
a data source's `read:`, maps a field of the parent's list record to the value
it must have, and is carried into `generator_config.yaml` and from there into
the manifest row's `parent.match`. `kion-import` skips a parent that does not
match without requesting it. `idms_group_association` and
`saml_group_association` carry `match: { idms_type_id: "3" }`, because the API
serves `/v3/idms/{id}/group-association` only for a SAML IDMS and fails for every
other type (#48).

```yaml
data_sources:
  saml_group_association:
    read:
      path: /v3/idms/{id}/group-association
      method: GET
      match: { idms_type_id: "3" }
```

A rule that cannot apply fails the build rather than filtering nothing, or
filtering out every parent. `kconfig` rejects `match` on a resource, on a read
whose path has no parent placeholder, and an empty map, field or value.
`kgen import-manifest` rejects a field the parent resource's schema does not
declare, and a data source whose manifest row has no single parent.
- `ou_permission_mapping` and `project_permission_mapping` create and update
  through the same PATCH upsert, and derivation attributes that one PATCH to
  create alone. Without a pinned `update` they become create-and-replace only.
- `project` calls two create ops, one per financial mode, so derivation would
  pick whichever appears first in the source and the schema would follow it.
  Only `/v3/project/with-spend-plan` may drive it: its body is `ProjectCreate`,
  carrying `project_funding`, where `/v3/project/with-budget`'s carries `budget`
  instead. One schema serves both modes, so it needs `project_funding` from the
  derivation and `budget` from `schema_overrides.yaml`.

Check a derived path before trusting it: it must exist in the spec with a GET,
and it must be the collection rather than the by-id route.

## Billing sources with no public API (`kind: payer`)

Kion creates every billing source type through one private route, `POST
/v1/payer/standalone`, with the type's fields nested under a per-type key, and
reads and updates them all as a whole payer at `/v1/payer/{id}`. No derived
archetype has a private create, so `kind: payer` emits the resource from
`payer.gtpl`, driven entirely by its `private_endpoints.yaml` entry: create
merges `write_shape` into `create_seed`, update merges it into the record it
reads first (the `PUT` overwrites every column), read uses `read_shape`, and a
`write_only` scalar is neither read back nor import-verified.

Its schema comes from `spec_additions.yaml` under schema-only path keys (e.g.
`/v1/payer/standalone/azure`), because the real create path is shared by every
type and one OpenAPI path cannot carry two bodies. Those keys are not routes.
kion-import lists these resources from `/v4/billing-source`
(`explicitFlatListOverrides`), filtered by the type's key. `billing_source_azure`
and `billing_source_anthropic` use it; a type Kion added after the oldest
tracked release declares `min_kion_version`, since version_support cannot
derive a gate from a private route.

## A verbatim resource body

`crud_archetypes.yaml` entries of `kind: entity` may carry
`resource_template: <name>`, which replaces the derived `<name>.go` with a
verbatim template registered in `resourceTemplates`
(`internal/kgen/crud/bespoke.go`). Unlike a bespoke *kind*, the resource stays in
`generator_config.yaml`'s `resources` list, so its schema, data source, sweeper
and acceptance tests are still derived; only the CRUD body is hand-authored.

It is for a resource whose control flow the entity archetype cannot express while
everything else about it is ordinary. `project` is the only one: create has to
choose `POST /v3/project/with-budget` or `POST /v3/project/with-spend-plan` on
the install's financial mode, because each endpoint refuses the other mode before
it parses the body, so no payload can satisfy the wrong one. (`POST /v3/project`,
which the derived body used, is the deprecated spend-plan-only alias and is
unusable on a budget-mode install — the portal's own swagger comment says so.)

Reach for it last. A verbatim body stops tracking schema and SDK changes, which
is the cost every bespoke archetype pays; take it only when a declared tweak
cannot do the job.

## What derivation cannot see

`kconfig` derives a resource's ops from the SDK calls its CRUD methods make. Five
entries have no such calls to find and are supplied by `config_overrides.yaml`:

- `funding_source_note` is served entirely over raw HTTP through `internal/conns`,
  so it calls no SDK op at all.
- `funding_source_enforcement` and `ou_enforcement` read a list under their parent
  (`GET .../enforcement`), which no naming convention finds. Without the override
  both render as `INCOMPLETE, missing read` and are dropped, because tfplugingen
  requires create and read.
- The `ou` and `user_group` data sources reach their read through a helper, so the
  SDK call is not attributed to the package.

## Traps

**An untyped spec property drops the whole type.** tfplugingen refuses a property
with no `type` and no `allOf`/`oneOf`/`anyOf`, and it discards the entire resource
or data source rather than that one field. It logs the skip and exits 0, so the
run looks successful while the previously generated file sits stale on disk.
`scope`, `scope_criteria` and the `custom_variable` data source were all skipped
this way, and stayed stale long enough to miss a schema-wide change. `kgen schemas`
now compares the produced code spec against the config and fails, naming each
missing type. Fix it by giving the property a type under `property_types` in
`spec_additions.yaml`, or by ignoring it in `config_overrides.yaml`.

**An `ignores` entry must match the path tfplugingen reports.** It is not fuzzy.
The `custom_variable` data source ignored `data.default_value` where the real path
is `data.items.default_value`, because the list response nests records under
`items`. The ignore was never applied and the data source was dropped anyway. The
path appears in the warning text.

**Blocks cannot be generated.** The provider code spec has no representation for
them, so tfplugingen emits nested attributes only. This is why `account` and
`aws_account` still declare their schema by hand: they expose
`aws_organizational_unit` and `move_project_settings` as blocks, and delegating
would change the HCL practitioners write from `foo { ... }` to `foo = { ... }`.
It is also why `kgen crud` owns data sources, 38 of which have a `filter` block.
Do not expect a configuration change to fix this; it needs block support in the
upstream toolchain.

**Overrides are add-if-missing.** Naming an attribute that does not exist creates
it rather than failing. Name one at the wrong nesting depth and you get a second,
unpopulated attribute alongside the real one. Nested overrides merge into the
existing children, so overriding one field of a nested object leaves its siblings
alone.

**`config_overrides.yaml` accepts `ignores` at the entry top level or under
`schema:`.** Both work. Unknown keys fail the generate rather than being dropped,
which they were until four billing-source resources were found leaking list-query
parameters into their schemas.

## Adding a resource

```bash
go run ./cmd/kgen service --name CloudRule --snakename cloud_rule
# add entries to crud_archetypes.yaml, and to config_overrides.yaml if the paths
# do not derive, then:
go run ./cmd/kconfig gen --write   # the resource must be in generator_config.yaml
go build -o /tmp/kgen ./cmd/kgen   # BEFORE generating: see below
SDK_DIR=$(go list -m -f '{{.Dir}}' github.com/kionsoftware/kion-sdk-go)
make version-support
for pass in 1 2; do
  /tmp/kgen versions --sdk "$SDK_DIR"
  /tmp/kgen schemas
  [ $pass = 1 ] && /tmp/kgen crud --sdk "$SDK_DIR" --force
done
/tmp/kgen import-manifest
make modules examples docs
```

`make generate` cannot bootstrap a new package. It runs `go run ./cmd/kgen`,
which compiles the scaffold, and the version gate it writes first names a type
the scaffold does not have. A prebuilt kgen avoids that, as in
`scripts/regen-from-scratch.sh`. `crud --force` is needed because plain `crud`
skips the scaffold's existing files. After the first run, `make generate` works.

Then refresh `codegen/schema_snapshots/new.json` (its README). The import
manifest and the references/bind completeness gates read the snapshot, so they
skip a resource it does not have.

Adding or removing a resource changes four generated trees: `internal/service/`,
`examples/`, `docs/` and `modules/`. `make ci` gates the last three, so a missed
regeneration fails the build rather than shipping stale documentation.

## Regenerating

```bash
make refresh-spec        # fetch the spec locally (maintainers; needs the SDK monorepo)
make generate-schemas    # tfplugingen: resource schemas
make crud-force          # kgen crud: CRUD and data sources, overwriting
make modules docs        # Terraform modules and registry docs
go run ./cmd/kconfig gen --write   # generator_config.yaml
```

`spec/openapi3.json` is gitignored. The generated code is committed, so a plain
clone builds, tests, and regenerates modules, docs and examples with no spec.
Only the two schema generators need it.
