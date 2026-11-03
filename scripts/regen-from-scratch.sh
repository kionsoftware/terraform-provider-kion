#!/usr/bin/env bash
#
# Throw internal/service away and rebuild it from codegen/, then report whether
# the result matches what was committed.
#
# The claim this checks is the provider's central one: nothing under
# internal/service is hand-maintained, so the whole tree can be deleted and
# rebuilt from the codegen inputs. It is not self-evidently true, and four
# things break a naive attempt -- all invisible during ordinary incremental
# work:
#
#   1. Bootstrap. `go run ./cmd/kgen` compiles the packages it is about to
#      write, so once the tree is gone the generator cannot run at all. kgen has
#      to be built BEFORE the wipe. A "regeneration did not restore it" result
#      is this trap until proven otherwise.
#
#   2. Scaffolding. The generators enumerate what to write from the directories
#      present under internal/service, so a removed directory is a resource that
#      no longer exists as far as they are concerned. `kgen service` recreates
#      them, and also re-registers each package in internal/provider.
#
#   3. Convergence. The generators feed each other: `kgen crud` needs the schema
#      models, and the schema tests plus some version gates read generated
#      source for a resource's constructor. One pass leaves schema tests and
#      version gates unwritten -- silently, because the tree still builds and
#      the missing files are tests that no longer exist to fail. Two passes
#      reach a fixed point.
#
#   4. Stubs. `kgen service` scaffolds a data source for every package, which a
#      resource-only archetype never overwrites. Generation prunes what it did
#      not emit, so the stub does not survive as output no input asked for.
#
# MODE=files wipes only files carrying a generated header, leaving directories.
# That is the weaker check: it passes even when a file with no header sits in
# the tree unreproduced, which is how 107 hand-written acceptance tests went
# unnoticed. The default wipes everything.
#
# Run from the repository root. Requires spec/openapi3.json (make refresh-spec),
# which is gitignored, so this cannot run in CI as-is.
set -euo pipefail

cd "$(dirname "$0")/.."

MODE="${MODE:-all}"

if [[ ! -f spec/openapi3.json ]]; then
	echo "spec/openapi3.json is missing; run 'make refresh-spec' first" >&2
	exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
	echo "the working tree has uncommitted changes; commit or stash them first" >&2
	echo "(this script wipes internal/service and compares against HEAD)" >&2
	exit 1
fi

SDK_DIR="$(go list -m -f '{{.Dir}}' github.com/kionsoftware/kion-sdk-go)"
KGEN="$(mktemp -d)/kgen"

echo "==> building kgen before the wipe (it compiles what it generates)"
go build -o "$KGEN" ./cmd/kgen

if [[ "$MODE" == "files" ]]; then
	echo "==> wiping generated files, keeping directories"
	wiped=0
	while IFS= read -r f; do
		if head -3 "$f" | grep -q 'Code generated'; then
			rm "$f"
			wiped=$((wiped + 1))
		fi
	done < <(find internal/service -name '*.go')
	echo "    removed $wiped generated files"
else
	echo "==> wiping internal/service entirely"
	wiped="$(find internal/service -name '*.go' | wc -l | tr -d ' ')"
	rm -rf internal/service
	echo "    removed $wiped files and every package directory"

	echo "==> scaffolding a package per resource"
	python3 - "$KGEN" <<'PY'
import subprocess, sys, yaml
kgen = sys.argv[1]
names = sorted((yaml.safe_load(open('codegen/generator_config.yaml')).get('resources') or {}).keys())
bad = 0
for n in names:
    pascal = ''.join(p.capitalize() for p in n.split('_'))
    r = subprocess.run([kgen, 'service', '--name', pascal, '--snakename', n],
                       capture_output=True, text=True)
    if r.returncode:
        bad += 1
        print('    FAILED', n, r.stderr.strip()[:200])
print(f'    scaffolded {len(names) - bad} package(s)')
sys.exit(1 if bad else 0)
PY
fi

pass() {
	echo "==> pass $1"
	"$KGEN" versions --sdk "$SDK_DIR" --config codegen/generator_config.yaml \
		--overrides codegen/config_overrides.yaml >/dev/null
	"$KGEN" schemas --config codegen/generator_config.yaml --spec spec/openapi3.json \
		--renames codegen/renames.yaml --schema-overrides codegen/schema_overrides.yaml >/dev/null
	if [[ "$1" == 1 ]]; then
		"$KGEN" crud --config codegen/generator_config.yaml \
			--config-overrides codegen/config_overrides.yaml --sdk "$SDK_DIR" \
			--archetypes codegen/crud_archetypes.yaml --test-values codegen/test_values.yaml \
			--version-support codegen/version_support.yaml --force >/dev/null
	fi
}

pass 1
pass 2
"$KGEN" import-manifest >/dev/null

echo "==> comparing against HEAD"
if [[ -n "$(git status --porcelain)" ]]; then
	echo
	echo "REGENERATION DID NOT REPRODUCE THE COMMITTED TREE:"
	git status --porcelain
	echo
	echo "Either a file is not actually generated, or a codegen input changed"
	echo "without the output being regenerated."
	exit 1
fi

echo "    reproduced exactly ($wiped files)"
echo "==> building"
go build ./...
echo "OK: the generated tree is fully reproducible from codegen/"
