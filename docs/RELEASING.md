# Releasing

**The provider's version is its own. It does not track Kion's.** Encoding the Kion release in the provider version leaves nowhere to put a provider patch: a fix for the provider against Kion 3.16.4 has no number available, and `~> 3.16` would read as a Kion pin while behaving as a provider pin.

Kion compatibility is a **runtime** mechanism instead. The provider reads `GET /api/version` at configure time and gates each resource against the Kion range its defining API operation exists in (`codegen/version_support.yaml`), failing with a clear diagnostic rather than a confusing API error. The supported window mirrors the SDK's: the current Kion release and three back.

| Change | Bump |
|---|---|
| Provider bug fix | patch, any time, independent of Kion |
| New Kion release adds resources or attributes | minor |
| A provider resource or attribute is removed or renamed | major |

Releases come off `main`; there are no per-Kion release branches. If a released line ever needs a back-patch, branch `release/<provider-major>.x`, keyed to the provider's major and never to a Kion version.

## Changelog

Entries are YAML fragments under `.changes/unreleased/`, one per change, so concurrent pull requests never conflict on `CHANGELOG.md`. `CHANGELOG.md` holds released versions only.

```bash
make changelog-new        # add an entry for your change (interactive)
make changelog-preview    # what the next release notes would say
```

## Cutting a release

```bash
make release-prepare RELEASE_VERSION=1.0.0   # make ci, then batch fragments into CHANGELOG.md
git add .changes CHANGELOG.md && git commit -m "release 1.0.0"
git tag -a v1.0.0 -m v1.0.0 && git push origin main v1.0.0
```

The tag is the version. Nothing in the repo records it. `make version` reports the latest release tag (`0.0.0-dev` on an untagged tree), and that is what `make install` uses for the local plugin path. `make release-snapshot` builds the real artifact set locally without publishing.

The tag triggers `release.yml`, which builds with goreleaser, signs `_SHA256SUMS` with the release GPG key, and attaches the registry manifest, the exact artifact set the Registry requires. Release notes are the changie version file for that tag.

Publishing needs two repository secrets: `GPG_PRIVATE_KEY` (the ASCII-armoured private key) and `PASSPHRASE` (its passphrase), with the matching public key registered on the Registry provider page.

`github.com/kionsoftware/kion-sdk-go` resolves from the module proxy, so the release workflow needs no SDK checkout, and the build uses exactly the version `go.mod` pins. Do not reintroduce a clone step: a branch checkout would silently override that pin and make the release unreproducible.
