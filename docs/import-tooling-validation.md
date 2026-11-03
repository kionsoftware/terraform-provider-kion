# kion-import: live validation

Record of running `kion-import` against real Kion installations. The collection
paths a parent-scoped or association resource reads through are authored in
`codegen/config_overrides.yaml` rather than derived -- codegen records by-id
reads, which those shapes have none of -- so they are unverified until a run like
this exercises them. Re-run this after changing an archetype, a read path, or one
of those entries.

**Run it against more than one install.** Two claims in earlier versions of this
document were properties of the single install they were measured on and read as
properties of the provider: one understated a defect's reach, the other
overstated it. Claims here are scoped to the installs that support them.

## Environments

The detailed sections below are a demo install at 3.16.3. The same build was
probed against four more, read-only:

| install | Kion version | records | ok | empty | error | unsupported |
|---|---|---:|---:|---:|---:|---:|
| demo A | 3.16.3 | 51,966 | 60 | 5 | **0** | 3 |
| demo B | **3.15.10** | 43,896 | 56 | 9 | **0** | 3 |
| QA | 3.16.3-dev | 11,955 | 47 | 16 | 2 | 3 |
| local dev | dev | 9,077 | 42 | 21 | 2 | 3 |
| production | 3.16.4 | 37,286 | 47 | 11 | 7 | 3 |
| billing-connected local | dev | 7,803 | 39 | 25 | 2 | 2 |
| local 3.18-dev, end to end | dev | 7,808 | 41 | 31 | **0** | 2 |

| | |
|---|---|
| Date | 2026-08-25, on `main` @ f4ec800 |
| Manifest | `codegen/import_manifest.json`, 68 resource types |
| Command | `kion-import --url https://kion.example.com --probe` |

Credentials come from `--api-key` or `KION_APIKEY`; none are recorded here. A
local dev install serves its API at the root, so it needs `--api-prefix ""`.

### The billing-connected local install (2026-09-11)

Added while working #99. It is the install that raised `KION_ACC_*` discovery
from 2/10 to 8/10, so it exercises paths the emptier ones skip. Its high `empty`
count is a property of the install, not a regression: it holds little data.

Its two errors are both #48 -- `kion_idms_group_association` and
`kion_saml_group_association` fail on `GET /v3/idms/1/group-association`, which
closes the connection. No other resource errored. `kion-import` has since
stopped reading that collection under non-SAML parents; see *The IDMS
group-association endpoint*.

`unsupported` is 2 rather than the 3 in the five-install table for the reason
already recorded below: the third was `kion_custom_variable_override`, refused
for a compound identity and since supported. The two that remain are the alias
types, refused by design.

Three rows changed meaning on this run, and each was a defect fixed rather than
a difference between installs:

- `kion_account_linkage` reads **ok** where it could not be read at all. The
  SDK cannot decode the response -- `delete_finalized_at` is typed non-nullable
  and Kion returns null -- so the read failed before reaching any record. It now
  reads raw.
- `kion_billing_rule` reads through a hand-declared shape. The spec's read
  schema is empty, so it previously imported as a bare id with every attribute
  missing. It shows `empty` here only because the install holds no billing rules.
- `kion_funding_source` recovers `permission_scheme_id` again. The route it used
  returns rows only for a scheme that already carries role permissions, so the
  attribute came back null for any scheme created alongside its funding source.

### End to end on a local 3.18-dev install (2026-10-20)

The first run here to go past `--probe`: a full `kion-import`, then
`terraform plan -generate-config-out` against a provider built from the same
commit, then a second plan of the generated configuration. Portal `master`
with demo data.

