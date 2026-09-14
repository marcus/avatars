#!/usr/bin/env bash
set -euo pipefail

# Refresh docs/images/logo.png from the project's identity logo in Ongoing.
#
# Ongoing stores the published logo as an impressions.logo.ref whose poster
# is a content-addressed PNG. This script resolves the current ref, fetches
# the poster, verifies its SHA-256, and records the revision alongside the
# image so a rerun is a no-op until the logo actually changes.
#
#   ./scripts/sync-logo.sh            # update if the logo changed
#   ./scripts/sync-logo.sh --check    # exit 1 if docs/images/logo.png is stale
#
# Environment:
#   AVATARS_ONGOING_PROJECT  Ongoing project slug (default: avatars)
#   ONGOING_URL              Ongoing web service origin (default: from `ongoing status`)

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
project=${AVATARS_ONGOING_PROJECT:-avatars}
image="$repo_root/docs/images/logo.png"
stamp="$repo_root/docs/images/logo.json"
check=false
[[ ${1:-} == --check ]] && check=true

die() {
  echo "Error: $*" >&2
  exit 1
}

command -v ongoing >/dev/null || die "ongoing CLI is not installed"

ref=$(ongoing get "$project" identity.logo 2>/dev/null) ||
  die "project '$project' has no identity.logo in Ongoing"

read -r revision poster_hash < <(printf '%s' "$ref" | python3 -c '
import json, sys
d = json.load(sys.stdin)
if d.get("kind") != "impressions.logo.ref":
    sys.exit("unexpected identity.logo kind: %s" % d.get("kind"))
print(d["revision"], d["poster"]["hash"])
') || die "could not parse identity.logo for '$project'"

[[ $poster_hash =~ ^[0-9a-f]{64}$ ]] || die "poster hash is not a SHA-256 digest: $poster_hash"

current_revision=""
if [[ -f $stamp ]]; then
  current_revision=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1])).get("revision",""))' "$stamp")
fi

if [[ -f $image && $current_revision == "$revision" ]]; then
  echo "logo is current (revision ${revision:0:12})"
  exit 0
fi

if $check; then
  echo "logo is stale: docs/images has ${current_revision:0:12}, Ongoing has ${revision:0:12}" >&2
  exit 1
fi

origin=${ONGOING_URL:-$(ongoing status 2>/dev/null | awk '$1 == "url" {print $2; exit}')}
[[ -n $origin ]] || die "cannot determine the Ongoing service URL; set ONGOING_URL"

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
curl -fsS -o "$tmp" "$origin/api/artifacts/$poster_hash" ||
  die "failed to fetch poster $poster_hash from $origin"

actual=$(shasum -a 256 "$tmp" | awk '{print $1}')
[[ $actual == "$poster_hash" ]] || die "poster digest mismatch: expected $poster_hash, got $actual"

mkdir -p "$(dirname "$image")"
mv "$tmp" "$image"
trap - EXIT
printf '{\n  "project": "%s",\n  "revision": "%s",\n  "poster": "%s",\n  "synced": "%s"\n}\n' \
  "$project" "$revision" "$poster_hash" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >"$stamp"
echo "updated docs/images/logo.png (revision ${revision:0:12})"
