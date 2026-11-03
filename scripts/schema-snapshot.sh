#!/usr/bin/env bash
# Write `terraform providers schema -json` for the provider this tree builds to
# $1. Used by `make schema-snapshot` and `make schema-snapshot-check`.
set -euo pipefail
out=${1:?usage: schema-snapshot.sh <output.json>}
case $out in /*) ;; *) out=$PWD/$out ;; esac

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/bin" "$work/cfg"
go build -o "$work/bin/terraform-provider-kion" .
cat >"$work/dev.tfrc" <<TFRC
provider_installation {
  dev_overrides { "kionsoftware/kion" = "$work/bin" }
  direct {}
}
TFRC
cat >"$work/cfg/main.tf" <<'TF'
terraform {
  required_providers {
    kion = { source = "kionsoftware/kion" }
  }
}
TF
# dev_overrides needs no init; the warning it prints goes to stderr.
(cd "$work/cfg" && TF_CLI_CONFIG_FILE="$work/dev.tfrc" terraform providers schema -json 2>/dev/null) >"$out"
