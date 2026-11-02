# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Unreleased changes are not listed here — they live as fragments in
`.changes/unreleased/` until a release is cut. Run `make changelog-preview` to
see them, and `make changelog-new` to add one.


## [1.8.0] - 2026-11-02


### ENHANCEMENTS

- kion-import: retries a read that gets a 401, 429, 502, 503 or 504, twice, after 1s and then 3s. The old importer retried a 401 because Kion occasionally rejects a valid key. A 401 that outlasts its retries marks the key as bad, and later reads fail at once, so a wrong key still fails quickly.

### BUG FIXES

- The kion_saml_group_association data source works for configurations written for the old provider again. It takes a required idms_id and an optional filter block, and returns the matching group associations in list, as the old provider's 0.3.36 did. Before this, it fetched only one association by id, so a 0.3.36 configuration failed. The by-id form is gone. To pick one association, set idms_id and filter on id.
- kmigrate: reports kion_azure_account, kion_custom_account and kion_gcp_account blocks that lack project_id or start_datecode. Both were optional in the old provider and are required now. Before this, kmigrate said nothing and terraform plan failed. The report notes that an account in the account cache belongs in kion_account_cache, because adding project_id to it would move it into that project.

### NOTES

- The migration and import guides are now Registry guides: "Upgrading from 0.3.x" and "Importing an existing Kion install". The upgrade guide adds the Terraform 1.0 requirement and a breaking-changes section, which includes how to move an account in the account cache to kion_account_cache. CHANGELOG.md now also carries the previous provider's history, 0.3.36 and earlier.

## [1.7.1] - 2026-11-02


### BUG FIXES

- Migrating kion_ou_cloud_access_role and kion_project_cloud_access_role from the old provider's 0.3.35 or later no longer loses the role-type attributes 0.3.35 added (cloud_access_role_type_id, aws_iam_role_trust_policy, aws_trusted_account_numbers, aws_trusted_services, aws_partition, aws_create_instance_profile, aws_session_tags, and the project role's cloud_provider_ids). The state upgrader had been setting them to null. aws_session_tags also changed from a map to a list of tag objects, so kmigrate now rewrites it and the state upgrader converts the stored map.

## [1.7.0] - 2026-10-21


### ENHANCEMENTS

- kion-import: flags records whose generated configuration will not apply to another install as-is (built-in permission schemes, the Administrators group, system- custom variables, and funding sources with no OU). They are listed in the report under "Will not apply to another install as-is", and each gets a "# not portable:" comment above its import block.

### BUG FIXES

