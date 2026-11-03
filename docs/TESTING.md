# Testing the Kion Terraform Provider

This document describes how to run tests for the Kion Terraform provider.

## Test Types

The provider includes two types of tests:

### Unit Tests

Unit tests test individual functions and helpers without making API calls. They're fast and don't require credentials.

**Run unit tests:**

```bash
make test
```

Or directly:

```bash
go test -v ./internal/...
```

### Acceptance Tests

Acceptance tests create real infrastructure using a live Kion instance. They validate the full lifecycle of resources (create, read, update, delete).

**⚠️ WARNING:** Acceptance tests create and destroy real infrastructure and will incur costs.

## Running Acceptance Tests

### Prerequisites

1. **Go 1.21+** installed
2. **Terraform CLI 0.12.26+** installed
3. **Access to a Kion instance** (preferably a dedicated test/dev environment)
4. **Valid Kion credentials**

### Environment Setup

Set the following environment variables:

```bash
# Required: Kion API endpoint
export KION_API_URL="https://your-kion-instance.com"

# Required: At least one authentication method
export KION_API_KEY="your-api-key"
# OR
export KION_AUTH_TOKEN="your-auth-token"

# Required: Enable acceptance tests
export TF_ACC=1
```

#### Per-resource environment variables

A resource whose test needs something install- or cloud-specific reads it from
the environment through `acctest.RequireEnv`, which **skips** that test when the
variable is unset rather than sending an empty attribute. The list is generated
from the `env_args` entries (values the config uses) and `require_env` entries
(prerequisites checked in `PreCheck` whose values it does not) in
`codegen/test_values.yaml`; regenerate this table after changing them:

```bash
python3 - <<'EOF'
import yaml
r = yaml.safe_load(open("codegen/test_values.yaml"))["resources"]
for name in sorted(r):
    env = (r[name].get("env_args") or []) + [e["name"] for e in r[name].get("require_env") or []]
    if env:
        print(f"| `kion_{name}` | {', '.join('`%s`' % e for e in env)} |")
EOF
```

| Resource | Variables |
|---|---|
| `kion_account_cache` | `KION_ACC_PAYER_ID`, `KION_ACC_AWS_ACCOUNT_NUMBER` |
| `kion_account_linkage` | `KION_ACC_AZURE_PAYER_ID` |
| `kion_ami` | `KION_ACC_ACCOUNT_ID`, `KION_ACC_AWS_AMI_ID` |
| `kion_azure_account` | `KION_ACC_AZURE_PAYER_ID`, `KION_ACC_AZURE_SUBSCRIPTION_UUID` |
| `kion_azure_arm_template` | `KION_ACC_AZURE_REGION_ID`, `KION_ACC_AZURE_PAYER_ID`, `KION_ACC_AZURE_SUBSCRIPTION_UUID` |
| `kion_azure_role` | `KION_ACC_AZURE_PAYER_ID` |
| `kion_billing_rule` | `KION_ACC_BILLING_SOURCE_ID` |
| `kion_billing_source` | `KION_ACC_AWS_ACCOUNT_NUMBER`, `KION_ACC_AWS_BILLING_ROLE`, `KION_ACC_AWS_BILLING_BUCKET`, `KION_ACC_AWS_REGION` |
| `kion_billing_source_anthropic` | `KION_ACC_ANTHROPIC_ADMIN_KEY` (a real Anthropic Admin API key: Kion validates it with Anthropic) |
| `kion_billing_source_aws` | `KION_ACC_AWS_ACCOUNT_NUMBER`, `KION_ACC_AWS_BILLING_ROLE` |
| `kion_billing_source_gcp` | `KION_ACC_GCP_BILLING_ACCOUNT_ID`, `KION_ACC_GCP_SERVICE_ACCOUNT_ID` |
| `kion_billing_source_govcloud` | `KION_ACC_PAYER_ID`, `KION_ACC_AWS_ACCOUNT_NUMBER` |
| `kion_billing_source_oci` | `KION_ACC_OCI_TENANCY_OCID`, `KION_ACC_OCI_USER_OCID`, `KION_ACC_OCI_FINGERPRINT`, `KION_ACC_OCI_PRIVATE_KEY`, `KION_ACC_OCI_REGION` |
| `kion_category` | `KION_ACC_BILLING_SOURCE_ID` |
| `kion_custom_account` | `KION_ACC_CUSTOM_PAYER_ID`, `KION_ACC_CUSTOM_ACCOUNT_NUMBER` |
| `kion_funding_source_permission_mapping` | `KION_ACC_APP_ROLE_ID` |
| `kion_gcp_account` | `KION_ACC_GCP_PAYER_ID`, `KION_ACC_GCP_PROJECT_ID` |
| `kion_gcp_service_account` | `KION_ACC_GCP_SERVICE_ACCOUNT_EMAIL` |
| `kion_global_permission_mapping` | `KION_ACC_GLOBAL_APP_ROLE_ID` |
| `kion_idms_group_association` | `KION_ACC_SAML_IDMS_ID` |
| `kion_ou_permission_mapping` | `KION_ACC_APP_ROLE_ID` |
| `kion_project_line_item` | `KION_ACC_BILLING_SOURCE_ID` |
| `kion_project_permission_mapping` | `KION_ACC_APP_ROLE_ID` |
| `kion_saml_group_association` | `KION_ACC_SAML_IDMS_ID` |
| `kion_service_catalog` | `KION_ACC_ACCOUNT_ID`, `KION_ACC_AWS_PORTFOLIO_ID`, `KION_ACC_AWS_REGION` |

