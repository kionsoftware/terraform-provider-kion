#!/usr/bin/env bash
# Fail if anything internal has crept into the tree.
#
# This repository is public and accepts outside contributions. Internal
# references, such as backend source paths, internal hostnames, internal
# repo/group names, and AI planning notes, are not secrets, but they leak
# implementation detail nobody outside can act on, and once published they live
# in every fork. A reviewer cannot reliably catch them by eye, so this runs in
# CI.
#
# A denylist of internal names is itself an internal reference: committing the
# literal strings publishes the index of what we are trying not to publish, and
# this script has to exclude itself from its own grep to avoid matching. So the
# committed list holds only patterns that name nothing in particular, and the
# specific literals are supplied at runtime:
#
#   INTERNAL_REF_PATTERNS       newline-separated extended regexes
#   INTERNAL_REF_PATTERNS_FILE  path to a file of the same, one per line,
#                               '#' comments and blank lines ignored
#                               (default: scripts/.internal-refs-private,
#                                which .gitignore keeps untracked)
#
# In CI, set INTERNAL_REF_PATTERNS from a repository secret. Locally, keep the
# private file. With neither, the committed patterns still run and the script
# says the private set was not loaded, so a silently weakened check is visible
# rather than assumed.
#
# Adding a term is cheap. Removing one should be deliberate.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

# Extended regexes, matched against TRACKED files only. Untracked working files
# (local notes, scratch dirs) are the contributor's business.
#
# Everything here is deliberately non-specific: the company domain is public, and
# the source-layout patterns describe a shape rather than naming a repository.
patterns=(
  '[a-z0-9-]+\.kion\.io'         # any internal host; use kion.example.com in docs
  'src/(app|domain)/'            # backend source layout
  '[[:space:]]portal[[:space:]]' # prose naming the backend
  'docs/superpowers'             # AI planning notes (gitignored; catch re-adds)
)

private_loaded=0
private_file="${INTERNAL_REF_PATTERNS_FILE:-scripts/.internal-refs-private}"
if [ -n "${INTERNAL_REF_PATTERNS:-}" ]; then
  while IFS= read -r line; do
    [ -n "$line" ] && patterns+=("$line")
  done <<<"$INTERNAL_REF_PATTERNS"
  private_loaded=1
elif [ -f "$private_file" ]; then
  while IFS= read -r line; do
    case "$line" in ''|\#*) continue ;; esac
    patterns+=("$line")
  done <"$private_file"
  private_loaded=1
fi

fail=0
for p in "${patterns[@]}"; do
  # `git grep -I` skips binaries. Exclude this script: the committed patterns
  # are generic, but a runtime-supplied one could match its own documentation.
  # .gitignore legitimately names docs/superpowers; that rule is what keeps the
  # notes out. Exclude it for that pattern only, not for every pattern.
  extra=()
  if [ "$p" = 'docs/superpowers' ]; then extra=(':!.gitignore'); fi
  if hits=$(git grep -nIE "$p" -- . ':!scripts/check-no-internal-refs.sh' "${extra[@]}" 2>/dev/null); then
    echo "✗ internal reference matched /$p/:" >&2
    printf '%s\n' "$hits" | head -20 | sed 's/^/    /' >&2
    fail=1
  fi
done

# The greps above match file CONTENTS. A planning doc re-added under
# docs/superpowers/ contains no such text, so it would sail through. Check the
# tracked path list separately. Additional forbidden path regexes may be
# supplied privately the same way, via INTERNAL_REF_PATHS.
forbidden_paths='^docs/superpowers/|^\.gitlab-ci\.yml$'
if [ -n "${INTERNAL_REF_PATHS:-}" ]; then
  forbidden_paths="$forbidden_paths|$INTERNAL_REF_PATHS"
fi
if paths=$(git ls-files | grep -E "$forbidden_paths"); then
  echo "✗ internal files are tracked:" >&2
  printf '%s\n' "$paths" | sed 's/^/    /' >&2
  fail=1
fi

# Commit MESSAGES are published too, and nothing above looks at them: the
# greps match tracked file contents and paths. That gap let 24 lines of backend
# source paths, internal environment names and production entity counts reach a
# public history, where the only remedy is rewriting published commits. Checked
# here so it fails before the push instead.
#
# Scans the commits being pushed. INTERNAL_REF_COMMIT_RANGE overrides the range;
# an unset upstream (a brand-new branch) falls back to the default branch.
if [ "${INTERNAL_REF_SKIP_COMMITS:-0}" != "1" ]; then
  range="${INTERNAL_REF_COMMIT_RANGE:-}"
  if [ -z "$range" ]; then
    base=$(git rev-parse --verify --quiet origin/main || git rev-parse --verify --quiet main || true)
    [ -n "$base" ] && range="$base..HEAD"
  fi
  if [ -n "$range" ] && msgs=$(git log --format='%H %s%n%b' "$range" 2>/dev/null) && [ -n "$msgs" ]; then
    for p in "${patterns[@]}"; do
      if hits=$(printf '%s\n' "$msgs" | grep -nE "$p"); then
        echo "✗ internal reference in a commit message, matched /$p/:" >&2
        printf '%s\n' "$hits" | head -10 | sed 's/^/    /' >&2
        echo "    (amend or reword before pushing: a published message cannot be" >&2
        echo "     fixed without rewriting history)" >&2
        fail=1
      fi
    done
  fi
fi

if [ "$fail" -ne 0 ]; then
  cat >&2 <<'EOF'

This repo is public. Remove the reference, or describe the behaviour without
citing internal sources: "the API returns X" rather than a backend file path.
If a term is now legitimately public, drop it from the pattern set in the same
change, so the reason is reviewable.
EOF
  exit 1
fi

if [ "$private_loaded" -eq 1 ]; then
  echo "✓ no internal references in tracked files"
else
  echo "✓ no internal references in tracked files (public patterns only;" \
       "private set not configured, see the header of $0)"
fi
