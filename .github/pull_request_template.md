# Description

<!-- What does this change do, and why? -->

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would change existing behavior)
- [ ] Documentation update
- [ ] Code generation update (spec or generator changes)
- [ ] Maintenance (dependency updates, refactoring, etc.)

## Related Issues

<!-- e.g. "Fixes #123" or "Relates to #456" -->

Fixes #

## Changelog Entry

<!--
Required for user-facing changes; not required for docs, tests, or refactors
with no behavior change.

This repo uses changie. Run `make changelog-new` and answer the prompts. It
writes a YAML file under `.changes/unreleased/`. Commit that file with your
change. Do not edit CHANGELOG.md by hand; it is generated at release time.

Kinds: BREAKING CHANGES, FEATURES, NEW RESOURCES, ENHANCEMENTS, BUG FIXES, NOTES

Preview what the next release notes would say: `make changelog-preview`
-->

- [ ] Ran `make changelog-new` and committed the file under `.changes/unreleased/`, or this change is not user-facing

## Testing

<!-- How did you verify this? -->

### Test Configuration

```hcl
# Terraform configuration used to test this change
```

### Test Output

```
# Output from `make test`, acceptance tests, or manual verification
```

## Checklist

- [ ] `make ci` passes locally (fmt, vet, lint, unit tests)
- [ ] I have performed a self-review of my own code
- [ ] I have commented anything non-obvious, explaining *why* rather than *what*
- [ ] Documentation is updated where behavior changed
- [ ] Tests cover the fix or the new behavior
- [ ] No internal paths, hostnames, or planning notes are included. This
      repository is public (`scripts/check-no-internal-refs.sh` enforces this)

## For Code Generation Changes

<!-- Only if this PR touches the generator or the OpenAPI spec -->

- [ ] Spec changes are described above
- [ ] Regenerated with `make generate` and committed the result
- [ ] `make modules-check` passes (generated modules are not stale)

## Additional Context

<!-- Anything else a reviewer should know -->

---

<!--
Thanks for contributing to the Kion Terraform Provider.

- Migrating from the previous provider? See docs/guides/upgrading-from-0.3.md
- CI details: .github/workflows/README.md
-->