Tests for resources not listed need only `KION_API_URL` and a credential.

### Running Tests

**Run all acceptance tests:**

```bash
make testacc
```

**Run specific resource tests:**

```bash
TF_ACC=1 go test -v -run TestAccKionLabel ./internal/provider/
```

**Run with verbose output:**

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

### Test Execution Details

- Tests run in parallel (default: 4 parallel tests)
- Default timeout: 120 minutes
- Tests use the `test-acc` prefix for all created resources
- Tests automatically clean up resources after execution

## Sweepers

Sweepers clean up orphaned test resources left behind by failed tests — **for the
37 resource types that have one.** 14 more cannot be swept and register nothing;
see [Resources without a sweeper](#resources-without-a-sweeper). Read that list
before assuming a clean run means a clean install.

**Run sweepers:**

```bash
make sweep
```

Or directly:

```bash
TF_ACC=1 go test -v -timeout 30m ./internal/sweep/... -sweep=all
```

The package path must come **before** `-sweep`: `go test` stops parsing package
arguments at the first flag it does not recognize, so a trailing `./...` is
handed to the test binary and the root package is run instead. `internal/sweep`
is also the only package whose `TestMain` calls `resource.TestMain`, which is
what actually runs registered sweepers.

**⚠️ WARNING:** Sweepers will destroy infrastructure. Only use in development accounts.

### How Sweepers Work

A sweeper lists its resource type, keeps the records whose string fields start
with `test-acc`, and deletes those by id. `internal/service/<name>/sweep.go` is
generated by `kgen crud` from the resource's resolved list and delete
operations — never hand-written. Where the collection is reachable only under a
parent (`kion_ou_note` under `kion_ou`, say), the sweeper walks the parents
first; that is declared per resource as `sweep_parent` in
`codegen/crud_archetypes.yaml`.

`internal/sweep/sweep_test.go` blank-imports each package that registers one,
which is what links its `init` into the sweep binary. `TestSweeperRegistration`
fails if that list drifts from the generated files.

### Resources without a sweeper

A resource that cannot be enumerated or cannot be deleted through the API gets
**no sweeper at all**, and its `sweep.go` says why. This is deliberate: a
registered sweeper whose body returned `nil` reported success to `make sweep`
while orphaned `test-acc` records piled up, which is worse than having none.

| Resource | Why |
|---|---|
| `kion_service_catalog`, `kion_user`, `kion_webhook` | no collection endpoint is configured, so test resources cannot be enumerated. Their **delete** works: all three delete through a private route declared in `codegen/private_endpoints.yaml`, which the generator now uses instead of refusing the resource. Give any of them a list endpoint in `codegen/config_overrides.yaml` and its sweeper generates itself. |
| `kion_idms_open_id` | neither a delete endpoint nor a resolvable collection |
| `kion_aws_resource_tag`, `kion_ou_cloud_access_role_exemption`, `kion_project_cloud_access_role_exemption` | no read, so no collection to enumerate |
| `kion_billing_source_govcloud`, `kion_budget`, `kion_idms_open_id_access_rule`, `kion_idms_open_id_group_association` | the configured collection endpoint is a by-id read, not a list |
| `kion_compliance_control` | parent-scoped collection plus a two-param delete the sweeper cannot source |
| `kion_billing_rule` | `BillingRule` has an empty OpenAPI schema, so no id or name to match |
| `kion_scope_criteria` | compound-key sub-resource with no independent collection |

Several more resource types ship no `sweep.go` because their archetype
(bespoke, association, singleton, parent-list, raw-HTTP) has no sweeper
renderer yet: `kion_account`, `kion_aws_account`, `kion_permission_scheme`,
`kion_project_note`, `kion_project_enforcement`, the permission mappings, and
`kion_custom_variable_override` among them.

`make crud-force` prints the current unswept set and its reasons at the end of
the run, so the table above can be regenerated rather than trusted.

## Writing Tests

### Test File Structure

Test files follow Go conventions:

```
internal/service/label/
├── label.go                      # Resource CRUD
├── label_test.go                 # Label resource tests
├── label_data_source.go
├── label_data_source_test.go
└── sweep.go                      # Sweeper, or the note saying why there is none

internal/sweep/
├── sweep_test.go                 # TestMain + the blank imports that link sweepers
└── registration_test.go          # Guards against no-op and unlinked sweepers
```

### Example Acceptance Test

```go
func TestAccKionLabel_basic(t *testing.T) {
    resourceName := "kion_label.test"

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckKionLabelDestroy,
        Steps: []resource.TestStep{
            // Create and Read
            {
                Config: testAccKionLabelConfig_basic("test-acc-label", "environment", "test"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckKionLabelExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "key", "test-acc-label"),
                    resource.TestCheckResourceAttr(resourceName, "value", "environment"),
                ),
            },
            // ImportState testing
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
            },
            // Update and Read
            {
                Config: testAccKionLabelConfig_basic("test-acc-label", "production", "blue"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckKionLabelExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "value", "production"),
                ),
            },
        },
    })
}
```

### Test Best Practices

1. **Prefix test resources** with `test-acc` for sweeper identification
2. **Use random names** to avoid conflicts: `acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)`
3. **Test the full lifecycle:** Create → Read → Update → Delete
4. **Test import functionality** for all resources
5. **Clean up in CheckDestroy** to verify resources are properly deleted
6. **Use aggregate checks** to run all checks even if one fails

## Common Issues

### Authentication Failures

**Error:** `KION_API_URL must be set for acceptance tests`

**Solution:** Ensure environment variables are set:

```bash
export KION_API_URL="https://your-instance.com"
export KION_API_KEY="your-key"
```

### Resource Already Exists

**Error:** `Label already exists`

**Solution:** Run sweepers to clean up orphaned resources:

```bash
make sweep
```

### Test Timeout

**Error:** `test timed out after 30m`

**Solution:** Increase timeout:

```bash
TF_ACC=1 go test -v -timeout 180m ./internal/provider/
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Acceptance Tests

on:
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 2 * * *' # Run nightly

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run Acceptance Tests
        env:
          TF_ACC: 1
          KION_API_URL: ${{ secrets.KION_API_URL }}
          KION_API_KEY: ${{ secrets.KION_API_KEY }}
        run: make testacc

      - name: Run Sweepers (on failure)
        if: failure()
        env:
          TF_ACC: 1
          KION_API_URL: ${{ secrets.KION_API_URL }}
          KION_API_KEY: ${{ secrets.KION_API_KEY }}
        run: make sweep
```

## Resources

- [Terraform Plugin Testing Documentation](https://developer.hashicorp.com/terraform/plugin/testing)
- [Plugin Framework Testing](https://developer.hashicorp.com/terraform/plugin/framework/acctests)
- [Testing Patterns](https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns)
