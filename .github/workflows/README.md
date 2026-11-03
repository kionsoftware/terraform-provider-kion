# GitHub Actions Workflows

Two workflows. `ci.yml` gates changes; `release.yml` publishes them.

## ci.yml

**Triggers:** pull requests, and pushes to `main`.

Every job runs independently and in parallel; `ci` is a required-status
aggregator that fails if any of them did, so branch protection only needs to
require the single `ci` check.

| Job | What it enforces | Local equivalent |
|---|---|---|
| `fmt` | `gofmt -s` is clean | `make ci-fmt` |
| `vet` | `go vet ./...` | `make ci-vet` |
| `lint` | golangci-lint, config in `.golangci.yml` | `make ci-lint` |
| `test-unit` | unit tests, race detector, coverage | `make ci-test` |
| `modules` | generated Terraform modules build, validate, test, and are not stale | `make modules-check` |
| `docs` | `docs/` and `examples/` are not stale | `make docs-check` |
| `acctest-config` | acceptance-test HCL matches the provider schema | `make ci-acctest-config` |
| `align-check` | no schema/SDK type mismatch or missing flex converter beyond `codegen/align_baseline.yaml`, and no stale entry in it | `make align-check` |
| `internal-refs` | no internal paths/hostnames in tracked files | `scripts/check-no-internal-refs.sh` |
| `secrets` | credential scan over the tree | `make ci-secrets` |
| `codeql` | GitHub CodeQL analysis for Go | GitHub only |

`make ci` runs the first four plus `acctest-config` locally. The Lefthook pre-push hook runs it
before every push, so CI should rarely be the first place a failure appears.

### The `modules` job

This one is worth understanding because it is the slowest and the least
obvious. It builds the provider, publishes it into a local filesystem mirror,
regenerates `modules/` from the provider schema, injects docs with
`terraform-docs`, and then asserts the working tree is unchanged. A diff means
someone edited a resource without regenerating, so `modules/` no longer matches
the schema it is supposed to describe. Regenerate and commit rather than
hand-editing the output.

### `align-check`

`kalign check` pairs each schema model with the SDK struct whose json tags
best cover its attributes, and reports two actionable signals: a primitive
type mismatch and a missing `flex` converter. Some are expected, because the
provider binds an attribute through a mechanism a one-struct comparison cannot
see (a nested object's id, a raw wire struct). Those are recorded with a reason
in `codegen/align_baseline.yaml`. The job fails on a finding the baseline does
not record, on an entry that no longer occurs, and on an entry still marked
TODO. `make align-baseline` rewrites the file, keeping existing reasons;
`make align-report` prints the full report, including the advisory DRIFT lines.
It needs only the SDK module go.mod pins, not the OpenAPI spec.

### `internal-refs`

This repository is public and takes outside contributions. The check blocks
internal source paths, hostnames, and planning notes from being published.
they are not secrets, but they leak detail nobody outside can act on, and once
pushed they exist in every fork. The pattern list and the rationale for each
entry live in `scripts/check-no-internal-refs.sh`.

## release.yml

**Triggers:** tags matching `v*`.

Runs GoReleaser (config: `.goreleaser.yml`), which cross-compiles the provider,
signs the checksum manifest with GPG, and attaches the artifacts to a GitHub
release in the layout the Terraform Registry expects.

**Required secrets.** The workflow fails without them:

| Secret | Purpose |
|---|---|
| `GPG_PRIVATE_KEY` | ASCII-armored private key whose public half is registered with the Terraform Registry |
| `PASSPHRASE` | passphrase for that key |

To cut a release: `make release-prepare`, then commit, tag and push.

```bash
git add .changes CHANGELOG.md && git commit -m "release X.Y.Z"
git tag -a vX.Y.Z -m vX.Y.Z && git push origin main vX.Y.Z
```

`make release-snapshot` builds the artifacts locally without publishing, which
is the cheap way to check a GoReleaser change before tagging.

## Changing a workflow

Validate before pushing. A broken workflow file is only reported after it is
on the default branch:

```bash
actionlint .github/workflows/*.yml
goreleaser check
```

Action versions are pinned to a major (`@v7`). Review them periodically;
Dependabot is not configured for this repository.
