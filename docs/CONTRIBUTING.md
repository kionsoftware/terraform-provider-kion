# Contributing

Thanks for helping improve the Kion Terraform provider. Nearly all of it is
generated from codegen inputs, so most changes go to `codegen/`, not to the
generated Go files. [DEVELOPMENT.md](DEVELOPMENT.md) explains the layout and the
generator, and [`codegen/README.md`](../codegen/README.md) documents its inputs.

## Workflow

1. Create a branch from `main`
2. Make changes and ensure `make ci` passes
3. Regenerate any affected generated surfaces (see [Code generation](DEVELOPMENT.md#code-generation))
4. Run `make changelog-new` if the change is user-visible
5. Open a pull request

For new resources, scaffold with `kgen service`, add the `codegen/` entries, and regenerate. Do not hand-write CRUD.

## Prerequisites and commands

See [Development](DEVELOPMENT.md#development) for the toolchain and the `make`
targets, and [TESTING.md](TESTING.md) and
[ACCEPTANCE-TESTING.md](ACCEPTANCE-TESTING.md) for running the tests.
Acceptance tests need a live Kion install and run only from your machine. CI
never runs them.

## Releases

Maintainers cut releases from `main`; see [RELEASING.md](RELEASING.md).