- kion_user_group: setting user_ids, viewer_user_ids or owner_user_ids to an empty set no longer fails the apply with "inconsistent result after apply: was cty.SetValEmpty, but now null". Kion returns an empty membership as null; the read now treats null as an empty set (#147).
- kion-import: kion_billing_rule no longer fails to list with 400 "invalid rule type" on an install that has allocation or custom line item billing rules. Kion's v3 list rejects the whole page when it holds a rule type the v3 shape cannot carry, so kion-import now lists only the adjustment types kion_billing_rule manages (#149).

## [1.6.0] - 2026-10-21


### NEW RESOURCES

- kion_billing_source_azure: manage Azure billing sources (CSP, EA, MCA and their Government and Secret variants). Kion publishes no public API for them, so the resource uses the same private routes as Kion's own UI. The client secret is never returned by Kion, so it is not read back, and set skip_validation to create or update one without Kion first testing the Azure credentials. kion-import imports existing Azure billing sources, and rewrite-refs resolves payer_id references to them.
- kion_billing_source_anthropic: manage Anthropic billing sources (Kion 3.16 or later). Kion validates the Admin API key with Anthropic when the billing source is created, so it must be a real sk-ant-admin key; organization_id is read back from Kion. Neither key is ever returned, so neither is read back. kion-import imports existing Anthropic billing sources, and rewrite-refs resolves payer_id references to them.

### BUG FIXES

- A kion_app_role deleted outside Terraform is detected and planned for recreation. Kion answers a read of a deleted app role with a 500 rather than a 404, which made every refresh fail; the provider now checks the public app role list before reporting an error.
- An OpenID IDMS group association deleted outside Terraform is now detected and planned for recreation. Kion's by-id read keeps returning deleted associations, so the provider confirms the record against its IDMS's association list, which leaves them out.
- An imported kion_idms_open_id_group_association keeps open_id_id and update_on_login. Import previously left both unset, so the next plan wanted to replace the association.
- make sweep no longer stops at the IDMS group association sweepers on an install with any non-SAML IDMS. Kion's association list serves SAML only and crashes on other IDMS types, so both sweepers now walk SAML IDMSs only.
- Changing an account's create-time settings after creation now produces a plan warning instead of a silent no-op. On kion_account, kion_aws_account and kion_account_cache: account_type_id, payer_id, commercial_account_name, gov_account_name, create_govcloud and the AWS organizational unit (plus account_number on the cache). Kion has no way to change these on an existing account, and replacing the resource would create a different AWS account.

### NOTES

- kion_aws_resource_tag has now been verified against a live install: create, refresh, import and detecting a tag deleted outside Terraform all work. It had only unit-test coverage, because no install checked had a tag.

## [1.5.1] - 2026-10-20


### ENHANCEMENTS

- kmigrate --version reports the release it was built from. kmigrate rewrites configuration for one provider release's schema, so it should match the provider version you are migrating to.

### BUG FIXES

- `kion-import -h` no longer prints the value of KION_APIKEY as the --api-key default. The variable is still read when --api-key is not given.
- kion-import exits non-zero and writes nothing when every read fails, as a rejected API key or an unreachable install causes. It used to exit 0 and write an imports.tf holding only "Not imported" comments.
- kion-import's generated configuration now requires the provider release that built it, as a lower bound, instead of pinning exactly 1.0.0, which lacks resource types the importer emits. The modules under modules/ now require ">= 1.0.0" instead of exactly 1.0.0, so they work with any release and alongside an imported configuration. `kion-import --version` reports the release.
- kion-import reads kion_billing_source_govcloud again. It listed billing sources through GET /v3/billing-source, which only accepts POST, so the type always reported an error; it now lists through /v4/billing-source.
- kion_automation_policy accepts engine 1 (Cloud Custodian), so existing Cloud Custodian policies can be imported. Kion still rejects creating one with engine 1, and says so itself.
- Records Kion holds without owners -- built-in compliance checks and standards, Kion-managed Azure and IAM policies, built-in cloud rules -- can now be imported. The "at least one owner" rule on the 20 resources that have one now applies when creating a resource, which is when Kion enforces it, instead of on every plan.
- kion_project reads its owners. An imported project kept no owners, and an owner change made outside Terraform never showed as drift; owners now come from the project's owner permission mapping on every read.
- kion_project reads its permission_scheme_id. No public read returns it, so an imported project had none and a scheme changed outside Terraform was never seen; it is now read from the project's edit payload.
- kion-import no longer imports every billing source as an AWS, a GCP and an OCI billing source at once. Each typed billing source now keeps only the payers of its own kind from the shared /v4/billing-source collection.
- kion_billing_source_aws reads linked_role and account_type_id, so an imported AWS billing source plans cleanly and a change to either made outside Terraform shows as drift.
- kion_billing_source_gcp can be imported. Read now loads big_query_export and account_type_id instead of carrying them from prior state, which on import left big_query_export without a type and failed with "Invalid GcpBillingAccountCreateValue Attribute Type".
- kion-import rewrite-refs turns payer_id into a reference to an AWS, GCP or OCI billing source, not only a custom one. Those references stayed as the source install's literal ids, which is what moving a configuration between installs has to avoid.

## [1.5.0] - 2026-10-15


### NEW RESOURCES

- `kion_oauth2_client` resource and data source: manage Kion OAuth2 clients (Kion 3.17.3+). The client secret is captured at creation and is not recoverable after import. Changing `is_public` replaces the client.

### ENHANCEMENTS

- Data sources whose Kion API exists only from a given release (`kion_oauth2_client`, `kion_ou_note`, `kion_scope`, `kion_scope_criteria`) now fail with a clear Kion version diagnostic on older installs instead of an opaque API error.

### BUG FIXES

- List attributes that are optional and computed are no longer sent as an empty list when left unconfigured on create, so Kion applies its own default instead of clearing the value (for example, an OAuth client's allowed_scopes). A list removed from configuration is still cleared.

### NOTES

- The provider now compiles against Kion API v3_17 (kion-sdk-go 0.14.0). Kion 3.14 through 3.17 remain supported through the runtime version gates; no configuration or state changes.

## [1.4.0] - 2026-10-02


### BREAKING CHANGES

- `kion_cloud_rule.source` and `kion_project_cloud_access_role.policytype` are removed. Both were filter parameters of the list endpoints that leaked into the resource schemas; no request ever carried them, so setting either had no effect. A configuration that sets one must drop it. Neither existed in the previous provider, so migrated configurations are unaffected.

### ENHANCEMENTS

- `kion_billing_source_aws` and `kion_billing_source_gcp` can now be updated in place. Before, any change failed with "Not supported". The update sends the current record back with the planned values applied, so settings Terraform does not manage are kept. Changing `account_type_id` (both) or the deprecated `only_dbr` (AWS) now forces replacement, because the update cannot change them. `key_id` is now marked sensitive.

  `kion_billing_source_aws` now rejects `key_id` without `key_secret` (or the reverse) at plan time. The API uses an access key pair only when both halves are set, so setting one alone was accepted and then ignored.

  `kion_billing_source_gcp` now reads `billing_start_date` back correctly, and its create now sends `big_query_export`. Before this fix, every refresh failed and the BigQuery export settings were dropped on create.

### BUG FIXES

- `kion-import` no longer reports `kion_idms_group_association` and `kion_saml_group_association` as failing on every non-SAML IDMS. The API serves `GET /v3/idms/{id}/group-association` only for a SAML IDMS and fails for every other type, so the import manifest now marks that collection's parent with `"match": {"idms_type_id": "3"}` and `kion-import` skips non-matching parents without a request. An install with no SAML IDMS reports both resources `empty` rather than `error`. Group associations exist only on a SAML IDMS, so no records are lost.
- `kion_account` and `kion_aws_account` now create through `POST /v3/account-cache/create?account-type=aws`. They previously requested a path that does not exist, so every create failed with a 404 (#102).

  Setting `account_number` now adopts an existing AWS account instead of failing with "not yet supported". Without `project_id` the account is added to the account cache (`account_type_id` defaults to 1, AWS Commercial, when unset); with `project_id` it is added to that project, and `start_datecode` is required. If you set attributes that only apply to creating a new account (`aws_organizational_unit`, `commercial_account_name`, `create_govcloud`, `gov_account_name`, or `use_org_account_info` when adding to the cache), you now get a warning that they were not sent. Creating a new account warns the same way for `skip_access_checking` and `use_org_account_info`, which the new-account route does not accept.
- `kion_ou_cloud_access_role_exemption` and `kion_project_cloud_access_role_exemption` now read from the parent's cloud access role listing (`GET /v1/ou/{id}/ou-cloud-access-role` and `GET /v1/project/{id}/ou-cloud-access-role`) instead of a collection that never contains them. Previously every refresh dropped the exemption from state, so the next plan proposed creating it again and each apply left another exemption behind (#80). `reason` is now read back as well; it is stored by the API, and the earlier note that it was discarded was wrong (#44).

  Reading these exemptions requires Browse All Cloud Access Roles and Manage OU (Manage Project for a project exemption) on the parent. A caller without those sees no exemptions, so a managed exemption refreshes as deleted. `kion-import` enumerates them from the same listing.
- The `kion_ami` data source now exposes `expires_at`, and `kion_gcp_iam_role` now exposes `gcp_role_launch_stage`, both as top-level attributes and in each `list` object. The `list` objects of `kion_project_cloud_access_role` now carry `created_at`. These attributes were previously left out of the data sources because the generator could not project their SDK types.
- The `kion_account_linkage` data source could not read a linkage that had not been finalized, which is every active one: Kion returns `delete_finalized_at` as null and the SDK decodes it as a non-nullable string, so both the `id` lookup and the `filter` listing failed on decode. The data source now reads `/v3/account-linkages` directly, as the resource already did, and tolerates the null.

  In `list`, a field Kion returns as null (`delete_finalized_at`, `deleted_at`, `updated_at`, `payer_id`, `user_id`, and so on) is now null rather than an empty string or zero.
- Changing an attribute the API cannot change in place no longer applies, reports success and leaves the value where it was. Each now forces replacement: `aws_iam_role_name`, `aws_iam_path`, and `ou_id`/`project_id` on `kion_ou_cloud_access_role` and `kion_project_cloud_access_role`; `kion_iam_policy.aws_iam_path`; `name` and `type` on `kion_custom_variable`; `idms_id` on `kion_saml_group_association` and `kion_idms_group_association`; `open_id_id` on `kion_idms_open_id_access_rule` and `kion_idms_open_id_group_association`; `ou_id` and `create_user_id` on `kion_ou_note`; the resource group on `kion_azure_arm_template`; `email`, `gcp_project_id` and `unique_id` on `kion_gcp_service_account`; and the cloud-side identity of an account (`kion_azure_account.subscription_uuid`, `kion_gcp_account.google_cloud_project_id`, `kion_custom_account.account_number`).

  Attributes that do have their own endpoint now take effect on update: `kion_project` owners (replaced in full through `POST /v3/project/{id}/owner`), `kion_user.user_group_ids`, and `user_ids`/`user_group_ids` on `kion_project_enforcement` and `kion_funding_source_enforcement`. `enabled` on both enforcements and `archived` on `kion_project` are applied when set at create, where the create body has no field for them.

  Changing `kion_ou.parent_ou_id` now moves the OU, with its child OUs and projects, instead of doing nothing; this uses Kion's internal OU move route and requires global admin. Changing `project_id` on `kion_azure_account`, `kion_gcp_account` or `kion_custom_account` now moves the account through `POST /v3/account/{id}/move`: Kion gives the account a new id, which the resource adopts, and financial history stays with the old project. A completed move is recorded in state immediately, so a failure later in the same update does not leave state naming the old parent or account id.

  State migrated from the previous provider no longer fails to decode for a `kion_project` that recorded `move_ou_settings`: the new block gained `spend_plan_setting`, which the upgrader now adds as null.

### NOTES

- Acceptance tests now come from a single generator. The `kion_azure_arm_template` and `kion_azure_role` tests skip unless `KION_ACC_AZURE_PAYER_ID` is set (and `KION_ACC_AZURE_SUBSCRIPTION_UUID` for the ARM template), and the `kion_gcp_service_account` test reads the account email from `KION_ACC_GCP_SERVICE_ACCOUNT_EMAIL` and skips when it is unset. Previously these tests ran against installs that could not satisfy them.

  The retired `kgen tests` command and `make tests-gen` target are removed.

## [1.3.0] - 2026-09-29


### FEATURES

- Acceptance tests are generated, for every resource that can have one: 71 of 72, gcp_regions being data-source-only. Before this only two resources had a test_values entry, so kgen produced 4 of 111 acceptance-test files and the rest were hand-maintained stale output -- five still carried TODO stubs from a template revision that had since grown real Exists/Destroy checks.

  The generator covers the entity, parent_list, blended, association, raw_http and no_read archetypes, each probing the API the way that archetype actually reads -- get by id, list the parent and match, scan the collection, or call the private route. Resource-specific cases are kept verbatim as companion templates, so a wipe reproduces them. test_values.yaml gains prereqs (HCL emitted ahead of the resource under test), fixtures (a shared HCL builder in internal/acctest), extra_args (Go expressions for a value HCL cannot state literally) and an hcl: prefix for an unquoted expression.

  Checks compare against the value the configuration used rather than merely asserting an attribute is set, and the update test checks the updated values rather than only the id, so a resource that accepts input and discards it -- or an update that silently no-ops -- now fails. A test needing an install-specific value reads it through acctest.RequireEnv and skips with the variable name when it is unset; docs/TESTING.md lists them per resource. A new CI gate (`make crud-check`) fails when internal/service/ does not match freshly generated output, which is what let the hand-maintained tests drift unnoticed.
- `kion_ou_cloud_access_role` and `kion_project_cloud_access_role` manage cloud access role types. Before this the public API created every role as a User role regardless, and the six fields that describe any other type were not exposed: `cloud_access_role_type_id` (1 User, 2 Custom Trust, 3 Account, 4 Service), `aws_iam_role_trust_policy`, `aws_trusted_account_numbers`, `aws_trusted_services`, `aws_partition` and `aws_create_instance_profile`. Requires kion-sdk-go 0.13.0.

  Changing the role type forces replacement. The create body carries it and the update body does not, so an in-place change would be accepted and ignored, leaving the role as it was with the plan reporting success.

  Each non-User type requires the field naming who may assume the role, enforced at plan time rather than as a 400 at apply: type 2 needs the trust policy, type 3 the account numbers, type 4 the services. `cloud_access_role_type_id` is validated to 1-4 and `aws_partition` to the four real AWS partitions.

  The fields exist in Kion 3.15.13, 3.16.5 and 3.17.1, and are absent from the releases between (3.16.0-3.16.4 and 3.17.0). Setting one against a release that lacks it fails at plan naming the version ranges, because the API accepts unknown body fields, ignores them, and answers 200.

### BUG FIXES

- kion_user, kion_webhook, kion_service_catalog and kion_scope_criteria were skipped by the generator entirely, so every committed file in those four packages was stale output that nothing reproduced and no drift gate could see.

  The first three delete through a private route the SDK has no typed method for. RawDeletePath exists for exactly that, but resolution treated the missing typed op as fatal before it was consulted; it is now fatal only when no private route is declared. kion_scope_criteria declares delete_retains, meaning the API cannot remove the record at all and the generated Delete forgets it with a warning, yet resolution still demanded a delete op -- making the archetype's own declaration unbuildable.

  This also lands the owner sync kion_webhook and kion_service_catalog were already declared to have: the declaration could not reach code the generator never produced, so owner_user_ids and owner_user_group_ids were accepted and discarded on update.
- internal/service is reproducible from a bare wipe. Several files were not: service_package.go was scaffolded and then hand-edited for type aliases, a resource-only archetype left the scaffolder's data source stub behind and registered it, two version gates were orphans no generator wrote any more, and the accounthelper and gcp_regions packages failed because generation did not create the directory it was writing into.

  service_package.go is generated for every archetype now, with aliases declared rather than edited in; a data source the archetype did not emit is pruned; and `make regen-from-scratch` deletes internal/service outright rather than only the files carrying a generated header. The weaker check had passed for months while 107 hand-written acceptance tests sat in the tree unreproduced, because a file with no header was invisible to it.
- Attribute version gates take a set of ranges rather than a single floor. Kion availability is not monotonic and a backport lands at a different patch on each line: the cloud access role types are in 3.15.13 and 3.16.5 and absent from 3.17.0, so one range spanning both would accept 3.16.0 through 3.16.4, where the field does not exist. The API does not reject unknown body fields, so it accepts such a field, ignores it, and answers 200; Terraform records success and the value never lands.

  The derivation had a second defect that no upper bound alone would have fixed. It took the newest tracked version's fields as the set to consider, so a field the newest release had dropped never entered the map and got no gate at all. It now takes the union across versions and records the first and last release carrying each field, bounded by the last release carrying the operation, so a dropped attribute stays distinguishable from a dropped endpoint.

  `attrMinKionVersion` becomes `attrKionVersions`, typed `map[string]framework.AttrVersions` where each entry is a list of ranges, Min inclusive and Before exclusive. The diagnostic names every range it accepts, so an instance sitting in a gap between two floors is told both.

  Derivation reads the SDK's per-minor packages, so on its own it can only say `3.NN.0`. Patch floors are authored in `codegen/attr_version_overrides.yaml`, which replaces the derived ranges for an attribute; the cloud access role types are the first entries there.
- Changing `ou_id` on a `kion_project` moves the project. It used to do nothing: `ou_id` is settable with no RequiresReplace, and the update body carries no `ou_id`, so the plan showed a change, the PATCH could not express it, the API answered 200, and the next refresh showed the same diff again. A project cannot simply be recreated instead, since it holds accounts, budgets and users.

  `POST /v3/project/{id}/move` became public in Kion 3.16 and is what the UI uses. Update calls it before the PATCH, so a failed move does not leave the other fields updated against the old OU.

  `move_ou_settings` now does something. It was accepted and discarded, and it carries the three choices the move requires: `cloud_rule_setting` (`convert` or `remove`), `financial_setting` (`preserve` or `move`) and the new `spend_plan_setting` (`keep` or `create`). None is inferable, because each decides what happens to data the project already has, so an omitted setting takes the option that keeps what exists rather than the one that discards it. All three validate their allowed values at plan time.

  This retires `project.ou_id:update` and `project.move_ou_settings:update` from `codegen/unbound_attributes.yaml`, where both sat as untriaged debt. `move_ou_settings:create` stays and is no longer marked untriaged: a project is created in an OU rather than moved into one, so there is no move to configure.
- Changing owners on an existing resource reaches the API. On `kion_ami`, `kion_azure_arm_template`, `kion_azure_role`, `kion_cft`, `kion_compliance_check`, `kion_iam_policy` and `kion_service_control_policy`, an update to `owner_user_ids` or `owner_user_group_ids` was accepted and discarded: the update body has no owner fields, so the plan applied, the API answered 200, and the owners stayed as they were. The same was true of `car_restricted_user_ids` and `car_restricted_user_group_ids` on `kion_azure_role`, `kion_gcp_iam_role` and `kion_iam_policy`. Update now syncs each list through the resource's add/remove endpoints, adding what the configuration gained and removing what it lost. The aliases `kion_aws_iam_policy` and `kion_aws_cloudformation_template` share the fix.

## [1.2.0] - 2026-09-15


### FEATURES

- `kion_automation_policy` manages Kion automation policies, which had no Terraform representation at all -- neither provider exposed them, so scheduling work had to go through a generic REST provider against the private API by hand. Every endpoint is internal `/v1`: the public spec carries an `AutomationPolicy` component but no path that uses it, so the request and response shapes are transcribed in `codegen/spec_additions.yaml` the way `funding_source_note`'s already were. The schema is the flat view a practitioner writes -- `name`, `engine`, `enabled`, owner and cloud-rule ids, a `scheduled_frequency` object and a `cloud_provider_policies` list -- while the wire body is an envelope that nests the policy's own scalars under `automation_policy` and renders `resources` and `regions` as arrays of one-key objects. That nesting is why it has its own archetype rather than reusing `raw_http`, which supports scalar attributes only. `name` and `cloud_provider_policies` are Required because the backend enforces both (a create with an empty policy list is rejected with "At least one cloud service provider policy must be selected"), and `engine` validates to 0 at plan time, since the API defines Cloud Custodian as 1 but rejects it. `scheduled_frequency`'s `start_date` and `end_date` cross the wire as `sql.NullTime` objects rather than bare strings -- the defect that made `project_note` and `dashboard` undecodable -- and are read tolerantly and written in the form the backend parses.

A matching `kion_automation_policy` data source looks a policy up by `id` or by `filter` blocks, the same dual-mode shape the generated list data sources use. It pages the index rather than reading it once: the endpoint serves ten records unless asked otherwise, so a single request would report the first ten policies as the whole collection, and a scheduling setup of a dozen or more policies is the normal case rather than an edge one. The walk is now `internal/paging`, shared with `kion-import`, which had the only other implementation of it.

Attaching policies to a cloud rule already worked through `kion_cloud_rule.automation_policy_ids`; now that they are a managed resource, `kion-import rewrite-refs` turns those literal ids into references, and `kion-import` enumerates policies into import blocks.

### BUG FIXES

- `make generate-schemas` runs again. It aborted with six "path not found in OpenAPI spec" errors and wrote nothing, so no resource's schema could be regenerated at all. Five were private update/delete routes that `generator_config.yaml` names and nothing typed -- tfplugingen resolves every path a resource declares, not just the create and read it builds the schema from -- and are now declared in `spec_additions.yaml` with no body, since they only need to resolve. The sixth was `azure_account.create`, which derivation emitted as the literal query string `/v3/account?account-type=azure`; the spec models those factory-creates as synthetic `/__qs__/` segments, which its `custom` and `google-cloud` siblings already carried. Nothing caught this because the spec is gitignored, so CI cannot run schema generation and `make ci` does not try. Regenerating afterwards reproduces every committed schema byte-for-byte, so the breakage was confined to the generator's ability to run.
- Generated examples are configuration Terraform will actually load. Two defects made them otherwise. A list- or set-nested attribute rendered as `attr = { … }` rather than `attr = [{ … }]`, which Terraform rejects with "Inappropriate value for attribute: list of object required" -- wrong for every resource carrying one (`kion_project`, `kion_cft`, `kion_aws_cloudformation_template`, `kion_idms_open_id`, `kion_ou_cloud_access_role`, `kion_permission_scheme` and `kion_project_cloud_access_role`), and unnoticed because those attributes are all Optional and so were emitted commented out; `kion_automation_policy` has a Required one, which renders live and made it visible. A map-nested attribute still renders as a bare object, because its keys are the practitioner's to choose and the schema cannot supply one. Separately, placeholder values now satisfy the attribute's own validators, by trying candidates rather than guessing: `engine` on `kion_automation_policy` emitted 1 against an enum accepting only 0, and the datecode attributes on `kion_azure_account`, `kion_custom_account`, `kion_gcp_account`, `kion_billing_source`, `kion_billing_source_aws`, `kion_billing_source_gcp` and `kion_billing_source_oci` emitted the literal example string against a four-digit-year-and-month pattern. Several of those are Required, so the example was live configuration the provider rejected at plan time rather than a commented hint. Finally, `kgen service` now indents the import it adds to `internal/provider/service_packages.go`; it wrote the line hard against the margin, so scaffolding a resource left the repository failing its own gofmt check.
- The committed provider schema snapshot (`codegen/schema_snapshots/new.json`) is current again. It had not been refreshed since August and was missing five resources -- the four `kion_billing_source_*` variants and `kion_automation_policy` -- which matters because it is the ground truth for three completeness gates, and a resource absent from it passes them vacuously. Refreshing it surfaced four real problems that had been invisible: `cloud_rule_ids` and `key_id` were unclassified foreign keys (the first is a `kion_cloud_rule` reference, the second an AWS access key that must never be rewritten, like `portfolio_id`); `kion_project_note.create_user_id` was still listed as a read-only drop, so `kmigrate` would have stripped it out of a working configuration even though the new provider requires it, exactly as the old one did; and `kion_app_config` and `kion_project_enforcement` widen attributes from list to set with no `state_upgrades.yaml` entry, so customer state had no declared upgrade path. `automation_policy_ids` also moved from `not_references` to a real reference now that `kion_automation_policy` exists, so a cloud rule's attached policies rewrite into references rather than staying literal ids.

## [1.1.1] - 2026-09-15


### BUG FIXES

- Published releases carry their notes again. Every release from v1.0.0 through v1.1.0 shipped with an empty body on GitHub, despite `release.yml` passing `--release-notes` the changie version file and verifying that file exists first. `.goreleaser.yml` also set `changelog: disable: true`, meaning "we supply our own notes" -- but `--release-notes` is read *by* the changelog pipe, so disabling it makes `Skip()` return true, `Run()` never execute, and the flag be discarded without a warning. The two settings read as reinforcing and in fact cancel. Removing `disable` restores the notes and does not reintroduce generated ones, because the flag already suppresses them: `Run()` loads the file, assigns it, and returns before the changelog is built. It also means a missing or unreadable notes file now fails the release rather than quietly publishing an empty one. The four affected releases had their notes backfilled from the version files already in the repository; their artifacts, checksums and signatures were not touched.
- `kion-import` now ships in the release. It was added as a headline feature in v1.1.0 and documented as a command to run, but goreleaser built only the provider and `kmigrate`, so the only way to get it was a source checkout -- and it cannot be `go install`ed, because `go.mod` declares a bare module path with no domain to install from. It is built for the same platforms as `kmigrate` (linux/darwin/windows on amd64/arm64), being a one-shot tool run against an install rather than something Terraform downloads per platform, and needs no companion file in the release because `codegen/embed.go` compiles `import_manifest.json` into the binary. Its archives are covered by `SHA256SUMS`, so the existing detached signature covers them too. `docs/IMPORTING.md` gained a section on obtaining it, which it never had; it noted the tool only as a command, with no indication of where it came from.

## [1.1.0] - 2026-09-15


### BREAKING CHANGES

- Two attributes changed in ways that reject a configuration which previously loaded. `kion_app_role.disabled` is now read-only: it was `Optional` and `Computed`, but nothing in the request body carried it, so setting it was accepted and discarded -- the API decides what is disabled. A configuration that sets `disabled` must drop the line; it never had any effect. `kion_project_note.create_user_id` is now `Required`: it was `Computed`, so the provider sent no author and Kion rejected every create, meaning the resource could not be used at all. A configuration that already has a `kion_project_note` block must add `create_user_id`, naming the user the note is attributed to. This release is a minor rather than a major because neither change can break a configuration that worked: `disabled` did nothing, and `kion_project_note` could not be created. Only a configuration that set an attribute which never functioned will now fail, and it fails at plan time with a message naming the attribute rather than at apply.

### FEATURES

- The provider now tracks Kion 3.14-3.17, matching `kion-sdk-go` 0.10.0. That release adds `v3_17` and removes `v3_12` and `v3_13`, so the SDK ships exactly the supported window: the current release plus three back. Version gates are regenerated against it -- a resource present since 3.13 or 3.14 now spans the whole window and needs no gate at all, and a resource introduced in 3.17 gates at `>= 3.17` rather than going ungated. The tracked-version set is also declared once now instead of twice; the SDK had the same set spread across nine places and a missed one caused a two-week pipeline outage, so the duplicate was worth removing while here.
- `kion-import` brings an existing Kion installation under Terraform management. It walks every resource the provider supports and writes a file of `import` blocks, which you hand to `terraform plan -generate-config-out` to generate the configuration and to `terraform apply` to build the state. The provider's own `Read` and `ImportState` do the serializing, so the generated state is correct by construction. The tool appends `/api` to the URL you give it, so `--url https://kion.example.com` is enough; `--api-prefix ""` removes it for an app served at the root. Resources it cannot read are reported with the reason rather than omitted, and `--probe` reports what each resource returned without writing anything.
- `kion-import rewrite-refs` turns the literal foreign keys in a generated configuration into references between resource blocks: `ou_id = 11` becomes `ou_id = kion_ou.kion_ou_11.id`. `terraform plan -generate-config-out` can only write literals, since Terraform has no way to know the value points at anything; the result is accurate for the install it came from but does not express its own dependency graph, so destroy-and-reapply does not order correctly and recreating a parent does not cascade — and against a different install the literal either dangles or silently binds to whatever record holds that id there. With references, Terraform resolves the value from the id it created, which is why moving a configuration between installs needs no id map. The id-to-label index comes from the import blocks, because the generated configuration contains no ids at all (they are Computed, so `-generate-config-out` omits them). Which attributes are foreign keys, and what each points at, is authored in the new `codegen/references.yaml`: the name is not enough to tell — `payer_id` is a billing source, `ugroup_ids` is a user group, and `portfolio_id` is an AWS id that must never be rewritten, sitting right beside `internal_portfolio_ids` which is a `kion_service_catalog`. A test fails when any new foreign-key-shaped attribute is left unclassified, so the table cannot silently rot as the provider grows (`make references-check`). Ids whose referent is not managed in the configuration are left as literals and reported by target type rather than silently kept — that boundary is closed by importing the referent too, or with a data source or a variable.
- `kgen import-modules` rewrites the configuration `terraform plan -generate-config-out` writes -- bare `resource "kion_*"` blocks -- into `module` calls against the modules under `modules/terraform-kion-*`, closing the gap between importing an install and managing it through the modules. The attribute-to-variable mapping comes from a new `modules/module_manifest.json`, written by `make modules`, rather than from `variables.tf`: `kgen module` renames any attribute colliding with a `module` block meta-argument, so `kion_cloud_rule`'s `source` becomes `cloud_rule_source` and `kion_compliance_program`'s `version` becomes `compliance_program_version`, and guessing those wrong does not fail cleanly -- Terraform reads a stray `source = "aws"` as the module's own source address. A resource block whose type has no manifest entry keeps its type, labels and attributes, and an attribute with no module variable is dropped and reported rather than guessed at. References are retargeted rather than broken: `kion-import rewrite-refs` emits `<type>.<label>.id` and every generated module exposes an `id` output, so `ou_id = kion_ou.kion_ou_11.id` becomes `ou_id = module.kion_ou_11.id` -- in unconverted blocks too, which would otherwise be left naming a resource that no longer exists. A traversal into any other attribute of a converted resource cannot be retargeted (a module exposes outputs, not attributes) and is reported with a non-zero exit rather than emitted as configuration Terraform will refuse to load.
- Cross-field API constraints are now enforced at plan time instead of surfacing as an API error at apply. Twenty resources require at least one of a pair of attributes -- `owner_user_ids`/`owner_user_group_ids` on most, `user_ids`/`user_group_ids` on the enforcements, `role_permissions`/`role_denials` on `kion_gcp_iam_role` -- and neither is `Required` on its own, so the schema could not express it. Configuring neither produced `Bad Request: Key: 'UGroupCreate.OwnerUserIDs' Error:Field validation for 'OwnerUserIDs' failed on the 'atLeastOneFieldPresent' tag`, naming a Go struct field the user never wrote; it now produces `At least one of these attributes must be configured: [owner_user_group_ids,owner_user_ids]` before any request is made. The constraints are authored in `codegen/config_validators.yaml` rather than derived: the OpenAPI spec carries almost none (3 enums, 16 patterns and 5 maxLength across 474 schemas), so the source is the portal's own `validate:"atLeastOneFieldPresent=..."` struct tags. Attribute names are resolved per package, because the API body's spelling is not always the provider's. Two tests guard it -- one that declaration and generated code agree in both directions, one that every named attribute exists in the resource schema, since a validator naming a missing attribute is a runtime panic rather than a compile error.

### ENHANCEMENTS

- `kion-import` can now be told what to import. `--include` and `--exclude` take resource type names, `--selection` reads the same lists from a YAML file so a repeated run can keep them under version control with comments, and `--list-types` prints what this build supports without needing credentials. A name matching nothing is an error rather than a silent no-op. Importing everything is rarely what an operator wants: one production-scale install enumerates 53,513 records, of which 35,690 are Kion's shipped Azure policy and compliance-check catalogs.

### BUG FIXES

- `make modules-test` no longer tests a stale provider. It copies `modules/` into a scratch tree, and a developer who has run terraform directly in `modules/` leaves a `.terraform/` directory and a lock file behind; copied along, terraform reuses the provider cached there instead of the one just built. The run then silently measures an old binary -- locally 12 of 72 modules failed against a schema the working tree had already changed, while CI, which generates a fresh tree, passed all 72. The copy now drops those artifacts.
- `kion_webhook`, `kion_user` and `kion_service_catalog` now delete the record instead of only dropping it from state. The public spec publishes no delete for any of them, so the generator emitted a `Delete` that raised a warning and returned -- every record a configuration created survived `terraform destroy`, left on the install with nothing in state pointing at it. A private delete route exists for each and works; verified on a live install, create then `GET 200` then `DELETE 200` then `GET 404`. The entity archetype now renders a raw-HTTP delete when `codegen/private_endpoints.yaml` declares one and the public spec does not, the same way the blended archetype already did, and the sweeper generator counts a private delete as a delete. Where a sweeper still cannot be registered the stated reason is now the real one -- no enumerable collection -- rather than the delete that does in fact exist.
- A renamed attribute is no longer silently unbound. `codegen/renames.yaml` gives an API field a different provider name, but the CRUD generator matched a request body field to a model attribute by the API's own name and never read that file -- so a renamed attribute stopped matching and the bind was skipped with no diagnostic, leaving the provider accepting the value and discarding it. `kion_project_enforcement.user_group_ids` was affected: setting it did nothing, and the value never reached the API. With the binder now consulting the rename map, all three enforcement resources also expose the list under the same name -- `kion_ou_enforcement` and `kion_funding_source_enforcement` previously surfaced the API's `ugroup_ids` spelling while `kion_project_enforcement` used `user_group_ids`, so one concept had two public names depending on which resource you were writing.
- The internal-reference gate now also scans the commit messages being pushed, and runs as a pre-push hook. It previously matched tracked file contents and paths only, which let backend source paths, internal environment names and production entity counts reach a public history in commit messages -- where the only remedy is rewriting published commits, so catching it before the push is the whole point. `make refresh-spec` is also fixed: a mangled shell line made it fail with a syntax error before doing anything, which is why nobody had noticed that its default spec path was wrong too.
- A create whose response carries an explicit id of `0` now fails instead of recording it. The envelope id path checked that the field was present but not that the value was usable, so an API returning `0` produced a resource whose state id was `"0"` -- a record that exists on the install and that Terraform can never read, update or delete. This is the SDK-side half of the same defect fixed for the raw-HTTP path, and it now routes through the same `errs.RawCreatedID` guard. Five resources use the envelope path: `kion_app_api_key`, `kion_compliance_control`, `kion_project_line_item`, `kion_scope` and `kion_user`.
- An explicitly empty id list on `kion_compliance_control` no longer fails the apply. Kion stores an empty array as null and returns null, so a configured `[]` came back as a null set and Terraform rejected the result: `.arm_template_definition_ids: was cty.SetValEmpty(cty.Number), but now null`. The read now collapses both to an empty set, the same treatment the permission mappings already had. A new `empty_collections` archetype knob declares which attributes behave this way, per resource, rather than assuming every collection does.
- `kion_compliance_control` can now be imported. `program_id` is not in the read payload and cannot be recovered afterwards, yet delete addresses the control through it -- so an imported control was left with `program_id` unset and its delete issued `DELETE .../program/0/control/{id}`. Its import id is now `"<program_id>/<id>"`, and `codegen/import_manifest.json` records the same format so `kion-import` emits ids the resource can actually parse. An entity archetype that declares `delete_extra_field` gets this automatically, since that knob already marks a record addressed through a parent the read does not return.
- `kion_category` can now be created by its acceptance test. `POST /v3/category` requires `payer_id`, which the schema and the spec it came from both mark optional, so the generated config omitted it and every create failed with `Key: 'CategoryCreate.PayerID'`. A `payer_id` is install-specific, so it cannot be a literal in `codegen/test_values.yaml`: a new `env_args` key names environment variables whose values the generated config interpolates, as `%[2]s`, `%[3]s` and so on. Previously the only way to get an install-specific id into generated HCL was to hand-write the whole test, which `internal/kgen/tests/resource_meta.go` records as the reason two resources are hand-written today.
- `labels` on `kion_ou` and `kion_cloud_rule` are now actually applied. Kion carries labels on a per-resource sub-resource (`GET`/`PUT /v3/<type>/{id}/labels`) and never in the resource's own request body, so the attribute was accepted and silently discarded on every apply. Create and Update now write them, and Read loads them, so a label changed outside Terraform is detected. Labels must already exist as `kion_label` resources: Kion answers `422 app label does not exist` for an unregistered key, and that status is a response variant rather than a Go error, so it was previously invisible -- the generated code now surfaces it instead of reporting success and dropping every label.
- Changing `aws_iam_policies`, `azure_role_definitions` or `gcp_iam_roles` on an existing `kion_ou_cloud_access_role` or `kion_project_cloud_access_role` now takes effect. All three are bound at create, but the update synced only the user and user-group lists, so a policy change was accepted, discarded server-side, and put back by the next refresh -- a plan that never converged. The associations endpoint carries all six lists; only three were declared in `codegen/memberships.yaml`. A new acceptance test swaps one attached policy for another and asserts the attribute afterwards, which fails against the previous behaviour with "Provider produced inconsistent result after apply".
- `kmigrate` now bumps the `required_providers` version constraint for `kionsoftware/kion`. It rewrote every resource to the new schema while leaving the constraint pinned to the old provider, so `terraform init` resolved the old one and every rewrite it had just made became invalid configuration. The migration guide asks for this by hand and shows a single block, but a real configuration has one per directory — the import tool emits eight. Providers other than `kionsoftware/kion` are left alone, matched on the source address rather than the local name.
- `kmigrate` parse errors now name the file that failed. The filename passed to the HCL parser was hardcoded to `config.tf`, so a run over a directory reported a path that existed nowhere in the tree and gave no way to find the offending file.
- `kmigrate` no longer leaves a configuration half-migrated. It exited on the first file it could not parse, having already rewritten every file before it — so one unparseable file (Terraform rejects them too) left a tree that was part old-provider syntax and part new, with no record of which was which. Unparseable files are now skipped untouched, listed at the end, and reported with a non-zero exit; everything else migrates. Re-running after fixing them is safe, because the rewrites are idempotent. Found by running it over a real configuration of 850 files, where one bad file stopped the migration after 125 of them.
- Association collections are now sets rather than lists. The OpenAPI spec describes them as JSON arrays, so the generator produced ordered `schema.ListAttribute`, but the SDKv2 provider modelled every one as a `TypeSet` and the Kion API returns them in no meaningful order. The difference was invisible until a configuration was migrated: state held the collection in one order and configuration in another, equal as sets and unequal as lists, so `terraform plan` reported a change removing and re-adding every element at a different position. Recognised by element type (a list of int64 is always an id collection in this API), by an `_ids` suffix, and by name for the string collections the API also returns unordered (`regions`, `supported_aws_regions`, `role_permissions`, `role_denials`, `notification_emails`, `scopes`). Measured against a live installation: a 843-resource migrated configuration went from 67 spurious in-place updates to a clean plan.
- Credential attributes are now marked sensitive, so Terraform redacts them from plan and apply output. The OpenAPI spec has no way to say "this is a credential" and nothing else supplied it, so every generated schema published its secrets in cleartext: `smtp_password` on `kion_app_config`, `key_secret` on `kion_billing_source_aws`, `private_key` on `kion_billing_source_oci`, `oauth_client_secret` on `kion_gcp_service_account`, and `tenant_client_secret` inside `kion_billing_source`'s `azure_connection`, along with the matching data sources. The previous SDKv2 provider did not mark them either, so these values were never protected. They now render as `(sensitive value)`.
- `kion_scope`, `kion_scope_criteria` and the `kion_custom_variable` data source had stopped regenerating and had drifted. The schema generator refuses a spec property carrying no type, and it drops the whole resource or data source rather than that one field, logging the skip and exiting successfully. The stale generated file stayed on disk, so nothing downstream noticed. As a result `kion_scope`'s data source still declared its association id collections as ordered lists, missing the set conversion every other type received, and `criteria` had degraded into an empty nested object. The `kion_custom_variable` data source schema had never been generated at all. `criteria` is raw JSON and is now typed as a string carrying `jsontypes.Normalized`; generation now fails loudly when a configured type is skipped.
- `kion-import` no longer reports a parent-scoped resource as failed just because some parents legitimately have none of it. A per-parent child-collection read that comes back 404 is an expected — “this parent has none” — answer, but it was being treated the same as a real server error, so a resource where every parent was empty (or a mix of empty and populated) could show `error` in the plan even though nothing was actually wrong. The client now returns a typed `StatusError` carrying the HTTP status, and the parent-scoped and multi-parent-scoped readers use it to count 404s as absences rather than failures: any records still make the result `ok`, no records with a real failure (e.g. a 502) is still `error`, and no records with only absences is `empty`, naming how many parents had none. When there are both records and 404s the absence is not mentioned at all — only failures worth investigating are called out, so the caveats block stays meaningful instead of noise. Verified with parent-scoped and multi-parent-scoped test cases covering all-404, mixed 404-and-records, a 502 among 404s, and 404-plus-502-plus-records.
- `kion-import` no longer imports aliased resources twice. `kion_aws_iam_policy` and `kion_iam_policy` are two names for one implementation over one endpoint, as are `kion_aws_cloudformation_template` and `kion_cft`, but each had its own manifest row with the same list path and both were enumerated. Against a real install that produced 1,348 identical IAM-policy ids and 296 CloudFormation-template ids, 1,644 duplicate import blocks that would have put two Terraform resources in charge of each Kion record, each reverting the other's drift. The alias is still listed, so every type the provider serves is accounted for, but only the current name is enumerated.
- `kion-import`'s manifest generator now also reads `codegen/private_endpoints.yaml`, which it previously never consulted, so three resources that were readable all along go from `unreadable` to correctly enumerable: `kion_aws_resource_tag` (flat list at `/v3/aws-resource-tag`, from the file's `list_read_only` section) and `kion_ou_cloud_access_role_exemption` / `kion_project_cloud_access_role_exemption` (parent-scoped under `/v3/ou` and `/v3/project`, child paths from the file's per-resource read entries). The lookup order is: a flat list-only path from `private_endpoints.yaml`, then the existing `generator_config.yaml` data-source path, then its resource read path, and only as a last resort a by-id read path from `private_endpoints.yaml` itself. Two more resources whose list endpoints appear in neither codegen file — verified live against a real install — are now enumerable by explicit override: `kion_dashboard` (`/v1/dashboards`) and `kion_funding_source_note` (parent-scoped under `/v3/funding-source`). Reading `/v1/dashboards` also required teaching the response unwrapper a fourth envelope shape: a named-collection object like `{"data":{"dashboards":[...],"hidden_dashboard_count":0}}` that wraps its list under a descriptive key -- alongside any number of sibling scalar/object metadata keys, which are ignored -- instead of `items` or a bare array.
- `kion-import` no longer silently drops `kion_scope_criteria` (reported as `kion_scope_criteria empty 9 record(s) skipped: no id` against a real install). `/beta/scope` returns scope objects with their criteria nested inline as a `criteria_records` array, not a flat list of criteria records reachable by a second per-parent call the way `parent_list` resources are, so every record was misread as a scope and then skipped for lacking a `criteria_id`. A new `nested_collection` read shape reads the parent list once and extracts each parent's nested array directly, building each id as `<scope id>/<criteria id>` and falling back to the element's own `id` field when the declared child-id field isn't present in the payload, which live criteria records never carry. The manifest generator now records the shape's collection key, parent-id field, and child-id field on the manifest row instead of leaving the enumerator to re-derive them.
- Resources that could not be imported now can. Importing a real installation surfaced 385 failures across 9,807 resources, in four classes. Five attributes were Required but never returned by the read (`kion_funding_source.permission_scheme_id`, `kion_project.permission_scheme_id`, `kion_gcp_account.google_cloud_project_id`, `kion_azure_account.subscription_uuid`, `kion_gcp_iam_role.role_permissions`); a Required attribute the read cannot supply makes import impossible, because `terraform plan -generate-config-out` writes configuration from the read and validation then rejects what the provider itself produced. The three enforcement resources set only `id` on import while their read looks the record up under its parent, so every import 404'd; they now take `<parent>_id/id`. `billing_start_date` was read from the private endpoint as the number 202406 and formatted as "202406", which the schema's own YYYY-MM validator rejected, so no billing source could round-trip. And a collection the API returns with a repeated id failed set validation outright.
- `kion-import` now emits parent-scoped import ids for resources whose read finds the record under a parent. `kion_project_enforcement`, `kion_ou_enforcement`, `kion_funding_source_enforcement` and `kion_scope_criteria` were given a bare id, which leaves the parent unset so the read looks under parent 0 and fails; against a real install that was 187 failed imports. It also no longer emits a record whose id is 0, which produced an import block for an object that does not exist.
- `kion-import` no longer gives a compound import id to resources that cannot parse one. Reaching a record through a parent is not the same as having an ImportState that splits on "/": `kion_compliance_family` and `kion_compliance_level` declare no archetype, so they use the plain id-only ImportState and merely fall back to a parent-scoped read when their flat list is not served. Handing them `<parent>/<id>` made their read parse the whole string as an integer. Only the archetypes whose generated ImportState actually splits, parent_list and compound_key_parent_read, take a compound id.
- Importing an existing installation now plans no changes. Verified against a production-scale install: 10,173 resources import and `terraform plan` reports 0 to add, 0 to change, 0 to destroy, where the same run previously reported 2,561 in-place updates and 385 errors. Four causes. Computed attributes carried no `UseStateForUnknown`, so every one planned as "(known after apply)" even though nothing had moved. JSON-valued attributes were compared byte-for-byte against the compact, key-sorted form `jsonencode` produces, so every policy, template and role definition reported a whitespace change; reads now emit the same canonical form, and a value that is not JSON, such as a YAML CloudFormation template, is left alone. Optional string attributes that the account endpoints return as "" were stored as empty rather than null, which the generated configuration writes as null. And 43 attributes were Required that the read never returns, which makes a resource unimportable by construction.
- Giving every Computed attribute `UseStateForUnknown` broke the plan for three resources outright: `kion_account_cache`, `kion_azure_policy` and `kion_billing_source` failed with `Attribute Missing`. tfplugingen generates `ValueFromObject` without the null/unknown guards its own inverse `ToObjectValue` and its sibling `ValueFromTerraform` both have, and the framework round-trips an object attribute's plan value through it after every object plan modifier. An Optional+Computed object absent from the config is unknown at plan time and has no attributes, so the generated code rejected it. The schema generator now supplies the guards, and `make ci-modules` runs `terraform test` over every module rather than only the staleness diff, which is why nothing local caught this.
- `kion_project_note` and `kion_dashboard` could not be imported at all: their private reads render several columns through Go's SQL null wrappers, so the whole response failed to decode against a wire struct declaring plain scalars. Verified against a live install, `/v2/project-note` returns `last_update_user_id` as `{"Int":0,"Valid":false}` and `updated_at` as `{"Time":"…","Valid":false}`, and `/v1/dashboard` returns `description` and `updated_at` the same way and `config` as the JSON object itself rather than a string of JSON. A new `read_kinds` map in `private_endpoints.yaml` retypes individual attributes to tolerant wire types that accept either rendering and encode back to the bare scalar, so making the read tolerant does not change what a write sends. A never-updated record now reads as null instead of the year-1 zero instant, and `config` is canonicalised to the byte form `jsonencode` emits.
- `kion_ou_cloud_access_role_exemption`, `kion_project_cloud_access_role_exemption` and `kion_aws_resource_tag` imported as empty resources: the `no_read` archetype's `Read` is a no-op and `ImportState` sets only the id, so `terraform plan -generate-config-out` wrote a body with no attributes at all, and the plan reported no changes because every attribute is Optional+Computed and null. `no_read` means the public spec has no by-id GET, not that the record is unreadable, so all three now declare a `parent_read` and get a real read. Two properties of the exemption collections are handled with it, both measured against a live install: they are inherited rather than owned, so one record is returned under every OU in its subtree (329 rows for 22 records) and the owner is taken from the record's own key rather than the path; and they mix cloud rule exemptions in with cloud access role exemptions, so a declared discriminator keeps records of the neighboring kind out (only 6 of 22 were this resource, and 0 of 12 on the project side). Their import ids are now `"<parent_id>/<id>"`.
- `kion-import rewrite-refs` no longer emits a configuration Terraform refuses to plan. A reference is an edge in Terraform's dependency graph, and some Kion relations are settable from both sides — a user names its groups and a group names its users — so rewriting both directions closed a loop and `terraform validate` failed with `Error: Cycle: kion_user.kion_user_119, kion_user_group.old_kion_developers`. Against a 321-resource install this was not one unlucky record: it produced one 84-node cycle plus six smaller ones. The rewriter previously guarded only self-reference, by string-matching the block's own label, which caught `parent_ou_id` pointing at its own OU and nothing longer. It now builds the dependency graph as it goes and refuses any edge that would close a cycle at any length, leaving that one literal in place. Refusal is per id rather than per attribute, so a list keeps the references that were never part of a loop. The literals kept this way are reported like the existing boundary output but counted separately, because the cause and the remedy differ: the referent is managed here, nothing is missing, and the other side already carries the reference. Edges are offered to the graph in a fixed order — block order in the file, then attribute name, then position in the expression — so the same edge loses on every run and the output does not change between invocations. References the configuration already holds seed the graph, so a second pass cannot add the edge the first pass refused.
- `kion_custom_variable_override` could not be imported at all. `kion-import` refused it because its read path carries two placeholders (`/v3/account/{account_id}/custom-variable/{custom_variable_id}`) and neither id can be discovered from the other; separately, its `ImportState` parsed the id as a bare integer and set only `id`, leaving `entity_type`, `entity_id` and `custom_variable_id` -- the three fields `Read` dispatches on -- unset, so an import would have failed even had one been generated. The compound path turns out to have a plural sibling that needs only one id: `/v3/{account,ou,project}/{id}/custom-variable` lists every variable at that entity with a non-null `override` on exactly the ones set there. The resource is now enumerated across all three parent kinds -- it is polymorphic, and on a demo install an account-only read would have found 5 of the 21 overrides that exist -- at one request per entity (151) rather than one per entity/variable pair (1,812). `ImportState` now splits the three-part `"<entity_type>/<entity_id>/<custom_variable_id>"` id that `Create` has always built, and validates both ids. Verified end to end against a live install: 21 imported, 0 changed, and a second plan reports no changes.
- `kion-import` reported every `require_valid_field` drop as "record(s) of another kind sharing the collection". That is true of the cloud-access-role exemption collections, which really do mix kinds, but wrong for `kion_custom_variable_override`, whose collection returns the same kind merely inherited rather than overridden at that entity. The message now names the discriminator instead of asserting a cause: `1099 record(s) in the collection with no "override" set`. Drops are still always counted and reported.
- The provider's HTTP client now allows 60s per request rather than 30s, matching `kion-import`. The timeout covers the whole request including the response body, and a filtered data source fetches its entire collection before filtering in-process, so a large one could not be read at all: on a production-scale install `GET /v3/azure-policy` returns 51 MB in roughly 35s, and every `data "kion_azure_policy"` with a `filter` block failed with a bare `context deadline exceeded`. `kion-import` enumerated the same endpoint without trouble, because it already used 60s.
- `kion-import` no longer loses records to paginated collections that pad their pages. Some endpoints pad every page out to `total` rather than to the page size: `GET /v4/compliance/program/5/family?count=100` returns 110 items on page 1 (100 real plus 10 all-zero fillers) and 110 again on page 2 (10 real, 100 fillers). Counting the padding made page 1 look complete, so paging stopped and the real records reachable only on later pages were never fetched. They surfaced as `record(s) skipped: no id`, which reads like junk being discarded rather than real records going missing. Zero-valued records are now dropped before the page count is compared against `total`. Against a real installation this recovers 3,602 records: `kion_compliance_control` goes from 2,414 to 6,006 and `kion_compliance_family` from 587 to the full 597.
- `make sweep` no longer reports success while cleaning nothing. Three defects stacked up: the target ran `go test -sweep=all ./internal/provider/...`, which put a package path after a flag `go test` does not recognize (so the path went to the test binary and the root package ran instead) and pointed at a package with no `resource.TestMain`, meaning no sweeper had ever run; 29 of 61 generated `sweep.go` files registered a sweeper whose body returned `nil`; and the hand-maintained import list that links sweepers into the run had drifted. The target now runs `./internal/sweep/... -sweep=all`, 36 resources have a real sweeper (five of them newly generated for parent-scoped collections such as `kion_ou_note`, declared as `sweep_parent` in `codegen/crud_archetypes.yaml`), and the 14 that cannot be enumerated or deleted through the API register nothing and say why in their `sweep.go` rather than reporting a clean sweep. `docs/TESTING.md` lists them. `TestSweeperRegistration` fails on a re-introduced no-op or a drifted import list.
- Lifecycle and other server-owned fields are now filterable on list data sources. 33 fields across 16 resources reached neither the resource schema nor the `filter` surface, so a configuration could not match on them or read them out of `list`: `created_at` and `deleted_at` (`kion_account`, `kion_azure_account`, `kion_gcp_account`, `kion_custom_account`, `kion_user`, `kion_account_linkage`, `kion_app_api_key`), `last_login` and `display_name` on `kion_user`, `car_restricted` on `kion_azure_role` and `kion_gcp_iam_role`, `identifier` and `criteria_error` on `kion_scope`, and others. Two of these never rendered at all: `OptNilString` had no projection into a `list` object, which silently dropped every nullable timestamp the API returns. Credential fields stay unexposed, now by policy rather than by accident -- exposure is default-deny for any credential-shaped string, and `codegen/unexposed_fields.yaml` names the rest.
- The remaining server-owned account fields are now filterable. `kion_azure_account`, `kion_gcp_account` and `kion_custom_account` gain `account_email`, `account_number` (`account_type_id` on custom), `car_external_id`, `include_linked_account_spend`, `linked_account_number`, `linked_role`, `service_external_id`, `use_org_account_info` and `skip_access_checking`, all `Computed`. These were withheld pending a per-cloud judgment that turned out not to apply: all five account resources read the same `/v3/account` endpoint, which returns one flat record shape with no separate per-cloud variant. Filtering on a field that is empty for a given account correctly matches nothing, whereas not exposing it meant no filter was possible at all. With this, every scalar the API returns is reachable from a configuration except `key_id` and `key_secret` on `kion_account_cache`, which are credentials and stay unexposed.
- `kion-import` no longer imports every billing source type as `kion_billing_source`. `GET /v4/billing-source` is the polymorphic collection of every payer type -- aws, azure, gcp, oci, custom, anthropic, openai -- but `kion_billing_source` is specifically the CUSTOM billing source: it creates through `/v3/billing-source/custom` and reads `/v1/payer/{id}`, unwrapping `custom_billing_source`. Enumerating the whole collection handed it records it cannot represent; the by-id read then found no `custom_billing_source` key, flattened every attribute to its zero value, and the resource refreshed as `id = "0"`. Every `payer_id` reference pointing at one of those resolved to 0, so an imported configuration planned to unlink accounts from their billing sources -- 6,611 of 12,909 resources on one installation. The collection is now filtered by the same discriminator the cloud-access-role exemptions use, and the dropped records are counted and reported rather than silently omitted.
- The four `kion_billing_source*` resources now report the payer id rather than the id of the record nested inside it. `/v1/payer/{id}` answers on the payer id and nests the type-specific record, which carries its own id from a different space -- payer 22 holds `custom_billing_source.id` 5, payer 2 holds `aws_payer.id` 1. Every endpoint addresses the payer: the import block, `DELETE /v3/billing-source/{id}`, `PATCH /v3/billing-source/custom/{id}` (which resolves the payer, then verifies it has a custom record), and `account.payer_id`. Reporting the inner id made every reference to a billing source resolve to the wrong number, and would have pointed `terraform destroy` at whatever billing source happened to hold that id. On one installation this was the last 35 spurious `payer_id` diffs after the collection filter; the imported configuration is now inert.
- `kion_scope_criteria` no longer plans a change on every record. Its `criteria` attribute is a JSON string, and the compound CRUD path stored the API's response verbatim while `terraform plan -generate-config-out` writes the same value as `jsonencode({…})` -- compact, with sorted keys. The two differ only in whitespace, but semantic equality is not consulted when a generated configuration is planned, so every record reported an update that changed nothing. The value is now canonicalised through `flex.NormalizedToFramework`, which the entity path already used for exactly this reason. Against a real installation the plan goes from 9 changes to a completely clean one.
- `kion_idms_open_id`, `kion_idms_open_id_access_rule` and `kion_idms_open_id_group_association` can be imported again. All three are entity archetypes reached through a parent, so their generated `ImportState` takes a bare id, but the manifest handed them a compound `"<parent>/<key>"`. `Read` parsed the whole string as an integer, `terraform plan -generate-config-out` wrote no configuration for them, and `terraform apply` then failed the ENTIRE import with `Configuration for import target does not exist` -- three unimportable records stopped 10,071 others from reaching state. Whether a resource may carry a compound id depends only on whether its `ImportState` splits on `/`, and that fact now lives in one predicate used by both the classifier and the parent overrides rather than in two places that disagreed.
- Every `Create` and `Update` now resolves still-unknown attributes to null before writing state, fixing `Provider returned invalid result object after apply` on 8 resources (`kion_cloud_rule`, `kion_compliance_standard`, `kion_compliance_check`, `kion_iam_policy` and others). An `Optional`+`Computed` attribute is unknown at plan time when the configuration omits it, and the read-back after create assigns only what the API returns -- so a write-only input like `cloud_rule_id`, or any field the API never echoes back (`last_updated`, the `kion_cloud_rule` attachment lists), reached state still unknown and Terraform rejected the apply outright. Null is the honest answer, since the API does not hold the value, and `UseStateForUnknown` then keeps it stable across later plans. Separately, the expansion of a nested object list no longer calls `ElementsAs` on a null or unknown collection, which failed with `Received unknown value, however the target type cannot handle unknown values` on `kion_cft` and `kion_permission_scheme`; it is now guarded the same way the `flex` slice helpers already were. Neither defect was reachable by `build`, `vet`, `lint` or the unit tests -- nothing below an actual apply inspects a post-apply value -- and both were found by the provider's first acceptance-test run.
- Acceptance tests now skip, rather than fail, when the target install cannot run them: `kion_azure_policy` and `kion_azure_arm_template` require a live Azure connection and gate on `KION_ACC_AZURE_PAYER_ID`, matching how a missing install-specific id already skips. A new `acctest.InstallCapabilities` reads the install's financial mode from `GET /v3/app-config` for the same purpose; a failure to read it is fatal rather than a skip, so an unreachable install cannot quietly turn tests green. The `kion_ou_note` tests were also repaired: they hardcoded `ou_id = 1` and changed `ou_id` and `create_user_id` on update instead of the note's own text, so they could only ever have passed on an install that happened to have an OU with that id, and their existence and destroy checks were TODO stubs that asserted nothing. They now create their own OU and verify through the API.
- The four permission-mapping resources (`kion_ou_permission_mapping`, `kion_project_permission_mapping`, `kion_funding_source_permission_mapping`, `kion_global_permission_mapping`) now converge instead of proposing the same write on every plan. Kion stores an empty member list and reports it back as JSON `null`, so `null` and `[]` are one state on the wire; the read mapped `null` to a null set, which never equalled the `user_ids = []` or `user_groups_ids = []` a configuration asked for, and each refresh re-proposed the change. The read now collapses both to an empty set, and `Create`/`Update` read the mapping back from the API before writing state, so a collection the configuration omits also lands on what Kion actually holds rather than on a guess left over from an unknown plan value. The read-back has a second effect: Kion answers 200 to a mapping for an app role the parent's permission scheme does not grant, then silently stores nothing, which used to leave a resource that planned as absent for ever; that now fails at apply with a diagnostic naming the cause. Two data-source acceptance tests that asserted an `id` their data source never sets were corrected: `kion_gcp_regions` has no `id` attribute at all, being the whole collection with nothing to select by, and `kion_account_linkage` read without an `id` or `filter` is in list mode, where `id` is the input rather than a result.
- `make testacc` ran zero acceptance tests. It targeted `./internal/provider/...`, which contains no `TestAcc` functions -- all 90 live in `./internal/service/...` -- so it exited 0 having run nothing, which is how the provider's 408 `Create`/`Update`/`Delete`/`ImportState` entry points reached 0% coverage without anybody seeing a failure. The target now runs the 71 service packages. Alongside it: `KION_APIPATH` lets the acceptance tests reach an installation that serves its API at the root rather than under `/api` (they write no provider block, so an environment variable was the only way to express it), `.github/workflows/acctest.yml` runs the suite on manual dispatch against a chosen GitHub Environment, and `scripts/discover-acc-env.py` reads the ten install-specific `KION_ACC_*` ids off a live install.
- `kion_custom_variable` accepted `default_value_string`, `default_value_list` and `default_value_map` and sent none of them, so every apply failed with `unknown custom variable type encountered during value creation: <nil>`; `kion_permission_scheme` sent an empty update body, so every update failed with `Incorrect data structure`. Both are the same generator failure: a request field it could not map was skipped without a word, leaving a struct literal that compiles and lies. `make bind-check` now fails when a settable attribute is read by neither Create nor Update, reading the generated output so bespoke and hand-written bodies are held to the same rule; the attributes that already went unsent are recorded in `codegen/unbound_attributes.yaml` as a ratchet. Three shapes the generator previously dropped now map: a nested request object whose sub-fields are the model's top-level attributes, a request array that a declared `read_shape` explode describes in reverse, and a polymorphic `jx.Raw` field the schema splits into typed `_string`/`_list`/`_map` attributes -- which now also comes back apart on read, so the value survives refresh and import.
- 17 service packages gained their first acceptance test, and running them against a live 3.16 install found four defects that compiled, linted and passed their schema unit tests: `kion_funding_source` drops `owner_user_ids` and `permission_scheme_id` on read so it cannot be imported (#68); `kion_budget` loses `amount`, which create expands into per-month `data` rows that `flattenBudget` ignores (#69); `kion_ou_enforcement` and `kion_funding_source_enforcement` drop `cloud_rule_id` (and `service_id`) because the record carries them as nested objects that `flatten` never unwraps, where `kion_project_enforcement` of the same archetype does unwrap them (#70); and every `kion_funding_source_note` is created with id 0, because `POST /v2/funding-source-note` returns no `record_id` and `Create` treats the missing id as success (#71). Those tests are left red on purpose, each naming its issue in a header comment. The generator behind them (`internal/kgen/tests`) was itself the reason these packages had none: it emitted `generated/v3_15` imports while the provider is on `v3_16`, so every file it wrote with an SDK get method failed to compile; it filled a Required attribute that was also a dependency's target field from the generic `1` fallback instead of referencing the dependency it had just created (`kion_compliance_family` posted a program id of 1 and got a 404); it asserted every computed data-source attribute was set, including optional ones the config never sets; and its registry carried the old SDKv2 provider's `owner_users { id = 1 }` block syntax, which no generated schema declares. It now also supports `RequiredEnv` (a resource whose create needs a record this install cannot invent SKIPs naming the variable, rather than failing), `ImportIDParentField` (the `parent_list` archetype's `"<parent>/<id>"` import id), `NoUpdate`, and `KnownIssues`.
- `kion_project` could not be created on any budget-mode Kion install. It created through `POST /v3/project`, which the API rejects with `spend plans not available in budget mode` before it parses the body, so no payload could satisfy it and the five packages that need a project to exist -- `kion_project`, `kion_project_cloud_access_role`, `kion_project_enforcement`, `kion_project_permission_mapping`, `kion_custom_variable_override` -- were unusable with it. Create now dispatches on the install's financial mode: `POST /v3/project/with-budget` in budget mode, `POST /v3/project/with-spend-plan` otherwise. The mode is read from `GET /v3/app-config` when the provider configures, and a `budget` or `project_funding` value in the configuration overrides it, which is also the answer for a credential without the global settings permission needed to read app-config; with neither signal the provider says so instead of guessing. `budget` is now sent -- it was in the schema and in the model but no request carried it, consistent with targeting an endpoint that does not accept it. `POST /v3/project` is deprecated per the portal's own swagger comment and is no longer used.
- Updating `kion_ou_cloud_access_role` failed with `decode response: unexpected status code: 200` on a write that had already succeeded, so Terraform marked the apply failed while the server held the change -- state and reality diverging, which is worse than a clean error. The spec declares `201` for the PATCH and the server answers `200`, so the ogen-generated decoder has no arm for the success it was given. Four operations are in that position: `PatchAppConfig`, `PatchAppRole`, `PatchOUCloudAccessRole` and `PatchProjectCloudAccessRole`, each annotated in the API spec as `201: OKResponse` -- self-contradictory, since `OKResponse` is the 200 model. (`PatchBillingRule`, also flagged, is correct: it declares `204` and `NoContent` is what it returns.) The real fix is upstream -- correct the four annotations, regenerate the spec and the SDK, bump the pin -- and until it lands the generated Update bodies treat a 2xx the spec does not declare as the success it is, taking their state from the read-back that follows, which is authoritative either way. Nothing else is tolerated, and once the spec declares 200 the decoder returns a response instead of an error and `errs.IsUndeclaredSuccess` never fires again.
- `kion_funding_source_note` is manageable at all. Every apply created a note with id `0`: `POST /v2/funding-source-note` answers `{"status":201,"data":""}` with no `record_id` anywhere in the body, the hand-rolled envelope decoded that silently as zero, and `Create` wrote `"0"` to state -- so refresh read id 0 and found nothing, destroy issued `DELETE /v2/funding-source-note/0`, and the real note stayed on the install forever. The note's id is now recovered by diffing its parent collection across the create, declared as `create_id` in `codegen/private_endpoints.yaml`. The fix is in the raw-HTTP create template rather than at the call site: any raw create that cannot produce a usable id now fails loudly (`errs.RawCreatedID`) instead of writing an unaddressable resource to state, which also closes the same unguarded path in `kion_dashboard`. `errs.CreatedID` rejects an explicit `record_id: 0` on the SDK path for the same reason.
- Five resources accepted an attribute and then lost it on read, so import dropped it and a configured value diffed for ever. `kion_ou_enforcement` and `kion_funding_source_enforcement` dropped `cloud_rule_id` (and OU's `service_id`) because their record carries the cloud rule as a whole object while the model stores its id -- `kion_project_enforcement` unwrapped it only because the SDK happens to name the identical field `cloud_rule_id`. The flatten resolver now maps a payload object `foo` onto the model's scalar `foo_id`, the mirror of the existing `foos` -> `foo_ids` rule, which fixes all three uniformly and also recovers `kion_project_enforcement`'s `service_id` and the `ou_id`/`project_id` that `kion_ou_cloud_access_role` and `kion_project_cloud_access_role` were dropping. `kion_budget` lost `amount`: create expands it into one data row per month and the read returns only those rows, so the archetype now declares `sum_from` and the flatten adds them back up. `kion_funding_source` lost `owner_user_ids`, `owner_user_group_ids` and `permission_scheme_id`, none of which `GET /v3/funding-source/{id}` returns at all; its read now also reads the owner-role row of the funding source's permission mapping and the app policy behind the permissions tab, so an imported funding source finally produces a configuration that applies -- `POST /v3/funding-source` rejects a body with no permission scheme. Changing `owner_user_ids` or `permission_scheme_id` on an existing funding source now fails at apply rather than silently recording a value Kion does not have: the public PATCH carries neither (already recorded in `codegen/unbound_attributes.yaml`).
- Five more service packages gained their first acceptance test — `kion_aws_resource_tag`, `kion_compliance_control`, `kion_ou_cloud_access_role_exemption`, `kion_user` and `kion_webhook` — taking acceptance coverage to 54 of 72 packages. `kion_aws_resource_tag` and `kion_user` pass outright; the other three found four defects. `kion_compliance_control` never sends or reads back `program_id`, its only Required attribute, which `Delete` is the sole consumer of, so an imported control cannot be destroyed (#77); the same resource turns an explicitly empty Set into `null` on read, because the API stores `[]` as `null` and the flatten uses `flex.Uint64SliceToFrameworkSet` rather than the `OrEmpty` variant #67 added, so the apply is rejected as an inconsistent result (#78). `kion_webhook` and `kion_user` have a `Delete` that deletes nothing, warning "no delete endpoint" — but `DELETE /v1/webhook/{id}` and `DELETE /v1/user/{id}` both exist and work, so every record either type creates survives destroy, and neither has a sweeper to reclaim it (#79). A `kion_ou_cloud_access_role_exemption` that `POST` returns a `record_id` for, and that `DELETE` removes cleanly, is absent from every collection the API exposes, so `Read` drops it from state on every refresh (#80). Those three tests are left red on purpose, each naming its issue in a header comment.

The test generator gained `RawCollectionPath` (with `RawCollectionParentField` and `RawCollectionDiscriminator`), so a resource with no single-record GET gets Exists and Destroy checks that read the same private collection its own `Read` uses. They were previously `// TODO` stubs that returned `nil`, reporting green whether or not the API had ever seen the record. `metaHasFormatVerbs` now also scans `ExtraHCLBlocks`, which is where a resource whose every attribute is Optional carries its only values — a verb left unscanned reached the config as the literal text `%[1]s`.

`conns.SharedClient` built a client carrying only `Client` and `APIURL`. `rawRequest` reads the credential off the struct rather than through the SDK's security source, so every raw call an acceptance check made answered 401 — which is the only way to check a resource whose read is a private collection.

### NOTES

- `kgen import-manifest` generates `codegen/import_manifest.json`, describing how to enumerate each resource and how to build its Terraform import id. It derives everything from `generator_config.yaml` and `crud_archetypes.yaml`, so the import id format always matches the `ImportState` the CRUD templates generate, and the list endpoints come from the same data-source paths the provider itself uses. `make generate` runs it, and `TestManifestIsCurrent` fails if the committed manifest is stale. The manifest is embedded into `kion-import`, so a shipped binary cannot be paired with a manifest from a different commit.
- `scripts/check-no-internal-refs.sh` now rejects any internal hostname under the company domain, not just the one host it happened to list, and no longer carries the internal literals it screens for. A denylist of internal names is itself an internal reference: committing the strings publishes the index of what we are trying not to publish, and the script had to exempt itself from its own grep. The committed patterns now name nothing in particular; the specific literals come from `INTERNAL_REF_PATTERNS` (a CI secret) or an untracked local file, and the script reports when that private set is absent so a weakened check is visible rather than assumed. Use `kion.example.com` in documentation.
- `reason` on `kion_ou_cloud_access_role_exemption` and `kion_project_cloud_access_role_exemption` is accepted by the API and never stored: there is no `reason` column on the exemption table, no such field on the backend entity, and the `/v3` controller that handles the create never reads it off the request, though the request model declares it with a swagger example. Setting it in a configuration is therefore silently a no-op, and an imported exemption has `reason = null` because no exemption has ever had one. Documented in `docs/IMPORTING.md` under "Attributes the API accepts and discards". This corrects an earlier note that described it as a value that was set but could not be read back — it sent operators to re-enter something that does not persist.
- Two more acceptance tests stop failing for reasons that are not the provider's. `kion_account_linkage` skips unless `KION_ACC_PAYER_ID` names a billing source: a linkage row carries a foreign key to `payer`, so on an install with none the create fails on the database constraint. `kion_user_group`'s fixtures now set `owner_user_ids`, which `POST /v3/user-group` requires -- the schema marks `owner_user_ids` and `owner_user_group_ids` Optional, which is true of each alone but not of the pair. The shared acceptance-test client also carries its credential onto the raw-HTTP helpers, which had been sending unauthenticated requests that 401 and read like a deleted record.

## [1.0.2] - 2026-08-24


### BUG FIXES

- Regenerated the documentation and examples for the 38 data sources whose lookup mode changed. Restoring `filter` support made `id` optional on those data sources, but the registry docs and examples still described `id` as required and showed no `filter` block, so they documented behaviour the released provider no longer had.
- `kmigrate` no longer needs `state_upgrades.yaml` alongside it. The `--upgrades` flag defaulted to `codegen/state_upgrades.yaml`, a repository-relative path that does not exist in a practitioner's Terraform directory — so the first command in the migration guide failed for everyone outside a clone of this repo, with `open codegen/state_upgrades.yaml: no such file or directory`. The ruleset is now built into the binary, which also makes it impossible to pair `kmigrate` with a ruleset from a different commit. `--upgrades` remains as an override.

### NOTES

- Acceptance-test configurations are now validated against the provider schema in CI. Acceptance tests only run against a live Kion instance, so nothing had ever checked that the HCL they feed Terraform matches the schema — and it did not: twelve configs used an `owner_users` block on resources whose schema declares an `owner_user_ids` list attribute, two set a read-only `id`, and the great majority were still the `kgen` scaffold's `# TIP: Fill in required attributes` placeholder with an empty resource body. All 87 configs are now filled in and validated by a new `acctest-config` job (locally, `make ci-acctest-config`), which builds the provider and runs `terraform validate` over each one — no Kion instance and no API call required.
- The migration guide is now linked from the provider's registry landing page and from the repository README, and explains how to obtain `kmigrate` — it is attached to every release as a single dependency-free binary. Previously neither entry point mentioned migration at all, so a practitioner upgrading from the SDKv2 provider had no path to the procedure.
- `docs/` and `examples/` are now drift-gated in CI. They were the only generated surfaces without one — the README said so — which is how 1.0.1 shipped registry documentation describing `id` as required on 38 data sources whose schemas had already made it optional: `internal/service/` and `modules/` were regenerated because CI enforced them, and these were not because nothing did. `make docs-check` is the local equivalent.

## [1.0.1] - 2026-08-24


### ENHANCEMENTS

- Releases now attach the `kmigrate` binary and `codegen/state_upgrades.yaml`, so migrating a configuration no longer requires cloning this repository. Neither could be fetched any other way: the module path carries no domain, so `go install` has no remote path to resolve. Both new artifacts are listed in `_SHA256SUMS`, so the existing signature over that file covers them.

### BUG FIXES

- Restored the `kion_funding_source_permission_mapping`, `kion_global_permission_mapping`, and `kion_project_note` data sources, which the old provider shipped but which the rebuild dropped. Their companion-path archetypes (`association`/`blended`) derive no data source, so each needs a registered companion template; only their `ou_permission_mapping`/`project_permission_mapping` siblings had one. The permission-mapping data sources also gained a `filter` block, matching those siblings.
- The `kion_azure_policy` data source regained its `filter` block and no longer requires `id`. Its read payload (`AzurePolicyAugmented`) carries no top-level `id` — the record, id included, is nested under `azure_policy` — so the generator read the payload as id-less and degraded the data source to id-only. Record-wrapper detection now also recognises a record nested under a model object attribute, so `id`, `description`, `name`, `parameters`, and `policy` are sourced from the wrapped record in both by-id and by-filter mode. `kion_azure_policy` also gained a real acceptance-test sweeper for the same reason.
- Upgrading a `kion_aws_cloudformation_template` resource from the SDKv2 provider no longer fails when it has `tags` set. The old schema took `tags` as `map(string)`; the new `kion_cft` schema declares it as a list of `{ tag_key, tag_value }` objects, but the generated v0 state upgrader passed the old map through verbatim, so `terraform plan` aborted with `AttributeName("tags"): invalid JSON, expected "[", got "{"` and the resource could not be migrated at all. The state upgrader now explodes the map into the object list, in sorted key order (which is the order Terraform recorded the map in), and leaves an already-migrated list untouched so re-running is a no-op. The map-to-object-list restructure is now expressed once, as a `kv_list` rule in `codegen/state_upgrades.yaml`, and read by both the state upgrader and `kmigrate`'s config rewriter — previously only the config half knew about it, which is why the state half was missed. New snapshot-driven guards assert that every old-to-new attribute whose JSON shape changed is covered by a rule, and that every block passed through into a nested-object attribute still matches it field for field.
- Data sources whose SDK list envelope puts the items directly in `Data` (for example `UserListResponse{Data []User}`) are no longer silently downgraded to id-only. List resolution looked the trimmed field type up in the struct index, which for that shape is the literal string `[]User` — a slice type that can never be a struct — so resolution aborted, the list model stayed empty, and the id-only template was rendered instead of the list one. The only signal was a line on stderr while `kgen` still exited 0, which is why a large number of data sources shipped without a `filter` block and requiring an `id` their old-provider configurations never had. That shape is now handled by the same helper the items-field path already used.
- `kgen crud` no longer silently skips `no_read`-archetype resources. The service-package template branches on a field that the `no_read` call site did not pass, and Go's `text/template` errors on a field absent from the data struct rather than treating it as empty — so template execution failed, the resource was skipped with a line on stderr, and `kgen` still exited 0. Three resources (`aws_resource_tag`, `ou_cloud_access_role_exemption`, `project_cloud_access_role_exemption`) were therefore unregenerable, and their committed output had drifted from the templates with nothing reporting it. A `no_read` resource can now also register a companion data source, lifting a limitation its own code comment described.
- `kmigrate` now rewrites ownership blocks written as `dynamic` blocks. Every ownership block in the coverage surface is written as `dynamic`, which is how a configuration generates a repeatable block from a variable, and the rewriter recognised only literal blocks — so all 22 `owner_users` / `owner_user_groups` / `owner_groups` findings survived migration untouched. An attribute has no dynamic form, so the construct collapses into a `for` expression over the same `for_each`, with the block's iterator resolved to the loop variable (the two-variable form only when the body reads `.key`); a resource mixing literal and dynamic blocks of the same name gets a `concat`. Every generated expression is round-tripped through the HCL parser, so a rewrite that would not parse is reported rather than written, and a malformed dynamic block comes back as a follow-up instead of a panic. The same support covers the object-bodied blocks (`project_funding`, `budget`, `move_ou_settings`).
- `kmigrate` now drops `kion_gcp_account.create_mode` and `kion_custom_account.skip_access_checking`, two old settable attributes with no home in the new schema. A migrated configuration that still carried them failed `terraform validate` with "Unsupported argument". They had been held out by a policy stating that the four `*_account` resources migrate via import and are therefore not subject to config rewriting — which was not true, since `kmigrate` already rewrites `name` to `account_name` on those same resources. The exclusion is now narrowed to what still holds.
- `kmigrate` now drops attributes the new schema made read-only. `kion_project_note.create_user_id` was a required input on the old resource; the new one derives it from the authenticated caller and reports it back as computed, so configuration that still set it failed with "Invalid Configuration for Read-Only Attribute". This could not be expressed as an ordinary drop, whose guard asserts the attribute is absent from the new schema rather than present-but-computed, so it uses a sibling table with the opposite guard, checked both ways round by a test — any future settable-to-computed change now fails the build until it is covered.

## [1.0.0] - 2026-08-20


### BREAKING CHANGES

- Provider SDK imports moved from `kion-sdk-go/generated` to per-version sub-packages. The provider builds against `generated/v3_16`; the root `kion` package no longer exports `NewClient`.
- `kion_user` now requires `email`, `first_name`, `idms_id`, `last_name`, and `username`. The previous provider's `kion_user` had no configurable attributes at all — only a computed `id`, with read and delete and no create — so it could only adopt an existing user for deletion. The new resource is full CRUD. Existing state migrates without action (the read repopulates the new attributes), but an existing `resource "kion_user"` block must gain the five arguments or `terraform plan` fails with "Missing required argument". `kmigrate` reports each affected block; it cannot fill the values in.

### FEATURES

- `kgen` code generator with Go templates for resources, data sources, tests, and examples; `kmigrate` rewrites configurations written for the previous provider.
- Runtime Kion-version compatibility: the provider detects the connected Kion version at configure time and gates resources whose defining API operation does not exist on that release, instead of shipping a separate provider release line per Kion version.
- AWS-style service-package architecture across 72 service packages, with generated CRUD, schemas, data sources, acceptance-test scaffolding, examples, registry docs, and a Terraform module per resource.

### ENHANCEMENTS

- `kmigrate` now reports blocks missing an attribute the new schema requires and the old provider had no equivalent for — today only `kion_user`. It cannot supply the values, so these are listed separately from the edits it made, under "block(s) need attention before `terraform plan` will pass". Rename targets are excluded, since the rewrite fills those in from the old name. A guard test derives the set from the schema snapshots, so a future spec that makes an existing resource require something new fails the build until it is recorded.
- Label data source: backwards-compatible filter mode.
- `OptStringToFrameworkLegacy` flex helper for legacy SDK compatibility.
- Restored the `PostAccountCacheCreateNewAWS` endpoint with govcloud and organizational-unit support.

### BUG FIXES

- Fixed the raw HTTP endpoints (`app_role`, `billing_source`, `billing_source_aws`, `billing_source_gcp`, `billing_source_oci`, `dashboard`, `funding_source_note`, `permission_scheme`, `project_note`) appending a hardcoded `/api` on top of the already-resolved API root — requesting `/api/api/…` under a default configuration and ignoring `apipath`. `apipath` now applies to the SDK and the raw endpoints alike.
- Setting an attribute the connected Kion is too old to accept now fails at plan, naming the attribute and both versions. Previously nothing checked it: gating worked at the level of whole operations, so a field added in 3.16 to a resource that exists in 3.13 had no gate at all. Portal decodes request bodies with `json.Unmarshal` and no `DisallowUnknownFields`, so the field was not rejected but dropped — the value never landed, and the practitioner saw "Provider produced inconsistent result after apply" or a diff that never converged, neither of which names the cause. `kgen versions` now walks the create body across every tracked SDK version and records the oldest Kion accepting each attribute, which finds `kion_billing_source.azure_connection`, `kion_cloud_rule.automation_policy_ids`, and `kion_project_enforcement.notification_emails` / `notification_frequency` — all 3.16-only fields on resources with no version gate of their own. Attributes no newer than the resource itself are omitted, since the resource gate already refuses those.
- Fixed SDK fixspec silently dropping 9 of 13 query-string-discriminated operations (`PostAwsAccount`, `PostAzureSubscription`, `PostGoogleCloudAccount`, and others).
- A Kion instance too old for a resource is now reported during `terraform plan` rather than part-way through `apply`. Gated resources gain a `ModifyPlan` (destroy plans stay allowed, so an unsupported instance cannot strand a resource in state), and the gate now runs in `Read`, `Update` and `Delete` as well as `Create` — as `RequireKionVersionInRange`'s own documentation already described. Gating `Read` matters most: on an instance without the endpoint the API answers 404, which the generated read treats as "resource gone" and silently removes from state. `codegen/version_support.yaml` was also stale, leaving `kion_billing_source` (3.14+), `kion_dashboard` (3.15+) and `kion_scope` (3.15+) with no gate at all; regenerating it brings them and several data sources under the same protection.
- `make version` reported a hardcoded `3.15.1`. It read `info.version` from `spec/swagger-*.json`, which the repo carried until that file was replaced by `openapi3.json`; the glob has matched nothing since, so the fallback — the Kion spec version it happened to be pinned to — became the only value it could produce. That number also named the directory `make install` wrote to, and it described Kion's API rather than the provider. The version now comes from the latest release tag (`0.0.0-dev` when untagged), matching how a release is actually cut. The unused `version-get`/`version-set` targets and their `version/VERSION` file are gone, as is the hand-rolled `make release` cross-compile, which produced bare binaries the Registry would reject; `make release-snapshot` builds the real artifact set locally instead.

### NOTES

- Supports Terraform Protocol version 6.
- Releases are built by goreleaser from a tag push, producing the artifact set the Terraform Registry requires: per-platform zips, a GPG-signed `_SHA256SUMS`, and the registry manifest. Changelog entries are changie fragments under `.changes/unreleased/`, batched into `CHANGELOG.md` when a release is cut.
- Built against kion-sdk-go 0.9.0 (`generated/v3_16`).
## Previous provider: 0.3.36 and earlier

The SDKv2 provider this one replaced, as published from
[kionsoftware/terraform-provider-kion](https://github.com/kionsoftware/terraform-provider-kion) up to v0.3.36.
Its version numbers are separate from the 1.x releases above.

### [0.3.36] - 2026-10-28

#### Fixed

- Fixed the `kion_saml_group_association` data source failing on every read
- It requested `/v3/idms/{id}/group-association` with the `{id}` placeholder never filled in, and offered no argument to supply the IDMS ID. A new required `idms_id` argument names the SAML IDMS whose group associations are listed
- Fixed the `kion_saml_group_association` data source example, which referenced a nonexistent `saml_group_association` attribute instead of `list` and passed regex patterns without setting `regex = true`

### [0.3.35] - 2026-09-16

#### Added

- Added cloud access role type support to `kion_project_cloud_access_role` and `kion_ou_cloud_access_role`, and to their data sources
- New `cloud_access_role_type_id` attribute selects the role type: 1 = User (default), 2 = Custom Trust, 3 = Account, 4 = Service. Types other than User are AWS only and require Kion 3.15.3, 3.16.5 or 3.17.1, depending on the release line. Note that 3.17.0 does not support them
- New `aws_iam_role_trust_policy` (Custom Trust), `aws_trusted_account_numbers` (Account), and `aws_trusted_services` (Service) attributes supply the per-type trust configuration
- New `aws_partition` and `aws_create_instance_profile` attributes control which AWS partition a non-User role syncs to and whether an IAM instance profile is created
- Added `aws_session_tags` to both cloud access role resources and `cloud_provider_ids` to `kion_project_cloud_access_role`; both have been available on the public API for some time but were never exposed by the provider
- `cloud_provider_ids` is write-only: Kion accepts it on create and update but omits it when reading a role back, so the provider cannot detect drift on it. It is not offered on the `kion_project_cloud_access_role` data source for the same reason
- Creating a non-User role now verifies the type Kion actually stored and fails with an explicit error if the installation ignored the request, rather than leaving a permanent diff
- Plan-time validation rejects per-type misconfigurations, such as a Custom Trust role without a trust policy or a non-User role with Kion users attached

#### Changed

- Building the provider from source now requires Go 1.26 or newer. `go.mod` declares `go 1.26.0`, and CI and the release pipeline build with Go 1.26.8. This does not affect practitioners using released binaries from the registry

#### Fixed

- Fixed `kion_project_cloud_access_role` and `kion_ou_cloud_access_role` planning a replacement on every run when `aws_iam_role_name` was left unset
- Kion derives a role name from `name` when the caller omits one, and the read stored it. Because `aws_iam_role_name` is `ForceNew`, that server-side value read as a change against an empty configuration and forced replacement on each plan. The attribute is now `Computed` as well, so a generated name is kept as-is. Explicitly changing the attribute still forces replacement

### [0.3.34] - 2026-04-16

#### Fixed

- Fixed `kion_custom_account` validation failing for `start_datecode` when using a computed value (e.g., `formatdate()`) (#143)
- Replaced `GetOk()` with `GetRawConfig()` in `CustomizeDiff` validation, as `GetOk()` returns false for unknown/computed values during the plan phase
- Fixed `kion_azure_policy` example in `examples/resources/kion_azure_policy/resource.tf` which produced a 400 from the Azure Policy Definition API (`Could not find member 'if' on object of type 'PolicyDefinitionProperties'`)
- The example's `policy` JSON now wraps the rule in a `policyRule` object and includes the parameter schema inside the policy JSON, matching what Azure expects in the `properties` body
- The example's top-level Terraform `parameters` attribute now uses the runtime-value format (`{ "paramName": { "value": ... } }`) instead of re-declaring the parameter schema

#### Security

- Bumped `google.golang.org/grpc` from v1.61.1 to v1.79.3 to address GHSA-p77j-4mvh-x3m3 / CVE-2026-33186 (thanks @enel1221, #146)
- Bumped Go release line from 1.22 to 1.25 (toolchain pinned to go1.25.9) to address Go stdlib CVE-2025-68121 (thanks @enel1221, #146)
- Pinned `release.yml` and `golangci-lint.yml` workflows to Go 1.25.9 so CI builds use the patched toolchain

### [0.3.33] - 2026-03-26

#### Fixed

- Fixed `kion_aws_account` failing to move newly created accounts from cache to project due to AWS service activation delays
- Account creation now verifies the account is accessible in AWS via `/v3/account-cache/{id}/status` before attempting to move it to a project
- The cache-to-project conversion now retries on all errors until the create timeout expires, handling transient AWS errors like `OptInRequired` during new account provisioning

### [0.3.32] - 2026-02-20

#### Fixed

- Fixed `kion_label` data source returning incomplete results when more than 100 labels exist in Kion
- The data source now paginates through all pages of the `/v3/label` API endpoint instead of only fetching the first page

### [0.3.31] - 2025-01-22

#### Added

- New `move_ou_settings` block for `kion_project` resource to support moving projects between OUs
- Support for `cloud_rule_setting` option: "convert" (default) or "remove" cloud rules during move
- Support for `financial_setting` option: "move" (default) or "preserve" financials during move
- Projects can now be moved between OUs by changing `ou_id` with configurable move behavior

#### Changed

- Project OU changes now use the project move API instead of requiring resource recreation
- Updated documentation and examples for project move functionality

### [0.3.30] - 2025-11-19

#### Fixed

- Fixed account field updates not being applied for Azure, Custom, and GCP accounts
- Account name, alias, and other field changes now properly persist in Kion via PATCH requests
- All account types (AWS, Azure, GCP, Custom) now use shared update logic for consistency

#### Changed

- GCP `create_mode` field is now optional and computed, preventing unnecessary drift on existing accounts
- Consolidated account update logic across all account types to eliminate code duplication

### [0.3.29] - 2025-09-30

#### Added

- New `kion_custom_account` resource for managing custom (non-cloud provider) accounts in Kion
- New resource docs for `app_config`

### [0.3.28] - 2025-09-04

#### Added

- New `kion_app_config` resource and data source for managing global Kion application configuration
- Support for configuring budget, cost savings, API key, AWS, SMTP, and permission settings
- Import support for bringing existing app-config under Terraform management

#### Fixed

- Fixed project import failure when Kion is in spend plan mode instead of budget mode

### [0.3.27] - 2025-08-27

#### Added

- Enhanced filtering capabilities for `kion_user` data source with generic filter fields (`name`, `values`, `regex`)
- Support for regex-based filtering in user data source
- Support for filtering by any user field (`id`, `username`, `enabled`) using generic filters

#### Changed

- Extended `kion_user` data source schema to support both legacy and new generic filtering approaches
- Updated examples to demonstrate both filtering methods and backwards compatibility

#### Fixed

- Fixed infinite loop in budget functionality when funding sources are optional
- Budget resources now properly handle cases where no funding source is set (funding source ID 0)
- Resolved state drift issues when budget funding sources are empty

### [0.3.26] - 2025-07-11

#### Changed

- Made `aws_iam_role_name` optional in `kion_ou_cloud_access_role` and `kion_project_cloud_access_role` resources
- Cloud access roles can now be created without specifying an AWS IAM role name
- Updated API models to use optional string pointers with proper nil handling

#### Fixed

- Fixed linting warnings for deprecated `GetOkExists` usage by adding appropriate ignore comments
- Fixed error message formatting to follow Go conventions (lowercase, no punctuation)
- Added proper `omitempty` JSON tags for optional fields

### [0.3.25] - 2025-07-02

#### Fixed

- Fixed budget amount calculation for projects using monthly budget data entries
- Resolved perpetual drift issue where Terraform would detect changes in budget amounts
- Added proper handling for Kion API responses that omit the 'amount' field when using monthly data
- Added validation to ensure monthly budget data totals match the declared budget amount
- Fixed floating-point precision issues causing false drift detection (e.g., 5000.000000000001)
- Added detection for auto-generated monthly budget entries to prevent false drift

#### Added

- Budget amount validation during plan phase using CustomizeDiff
- Helper functions for budget calculations: `IsAutoGeneratedBudgetData()`, `AlmostEqual()`, `roundToTwoDecimals()`
- Clear error messages when monthly budget totals don't match declared amounts

#### Changed

- Improved budget amount field description to clarify validation requirements

### [0.3.24] - 2024-05-08

#### Fixed

- Updated Funding Sources Example Documentation to use the correct permission_scheme_id

### [0.3.23] - 2024-05-02

#### Fixed

- Resolved concurrent OU creation issue causing database constraint violations
- Prevented "Duplicate entry" errors when creating multiple OUs in Terraform
- Added mutex synchronization for OU hierarchy operations to maintain data integrity

### [0.3.22] - 2024-03-13

#### Added

- Enhanced account location detection with explicit location handling in state
- Added support for preserving original account IDs during import operations
- Improved logging with additional context for account operations

#### Changed

- Refactored account location handling to prioritize explicitly set locations
- Updated account import logic to better handle account_id and account_cache_id prefixes
- Enhanced documentation structure and examples across all resources
- Modernized provider configuration with clearer environment variable guidance

#### Fixed

- Improved account lookup logic to prevent unnecessary API calls
- Enhanced error handling for account operations with better context
- Fixed account ID handling during import operations

### [0.3.21] - 2024-28-28

#### Added

- New `kion_project_note` resource and data source for managing project notes
- New filtering options for AWS IAM policies:
  - `query` parameter for name matching
  - `policy_type` filter for AWS managed policies
  - Pagination support with `page` and `page_size`
- Validation for `start_datecode` format in AWS accounts (must be YYYY-MM)

#### Changed

- Improved account location detection to handle both project and cache locations
- Updated documentation structure:
  - Moved examples from README to dedicated examples directory
  - Added organized list of available examples by resource type
  - Updated AWS IAM policy examples to demonstrate data source usage

#### Fixed

- Account reading logic to better handle accounts that move between projects and cache
- Validation for date code formats in AWS accounts

#### Documentation

- Added new examples for project notes
- Updated AWS IAM policy documentation with new filtering options
- Improved organization of examples by resource type
- Added missing fields in various resource documentation

### [0.3.20] - 2024-02-20

#### Added

- Added support for custom variables with new resources `kion_custom_variable` and `kion_custom_variable_override` and corresponding data sources
- Added v4 API support for IAM policy list operations with enhanced filtering capabilities (query, policy_type, pagination)
- Added concurrent CloudFormation template sync option to cloud rules via `concurrent_cft_sync` parameter to control parallel/sequential template deployment
- Added funding source ID field to funding source data source output for better resource referencing
- Added support for non-allocation mode in funding sources by making `ou_id` optional

#### Changed

- Enhanced account management with improved import functionality and validation:
  - Simplified account import process with multiple import methods
  - Added better validation for account moves between projects
  - Improved error messages and logging for account operations
- Improved project budget management:
  - Added proper state tracking for budget changes
  - Enhanced validation for budget periods and date formats
  - Implemented proper ordering of budget operations
- Refactored AWS IAM policy handling:
  - Maintained v3 endpoint compatibility for single resource operations
  - Added v4 API support for enhanced filtering and pagination
- Standardized error handling across resources using consistent error wrapping patterns

#### Fixed

- Fixed compliance check removal in compliance standards to properly handle "compliance check not found" errors
- Fixed project budget updates to properly apply changes and prevent overlapping periods
- Fixed permission scheme handling in funding sources to prevent unnecessary updates
- Fixed account cache operations and removed unused testing functions
- Improved handling of budget period modifications to prevent conflicts during splits or updates

### [0.3.19] - 2024-09-26

### What's Changed

#### Fixed

- Ensured consistent handling of funding source `amount` fields across API models and Terraform schema definitions to prevent type mismatches.

#### Changed

- Updated the `amount` field in the `FundingSource` schema from `TypeInt` to `TypeFloat` for more accurate representation of funding values.
- Modified the `models_funding_source.go` to reflect the new `float64` type for the `amount` field in all funding source-related structures.
- Adjusted the `resourceFundingSource` and `resourceFundingSourceUpdate` methods to use `float64` for the `amount` attribute during creation and updates.

### [0.3.18] - 2024-08-15

### What's Changed

#### Fixed

- Corrected a regression in the `NewFilterable` function that caused issues with filter list processing and key-value handling for HCL-defined filters.
- Fixed formatting issues in the `clone_resource` function to prevent unnecessary line breaks in resource names.

### [0.3.17] - 2024-08-13

### What's Changed

#### Added

- Introduced permission mapping management for funding sources, OUs, projects, and global permissions.
- Added support for managing webhooks and retrieving user data with new resources and data sources.
- Account alias field added to all account-related resources for better identification.

#### Changed

- Refactored helper functions for improved efficiency and code quality.
- Enhanced error handling across account-related resources and project enforcement.

#### Removed

- Deprecated `OptionalString` function in favor of the more versatile `OptionalValue`.

#### Fixed

- Standardized the handling of optional fields across various resources.
- Resolved issues with redundant error handling by centralizing logic.
- Normalized JSON fields in webhook resources to prevent unnecessary changes.

### [0.3.16] - 2024-06-07

### What's Changed

#### Added

- Added GCP IAM Roles to Cloud Access Roles [pull/77](https://github.com/kionsoftware/terraform-provider-kion/pull/77)
- Add Project Enforcement Support [pull/75](https://github.com/kionsoftware/terraform-provider-kion/pull/75)
- Add golangci-lint Workflow [pull/71](https://github.com/kionsoftware/terraform-provider-kion/pull/71)
- Added `.golangci.yml` configuration file for linting and then fixed nearly everything that it found [pull/68](https://github.com/kionsoftware/terraform-provider-kion/pull/68)
- Added New Azure Policies Support [pull/74](https://github.com/kionsoftware/terraform-provider-kion/pull/74)

#### Changed

- Update `.gitignore` and enhanced `README.md` for Terraform Importer Script [pull/67](https://github.com/kionsoftware/terraform-provider-kion/pull/67)
- Refactor Kion Client Codebase [pull/69](https://github.com/kionsoftware/terraform-provider-kion/pull/69)

### [0.3.15] - 2024-05-17

- No changes.  This change was needed to get the Terraform Registry back in sync with Github.

### [0.3.14] - 2024-05-15

- Fixed an issue where the provider would attempt to update project labels even when labels were not defined in the HCL. The provider now checks if labels are explicitly set and non-empty before making update requests, preventing unnecessary API calls and avoiding the `405 Method Not Allowed error` from any Kion version before v3.7.7.
- Added examples into example directory that previously only were in the `README.md` file.  This allows each resources to have more example information directly in the resource docs.

### [0.3.13] - 2024-04-09

- Changed `aws_cloudformation_templates` and `azure_arm_template_definitions` from `TypeSet` to `TypeList` in the Terraform provider schema to preserve the order of templates as specified in Terraform configurations.
- Updated the `resourceCloudRuleCreate` function to build lists of template IDs directly from the Terraform state, ensuring order preservation during cloud rule creation.
- Modified the `resourceCloudRuleRead` function to correctly parse and set ordered template lists in the Terraform state, maintaining consistency with the API.
- Bump Go Version to 1.22

### [0.3.12] - 2024-03-29

- Introduced several enhancements and refactoring changes to the Makefile and the Terraform provider. The key changes include adding a development build option, introducing a sync mechanism for AWS account creation, allowing more than one `kion_aws_account` to be created at once.

### [0.3.11] - 2024-03-20

- Introduced retry logic in the AWS account creation process within the Terraform Provider. This improvement is designed to handle scenarios where the account creation is temporarily hindered by ongoing AWS service operations. The logic includes a default retry count of three attempts with a delay of 30 seconds between each attempt, enhancing the robustness of the account creation process in fluctuating AWS service conditions.

### [0.3.10] - 2024-02-29

- Added support for `azure_role_definitions` in OU Cloud Access Roles. This enhancement includes the ability to define and manage Azure role definitions directly within Terraform.

### [0.3.9] - 2024-02-15

- Fixed an issue where users were unable to create Kion Azure Policy resources due to certain attributes (`name`, `description`, `policy`, `parameters`) being set to "Read-Only" in the Terraform Kion Provider. This fix involves changing the `Computed: true` parameter for these attributes to either `Required` or `Optional`, based on the API documentation for azure-policy. This change allows for the proper creation and management of Azure Policy resources within Terraform.

### [0.3.8] - 2024-01-03

- Fixed an issue that prevented importing existing Kion accounts into the terraform state.  When importing existing accounts, the user should specify the account ID using an `account_id=` or `account_cache_id=` prefix to tell the terraform provider whether the provided ID is an account ID or a cached account ID.  See the README for more information.

### [0.3.7] - 2023-12-21

- Added `kion_aws_account`, `kion_gcp_account` and `kion_azure_account` resources ([#32](https://github.com/kionsoftware/terraform-provider-kion/pull/32))
- Added `kion_account` and `kion_cached_account` data sources
- Upgrade terraform-plugin-sdk from v2.10.0 to v2.30.0

### [0.3.6] - 2023-11-20

- Added a `kion_label` resource ([#31](https://github.com/kionsoftware/terraform-provider-kion/pull/31))
- Added a `labels` attribute to OUs, Projects, Funding sources and cloud rules

### [0.3.5] - 2023-09-22

- Added Funding Source resource ([#25](https://github.com/kionsoftware/terraform-provider-kion/pull/25))

### [0.3.4] - 2023-03-27

- Added Tags support to Cloudformation template resource (#24)
- Fix KeyError bug in import script (#7)

### [0.3.3] - 2023-03-10

- Added Better documentation around datasource filtering.
- Fixed a bug where Kion's provider would panic when applying some resources.

### [0.3.2] - 2023-03-09

- Added documentation that clarifies that either an owner user or owner group must be defined for some resources.
- Added better error handling when a user attempts to create a Kion resource without an owner user or owner group.
- Fixed a bug where Terraform expected resources to be in a specific order on unordered resources.

### [0.3.1] - 2023-02-27

- Fix the description for IAM policies in the documentation to be more accurate.
- Added clarity around creating compliance standards.
- Allow project creation with budget when enabled in Kion.

### [0.3.0] - 2022-02-25

- Added Support for creating, updating, and deleting resources for: AWS Service Control Policies.
- Added Support for adding and removing AWS Service Control Policies on Project and OU Cloud Rules.
- Added Support for creating, updating, and deleting resources for: Azure ARM Templates.
- Added Support for adding and removing Azure ARM Templates on Project and OU Cloud Rules.
- Added Support for creating, updating, and deleting resources for: Azure Role Definitions.
- Added Support for adding and removing Azure Role Definitions on Project and OU Cloud Rules.
- Rebrand from cloudtamer.io to Kion.
- Change provider name: `cloudtamer-io/cloudtamerio` to `kionsoftware/kion`.
- Change resource and data source names prefix: `cloudtamer_` to `kion_`.
- Change environment variables: `CLOUDTAMERIO_URL` and `CLOUDTAMERIO_APIKEY` to `KION_URL` and `KION_APIKEY`, respectively.
- Made the `created_by_user_id` field for Compliance Checks optional. This field will default to the requesting user's ID if not specified.

### [0.2.0] - 2021-11-19

- Added Support for creating, updating, and deleting resources for: user groups.
- Added Support for creating, updating, and deleting resources for: SAML IDMS user group associations.
- Added Support for creating, updating, and deleting resources for: Projects
- Added Support for creating, updating, and deleting resources for: Google Cloud IAM Roles.
- Added Support for adding and removing Google Cloud IAM Roles on Project and OU Cloud Rules.
- Fix several requests that use the wrong user & user group IDs to remove owners from a resource.

### [0.1.4] - 2021-08-09

- Added Support for creating, updating, and deleting resources for: OUs. (Requires Kion v2.31.0 or newer)

### [0.1.3] - 2021-06-29

- Fix bug on project cloud access role creation so 'apply_to_all_accounts' and 'accounts' fields are mutually exclusive.
- Remove unused errors throughout the code.

### [0.1.2] - 2021-04-01

- Added Support for creating, updating, and deleting resources for: OU cloud access roles and project cloud access roles.
- Fix bug on compliance standard creation so compliance checks are attached during creation instead of requiring another `terraform apply`.
- Fix bug on cloud rule creation so associated items are attached during creation instead of requiring another `terraform apply`.

### [0.1.1] - 2021-03-30

- Added Ability to import resources using `terraform import`.

### [0.1.0] - 2021-02-08

- Initial release of the provider.
- Added Support for creating, updating, and deleting resources for: AWS CloudFormation templates, AWS IAM policies, Azure policies, cloud rules, compliance checks, and compliance standards.
- Added Support for querying data sources for: AWS CloudFormation templates, AWS IAM policies, Azure policies, cloud rules, compliance checks, and compliance standards.
## [0.1.0] - 2025-10-21

### NOTES

- Initial development version
- CI/CD infrastructure setup complete
- Automated changelog generation configured
