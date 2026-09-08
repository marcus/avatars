#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
proof_dir="${1:-$(mktemp -d "${TMPDIR:-/tmp}/avatars-companions.XXXXXX")}"
mkdir -p "$proof_dir"
proof_dir="$(cd "$proof_dir" && pwd)"
AVATARS_COMPANION_PROOF_DIR="$proof_dir" go test ./pkg/avatar -run '^TestCompanionProof$' -count=1 -v
printf 'Contact sheets and HTML: %s\n' "$proof_dir"