| stage | before | after |
|---|---|---|
| probe errors | 1 (GovCloud, #127) | 0 |
| import blocks | 7,817 | 7,808 |
| plan errors, generating config | 11, then 731 once the walk got further | 0 |
| plan of the generated config | -- | 7,808 to import, 0 to change |
| `terraform apply` | -- | 7,808 imported, 0 added/changed/destroyed, 37 s |
| plan after apply, from state | -- | No changes, 0 errors |

Six classes of defect only showed past the probe, because each was in the
provider's handling of an existing record rather than in enumeration: owner
validators that rejected ownerless built-ins (#129), Reads that left required
or settable attributes unset (#128, #130, #132, #133), a create-time validator
on `engine` (#131), and typed billing sources claiming every payer kind
(#134). **A clean probe does not mean an install imports.** Run the plan.

The apply only reads, so it writes nothing to Kion. Its one warning is
`kion_dashboard`'s version gate, "Kion version could not be determined",
because a local dev build reports an empty version; a released install does
not.

### Repeat on a fresh local install, at 1.8.0 (2026-11-02)

The same end-to-end cycle, against a freshly started local dev install and a
provider built at `v1.8.0`, after 1.6.0–1.8.0 changed the billing sources,
cloud access roles, saml_group_association and kion-import itself. The API is
served at the root, so `--api-prefix ""` and `KION_APIPATH=""`.

| stage | result |
|---|---|
| `kion-import` | 76 resource types, 7,810 import blocks, 0 read errors, 12 read with a caveat |
| plan, generating config | 7,810 to import, 0 to change, 0 errors |
| plan of the generated config | 7,810 to import, 0 to change |
| `terraform apply` | 7,810 imported, 0 added/changed/destroyed, 40 s |
| plan after apply, from state | No changes, 0 errors |

The caveats are the expected empties on an install with only the internal
IDMS (no SAML or OpenID records). A grep for `Error:` in the logs matches eight
lines, all inside imported CloudFormation templates' embedded Python.

### Read that table by column, not by row

**Record counts are properties of the install** and are not comparable between
them. Only counts from the same install at two commits mean anything, and this
document flags those where they appear.

**Status columns are mostly properties of the tooling.** The exceptions are
named below, and they are why this now covers five installs rather than one.

### What holds on all five

- **68 rows in, 68 results out, everywhere.** No resource is silently absent
  from a report on any install.
- **No unexplained failure anywhere.** Every non-`ok` row names a cause.
- **The three `unsupported` rows are identical across all five**, and are
  refusals by design: two alias types, and one compound identity. (Down from
  four -- `custom_variable_override` moved out; see *A compound identity with a
  plural sibling*.)
- **No read-shape defect appeared on 3.15.10 that 3.16.x does not have.** Demo B
  is a full Kion minor behind and produced zero errors. First evidence the read
  shapes and version gates hold across 3.15.10 -> 3.16.4, and the negative
  result worth having.

### What varies

**A scoped API key looks exactly like a broken read.** The production install
accounts for 7 of the 11 errors across all installs, and 6 are `403 Insufficient
permissions` -- on `aws_resource_tag`, `ou_cloud_access_role`,
`project_cloud_access_role`, `funding_source_enforcement`, `ou_enforcement` and
`project_cloud_access_role_exemption`. Nothing is wrong with those reads. Check
a key's scope before filing a bug.

**An `empty` is not evidence of anything.** The smaller installs report 21 and 16
empties against demo A's 5; they simply hold less. `empty` means the collection
responded and had no records.

**Two findings recorded here as universal were demo-A-only:** the exemption
collections "mixing kinds" (237 records of another kind -- every other install
returns those collections empty), and `kion_ami`, `kion_category`,
`kion_funding_source_note`, `kion_project_line_item` and `kion_service_catalog`
having records at all. The ratio was a property of that install's data. (The
exemption resources have since moved to a different collection that needs no
kind filter; see *Correction: the exemptions read the wrong collection*.)

**One finding recorded as install-specific is universal** -- the next section.
That is the mistake this table exists to prevent.

### The IDMS group-association endpoint

`GET /v3/idms/{id}/group-association` was recorded here as "server-side and
specific to IDMS 1 on that install". It is neither.

| install | version | result by IDMS |
|---|---|---|
| demo A | 3.16.3 | `id1(type1)=502`, `id3,16,17,18,19(type3)=200` |
| demo B | 3.15.10 | `id1(type1)=502`, `id3(type2)=502`, `id2,4,9,11(type3)=200` |
| QA | 3.16.3-dev | `id1(type1)=502`, `id2(type5)=502` |
| production | 3.16.4 | `id1(type1)=502`, `id2(type3)=200` |
| local dev | dev | `id1(type1)=` connection dropped |

**Type 3 (SAML) returns 200 everywhere; every other type 502s everywhere** --
four installs, three Kion versions, production included. `idms_type_id` is 1
Internal Directory, 2 Active Directory, 3 SAML.

The local dev install does not 502, it drops the connection. A 502 from a hosted
install is a proxy rendering a backend that died, so that is likely the same
fault without a proxy in front of it.

`kion_idms_group_association` and `kion_saml_group_association` both enumerate
through this endpoint. Before #48 was addressed, an install whose IDMSes were not
all SAML reported every other IDMS as a failed parent, and the QA install, with
no SAML IDMS at all, reported both resources as `error`.

`kion-import` now reads the collection only under SAML parents: the manifest
row carries `"match": {"idms_type_id": "3"}` on its parent block, and a parent
whose list record does not match is skipped without a request. Group
associations exist only on a SAML IDMS, so nothing is lost. An install with no
SAML IDMS reports both resources `empty`, naming the skipped parents. The
server defect itself is still upstream; the rows below predate the change.

## Result

```
Coverage: 68 resource types, 51987 records
  empty: 5, ok: 61, unsupported: 2
```

51,901 `import` blocks generated, **zero errors**, and zero records skipped for a
missing id. Every one of the 68 manifest rows produced a result.

The two remaining `unsupported` rows are refusals by design, not gaps: alias
types (`kion_aws_cloudformation_template`, `kion_aws_iam_policy`) that would put
two Terraform resources in charge of one record.

`kion_custom_variable_override` used to be the third, refused for a compound
identity. It now enumerates; see *A compound identity with a plural sibling*.

### How this run compares

| run | install | records | ok | empty | error | unsupported |
|---|---|---|---|---|---|---|
| 2026-08-24 | production-scale | 53,513 | 49 | — | 15 | 4 |
| 2026-08-25 (earlier) | production-scale | 53,513 | 59 | 2 | 3 | 4 |
| 2026-08-25 @ 69c566c | demo (3.16.3) | 48,355 | 59 | 6 | 0 | 3 |
| 2026-08-25 @ d5a9200 | demo (3.16.3) | 51,966 | 60 | 5 | 0 | 3 |
| **2026-08-25 @ f4ec800** | **demo (3.16.3)** | **51,987** | **61** | **5** | **0** | **2** |

The last row differs from the one above it by exactly the 21
`custom_variable_override` records, which is the whole of the +21.

**The rows are not the same install** — the first two ran against a
production-scale install, the last two against a demo one. Compare the *status
columns*, which are properties of the tooling; the record counts are properties
of the install and are not comparable across rows.

The last two rows *are* the same install, so their counts are comparable, and
the 3,611-record difference between them is entirely records the earlier run
never fetched: 3,592 `compliance_control` and 10 `compliance_family` to the
padding fix, 9 `scope_criteria` to the `nested_collection` read shape. The
resource that moved from `empty` to `ok` is `scope_criteria`, which had read
nothing at all.

The last three errors and the fourth `unsupported` were cleared by the two fixes
recorded below (*Private reads that render SQL null wrappers* and *`no_read` did
not mean unreadable*). `kion_dashboard` and `kion_funding_source_note` now read;
`kion_aws_resource_tag` and both `*_cloud_access_role_exemption` types moved out
of "structurally unreadable" entirely.

The production-scale figures are the ones quoted in
[the importing guide](guides/importing.md): 53,513 records, of which 19,425 are Azure policy
definitions and 16,265 are compliance checks.

## Defects this run found

Five, none of which unit tests could have caught — each depends on a response
shape only a real install produces.

### 1. `/api` prefix, and an HTTP 200 that is not JSON

`https://kion.example.com` returns the web application's HTML **with status 200**;
only `https://kion.example.com/api` returns API JSON. A status-code check cannot
distinguish them, so the original client fed HTML to the JSON parser for all 68
resources and produced 68 identical unmarshal errors.

Fixed: the client appends `/api` by default, `--api-prefix` overrides it (its only
real use is an employee hitting an app directly on localhost, where the API is at
the root), and a non-JSON body now produces one error naming the likely cause
instead of 68 parse failures. Matches `kion-env-copy`'s `KION_API_PREFIX`
convention.

### 2. Doubly-nested list envelope

`/v4/billing-source`, `/v3/label`, `/beta/scope` and others return:

```json
{"status":200,"data":{"pagination":{…},"total":17,"items":[…]}}
```

`unwrap` handled `{items,total}` at the top level and `data` as an array, but when
`data` was an object it took the single-object branch and returned the inner
envelope itself as one record — producing exactly one id-less record per resource.

Fixed: an object `data` carrying `items` is unwrapped, and the inner `total` drives
paging. A genuine single-object `data` still yields one record.

### 3. Per-type record wrapper

A family of `/v3/*` endpoints wraps each record under a type key alongside sibling
arrays:

```
/v3/cft                -> {"cft":{…,"id":296}, "owner_users":[], "owner_user_groups":[], "tags":[]}
/v3/azure-role         -> {"azure_role":{…}, "car_restricted_users":[], "car_restricted_ugroups":[], …}
/v3/service-catalog    -> {"service_catalog_portfolio":{…}, …}
/v3/gcp-iam-role       -> {"gcp_role":{…}, …}
```

The real `id` is one level down, so every record was skipped — about 23,000 across
ten resources.

**The wrapper key is not the resource kind.** `service_catalog` wraps under
`service_catalog_portfolio` and `gcp_iam_role` under `gcp_role`. Detection must be
structural, never derived from the tf_type.

Fixed with a rule that fits every observed shape: if a record has no top-level `id`
and **exactly one** of its keys maps to an object that itself contains an `id`, that
inner object is the record. Sibling arrays and scalars are ignored; two candidates
means ambiguous and the record is left untouched. `Record.Raw` keeps the outer
object.

### 4. Alias types were imported twice

`kion_aws_iam_policy` and `kion_iam_policy` are two names for one implementation
over one endpoint, as are `kion_aws_cloudformation_template` and `kion_cft`. The
manifest gave each its own row with the same `list_path`, so both were
enumerated: 1,348 identical ids under two type names, plus 296 more, for 1,644
duplicate `import` blocks. Applying them would have put two Terraform resources in
charge of each Kion record, each reverting the other's drift.

The schema snapshot the manifest is built from lists every tf_type the provider
serves, aliases included. `kindAliases` existed to map an alias back to its
generator_config key for lookup, but nothing stopped the alias getting its own
enumerable row.

Fixed: an alias row carries `alias_of` and `readable: false`, so it is still
accounted for in the report but never enumerated. `--list-types` shows the legacy
name against the current one, which is what an operator migrating a configuration
needs to see.

### 5. Pages padded to `total`, hiding real records

Found by investigating the `record(s) skipped: no id` caveats an earlier run
recorded, rather than accepting them as a documented quirk. This run is the
first with the fix in place.

Some collections pad every page out to `total` instead of to the page size.
`/v4/compliance/program/5/family` holds 110 families:

| request | items | real | zero-valued padding |
|---|---:|---:|---:|
| `count=100&page=1` | 110 | 100 (Access…Temporary) | 10 |
| `count=100&page=2` | 110 | 10 (Transaction…Wireless) | 100 |
| `count=500&page=1` | 110 | 110 | 0 |

`Client.List` unwrapped 110 records and `total: 110` from page 1, evaluated
`len(records) < total` as `110 < 110`, and stopped. The ten real records that
exist only on page 2 were never requested. `toRecords` then dropped the ten
fillers for having no id and reported `10 record(s) skipped: no id`.

The count was right and the wording was wrong in the way that matters: it
describes records being *ignored* when the event was records never being
*fetched*. It also happened to equal the number missing, which made it look
self-consistent. What gave it away was the alphabet -- every missing family
sorted between Transaction and Wireless, which is a lost final page, not
scattered bad data.

Fixed: zero-valued records are dropped before the page count is compared against
`total`. Such a record carries no id and could never have produced an `import`
block, so nothing is lost, and the count now reflects what was retrieved.

**3,602 records on this install were missing, not 10.** `compliance_family` was
the resource that exposed the mechanism, and its ten losses are what made the
alphabetical clustering visible, but `compliance_control` was losing 3,592 by
the same route:

| resource | before | after |
|---|---:|---:|
| `compliance_control` | 2,414 (3,592 "skipped") | **6,006** |
| `compliance_family` | 587 (10 "skipped") | **597** |

597 matches a direct count over the API. Measured against a binary built from
the pre-fix commit, not inferred: the first reading of this assumed
`compliance_control`'s skips were harmless padding and that it recovered
nothing, which was wrong by 3,592 records.

The padding itself is a server-side defect and is worth reporting separately:
any client paging that endpoint hits it.

## Still failing on this install

Nothing errors. Three rows are refused by design, six are legitimately empty, and
five read with a caveat. **All are recorded, not silently omitted** -- they
appear in both the report and the generated file.

### Refused by design (2)

`aws_cloudformation_template` and `aws_iam_policy` are alias tf_types; see
*Alias types were imported twice* below.

This list used to include `aws_resource_tag`, `ou_cloud_access_role_exemption`
and `project_cloud_access_role_exemption` as "structurally unreadable". That was
wrong -- see *`no_read` did not mean unreadable*. It also used to include
`custom_variable_override`; see *A compound identity with a plural sibling*.

### Read with caveats (6)

| resource | caveat |
|---|---|
| `idms_group_association` | 2 parents failed; `GET /v3/idms/1/group-association` 502 |
| `saml_group_association` | same 502, same parent |
| `idms_open_id_access_rule` | 10 parents had none |
| `ou_cloud_access_role_exemption` | 237 records with no `ou_cloud_access_role_id` set |
| `project_cloud_access_role_exemption` | 12 such records, so none remain |
| `custom_variable_override` | 1,791 records with no `override` set |

The 502 was recorded here as specific to IDMS 1 on that install. It is not: the
API fails this read for every IDMS that is not SAML, on every install measured
(see *The IDMS group-association endpoint*, #48). IDMS 1 is the Internal
Directory. `kion-import` now skips non-SAML parents, so these two rows no longer
report a failed parent.

The two exemption caveats are the kind-mixing filter working; see below. Those
rows were read from a collection the exemption resources no longer use; see
*Correction: the exemptions read the wrong collection*.

This table used to carry three more rows -- `compliance_control` (3592),
`compliance_family` (10) and `scope_criteria` (9), each reported as
`record(s) skipped: no id`. None of them were what that wording implied; see
*Pages padded to `total`* below for the first two, and the `nested_collection`
read shape for `scope_criteria`. All three now read clean:

| resource | then | now |
|---|---|---|
| `compliance_control` | 2414, 3592 skipped | **6006**, no caveat |
| `compliance_family` | 587, 10 skipped | **597**, no caveat |
| `scope_criteria` | 9 skipped, 0 read | 9, no caveat |

### Empty (5)

Five collections are genuinely empty on this install, including
`kion_aws_resource_tag`. No install had a tag, so its read was first exercised
on 2026-10-20 (#30) by creating one locally: the payload is exactly the spec's
`AWSResourceTag` (`id`, `resource_key`, `resource_value`), and create, refresh,
`kion-import` plus import, and detecting an out-of-band delete all work.


## What this run validated

- **The parent-scoped fallback works.** `compliance_family` and `compliance_level`
  405 on their flat lists and fall back automatically, returning 597 and 64 records.
- **Skip counting is load-bearing.** Before the wrapper fix, this run reported
  `kion_azure_policy: empty, 0 records` as a clean result. With skip counts it
  reported `19425 record(s) skipped: no id`. Roughly 23,000 records were being
  dropped silently; the count is what surfaced it.
- **A skip count is a symptom, not a diagnosis.** The same mechanism that
  surfaced the wrapper bug later described ten *missing* `compliance_family`
  records as ten *skipped* ones, and the number was accurate enough that the
  caveat sat in this document as settled. A skip count says something was not
  emitted; it does not say the records were reachable and ignored. Reconcile
  against the API before recording one as understood.
- **The one-result-per-row contract holds** against the real manifest: 68 rows in,
  68 results out, every failure attributed.

## Private reads that render SQL null wrappers

Enumerating a record is not the same as reading it. `--probe` only exercises the
list endpoint, so `kion_project_note` and `kion_dashboard` both reported `ok`
while every one of their records failed at `terraform plan
-generate-config-out`, where the provider's own `Read` runs:

```
decoding response: json: cannot unmarshal object into Go struct field
project_noteWire.data.updated_at of type string
```

Their private reads render several columns through Go's `sql.Null*` types, so
those arrive as objects and the whole response fails to decode. Observed
directly against a live install:

| endpoint | attribute | on the wire |
|---|---|---|
| `/v2/project-note/{id}` | `last_update_user_id` | `{"Int":0,"Valid":false}` |
| `/v2/project-note/{id}` | `updated_at` | `{"Time":"0001-01-01T00:00:00Z","Valid":false}` |
| `/v1/dashboard/{id}` | `description` | `{"String":"…","Valid":true}` |
| `/v1/dashboard/{id}` | `updated_at` | `{"Time":"…","Valid":true}` |
| `/v1/dashboard/{id}` | `config` | the JSON **object**, not a string of JSON |

`created_at` is a plain string on both, so the wrapping is per-column and cannot
be inferred from the attribute's name or type. `spec_additions.yaml` declares the
private Dashboard by hand and had all three of these wrong, which is why the
generated wire struct disagreed with reality.

Declare the real shape per attribute in `private_endpoints.yaml` under
`read_kinds` (`null_int`, `null_string`, `null_time`, `json_string`). Unlike
`read_shape` this keeps the single wire struct, so a resource whose write is also
raw keeps working: the `flex.Null*` types encode back to the bare scalar.

**When a resource fails only under `plan`, fetch its private read directly and
compare it field by field against the generated wire struct.** The probe cannot
see this class of failure.

## A compound identity with a plural sibling

`kion_custom_variable_override` was refused because its read path carries two
placeholders:

```
/v3/account/{account_id}/custom-variable/{custom_variable_id}
```

Neither id can be discovered from the other, so derivation's compound-identity
veto marked it unreadable. That veto is still correct **for that path**. What
was never checked is whether the path has a sibling that needs only one id --
and it does:

```
/v3/account/{id}/custom-variable   -> every custom variable at that account
/v3/ou/{id}/custom-variable        -> same, per OU
/v3/project/{id}/custom-variable   -> same, per project
```

So the cross product is enumerable in **one request per entity**, not one per
(entity, variable) pair. Measured on the demo install:

| | accounts | OUs | projects | total |
|---|---:|---:|---:|---:|
| entities | 92 | 22 | 37 | **151** |
| overrides found | 5 | 12 | 4 | **21** |

12 custom variables exist, so the cross product is 151 x 12 = 1,812 cells, of
which 21 carry an override. Probing each pair directly would have been 1,812
requests; the plural collection makes it 151, all of which returned 200.

Three things had to be true, each verified live:

- **`override` is local, not inherited.** Each element carries both an
  `inherited` block (where the effective value comes from) and an `override`
  that is non-null only where a value is actually set at that entity. Project 9
  sits under OU 11, which overrides variable 4 -- and project 9 reports
  `override: null` for variable 4 while carrying its own overrides on 3 and 8.
  If inheritance leaked into `override` the count would have been far above 21.
- **All three parent kinds carry records.** An account-only fix would have
  found 5 of the 21. The OU set is the largest.
- **Entity ids are per-kind**, so `account/2` and `ou/2` are different parents.
  The import id has to name the kind (`FormatKindParentSlashKey`); a two-part
  `2/4` would collide and `RenderImports` would drop one as a duplicate.

### What the walk costs on other installs

One install does not characterise the product, and the cost here is
`accounts + OUs + projects` requests -- which scales with install size, unlike
most parent-scoped resources, whose parents are only OUs and projects. Measured
across three installs (the production-scale one was read with entity-list GETs
only -- four requests, no child walk):

| install | accounts | OUs | projects | entities | variables | overrides |
|---|---:|---:|---:|---:|---:|---:|
| demo, the one above | 92 | 22 | 37 | **151** | 12 | **21** |
| a second demo | 96 | 129 | 175 | **400** | 0 | **0** |
| production-scale | 1,311 | 52 | 126 | **1,489** | 2 | not walked |

Two things this changes:

- **Accounts dominate on a large install.** 1,311 of the production-scale
  install's 1,489 entities are accounts, and no other resource walks a
  per-account parent set. The walk is still bounded, one-shot, and cheaper than
  the pair probe (2,978 there), but it is the most expensive parent set the tool
  has. `--exclude kion_custom_variable_override` skips it, the same way the docs
  already recommend excluding the shipped policy and compliance catalogs.
- **An install with no custom variables spends the whole walk for nothing.** The
  second demo made 400 child calls and found 0 overrides, because it defines 0
  custom variables. A cheap short-circuit exists -- `/v3/custom-variable` is one
  request and an empty result makes overrides impossible -- but the enumerator
  has no notion of a precondition on another resource's collection, and
  inventing one for a single resource is not obviously worth it. Recorded here
  rather than built.

### The read was never the problem; `ImportState` was

Even with enumeration working, every import would still have failed.
`cvoverride.gtpl` parsed `req.ID` as a bare integer and set only `id` -- into a
`StringAttribute`, and leaving `entity_type`, `entity_id` and
`custom_variable_id` unset, which is exactly what `Read` dispatches on. `Create`
has always built a three-part id (`"<entity_type>/<entity_id>/<custom_variable_id>"`);
`ImportState` now splits it back out and validates both halves.

This is the same class of defect as the `no_read` empty shells below: the
enumeration side looked fine in isolation, and only an end-to-end
plan-and-apply showed it.

### Verified end to end

Against the demo install, with the provider built from this branch:

```
Plan:     21 to import, 0 to add, 0 to change, 0 to destroy.
Apply complete! Resources: 21 imported, 0 added, 0 changed, 0 destroyed.
```

A second `plan` reports **No changes**, so the round trip is clean. Every
generated block carries real attributes -- `entity_type`, `entity_id`,
`custom_variable_id` and one of `value_string`/`value_list`/`value_map` -- not
the empty bodies the `no_read` resources used to produce.

`entity_id` stays out of `rewrite-refs`: `codegen/references.yaml` records it as
unclassifiable because its target depends on `entity_type`. Import blocks
resolve regardless; the generated configuration just holds a literal id.

### The caveat wording was wrong for this shape

`RequireValidField` drops were reported as "record(s) of another kind sharing
the collection". True on the exemption collections, false here -- those 1,791
records are the *same* kind, merely inherited rather than overridden at that
entity. The message now names the discriminator instead of asserting a cause:

```
1099 record(s) in the collection with no "override" set
237 record(s) in the collection with no "ou_cloud_access_role_id" set
```

The drop is still always counted and reported. 1,791 + 21 = 1,812 reconciles
against the cross product exactly, which is what makes the count worth printing
even though it dwarfs the result.

## `no_read` did not mean unreadable

Three resources — `kion_ou_cloud_access_role_exemption`,
`kion_project_cloud_access_role_exemption` and `kion_aws_resource_tag` — used the
`no_read` archetype, whose `Read` is a no-op that echoes state. `ImportState` sets
only the id, so `-generate-config-out` wrote **empty bodies**:

```hcl
# __generated__ by Terraform from "14"
resource "kion_ou_cloud_access_role_exemption" "kion_ou_cloud_access_role_exemption_14" {
}
```

The plan reported `0 to change` because every attribute is Optional+Computed and
null, so this looked green while describing nothing.

`no_read` means the public spec has no **by-id** GET, not that the record cannot
be read. Both exemptions are readable from the private
`/v1/{ou,project}/{id}/cloud-access-role-exemption` collection, and
`aws_resource_tag` from the public flat `/v3/aws-resource-tag`. All three now
declare a `parent_read` in `private_endpoints.yaml`.

Two properties of those collections had to be handled, both measured live:

- **Inherited, not owned.** `/v1/ou/{id}/cloud-access-role-exemption` returns
  every exemption visible to that OU's subtree, so one record comes back under
  many OUs — 329 rows for 22 distinct records, 22 OUs returning 5 distinct sets.
  The id in the path is therefore *not* the record's owner; its own `OUID` is.
  `Parent.ParentIDJSON` tells the enumerator to use it, which is what makes the
  `"<parent>/<id>"` import id resolve.
- **Kind-mixing.** The same collection returns cloud **rule** exemptions
  alongside cloud **access role** exemptions. Only 6 of the 22 OU records were
  the latter, and **0 of 12** on the project side. `Resource.RequireValidField`
  names the discriminator (`ou_cloud_access_role_id` being `Valid`); the drops
  are reported in the run's Reason, never silent.

After the fix, the same install enumerates 6 OU exemptions with their true owners
(`1/5`, `1/7`, `1/24`, `1/25`, `11/4`, `42/26`) and generates real configuration
for each. `reason` is in the schema and in the create body but absent from the
read payload, so it stays unreadable.

**A green `terraform plan` is not proof of coverage.** Check that the generated
configuration has attributes in it.

### Correction: the exemptions read the wrong collection

The section above is a faithful record of what that run measured, but the
collection it measured is not where these resources' records live (#80, #44).
`/v1/{ou,project}/{id}/cloud-access-role-exemption` lists an older,
cloud-rule-linked exemption kind. `POST /v3/ou-cloud-access-role-exemption` and
`POST /v3/project-cloud-access-role-exemption` store somewhere else, which is why
an exemption created through the provider was never found there, and why the
"6 OU exemptions" above were not records these resources could have created.

Both resources now read the parent's cloud access role listing,
`GET /v1/ou/{id}/ou-cloud-access-role` (`data.ou_exemptions`) and
`GET /v1/project/{id}/ou-cloud-access-role` (`data.project_exemptions`). Each
record there is `{id, ou_id|project_id, ou_cloud_access_role_id, reason}`: the
list holds only exemptions the parent itself owns and only this kind, so neither
the owner key nor the kind filter is needed, and `reason` is stored and read
back. The lists are filled only for a caller with Browse All Cloud Access Roles
and Manage OU / Manage Project on the parent.

This was derived from the API's source, not from a live run. It has not yet been
re-measured with `kion-import --probe`.

## Reproducing

```sh
make build-tools
export KION_APIKEY=…
./bin/kion-import --url https://kion.example.com --probe          # read outcomes only
./bin/kion-import --url https://kion.example.com --out imports.tf # generate
```

Narrow a run with `--include` / `--exclude` / `--selection`; see
[the importing guide](guides/importing.md). Importing everything is rarely what an operator
wants: Kion's shipped policy and compliance catalogs dominate the count. On the
install above, excluding `kion_azure_policy`, `kion_compliance_check`,
`kion_compliance_control` and `kion_compliance_standard` takes the run from
48,355 records to 10,131.

Add `--api-prefix ""` for a localhost app serving the API at the root.
