#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
proof_dir="${1:-$(mktemp -d "${TMPDIR:-/tmp}/avatars-field-birds.XXXXXX")}"
mkdir -p "$proof_dir"
proof_dir="$(cd "$proof_dir" && pwd)"
AVATARS_BIRD_PROOF_DIR="$proof_dir" go test ./pkg/avatar -run '^TestFieldBirdsProof$' -count=1 -v
printf 'Contact sheet and size/circle evidence: %s\n' "$proof_dir"
