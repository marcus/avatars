#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
# shellcheck source=publish-homebrew-tap.sh
source "$repo_root/scripts/publish-homebrew-tap.sh"

fail() {
  echo "release publication test failed: $*" >&2
  exit 1
}

for valid in v0.1.0 v1.0.0 v12.34.56; do
  validate_release_version "$valid" || fail "rejected valid version $valid"
done
for invalid in 1.2.3 v1.2 v01.2.3 v1.02.3 v1.2.03 'v1.2.3;false'; do
  if validate_release_version "$invalid"; then
    fail "accepted invalid version $invalid"
  fi
done

[[ $(compare_versions v1.2.3 v1.2.3) == 0 ]] ||
  fail "equal version comparison"
[[ $(compare_versions v1.10.0 v1.9.9) == 1 ]] ||
  fail "newer version comparison"
[[ $(compare_versions v0.9.9 v1.0.0) == -1 ]] ||
  fail "older version comparison"

temporary=$(mktemp -d)
cleanup() {
  rm -rf "$temporary"
}
trap cleanup EXIT
mkdir "$temporary/current" "$temporary/expected"

sha=0000000000000000000000000000000000000000000000000000000000000000
"$repo_root/scripts/render-homebrew-formula.sh" \
  v1.2.3 "$sha" "$temporary/expected/avatars.rb" >/dev/null
check_formula_transition \
  "$temporary/current/missing.rb" "$temporary/expected/avatars.rb" v1.2.3 ||
  fail "rejected the first formula publication"
"$repo_root/scripts/render-homebrew-formula.sh" \
  v1.2.2 "$sha" "$temporary/current/avatars.rb" >/dev/null
check_formula_transition \
  "$temporary/current/avatars.rb" "$temporary/expected/avatars.rb" v1.2.3 ||
  fail "rejected an upgrade"

cp "$temporary/expected/avatars.rb" "$temporary/current/avatars.rb"
set +e
check_formula_transition \
  "$temporary/current/avatars.rb" "$temporary/expected/avatars.rb" v1.2.3
status=$?
set -e
[[ $status == 10 ]] || fail "exact formula was not idempotent"

sed -i.bak "s/$sha/1111111111111111111111111111111111111111111111111111111111111111/" \
  "$temporary/current/avatars.rb"
rm "$temporary/current/avatars.rb.bak"
if check_formula_transition \
  "$temporary/current/avatars.rb" "$temporary/expected/avatars.rb" v1.2.3 \
  >/dev/null 2>&1; then
  fail "accepted a same-version SHA change"
fi

"$repo_root/scripts/render-homebrew-formula.sh" \
  v1.2.4 "$sha" "$temporary/current/avatars.rb" >/dev/null
if check_formula_transition \
  "$temporary/current/avatars.rb" "$temporary/expected/avatars.rb" v1.2.3 \
  >/dev/null 2>&1; then
  fail "accepted a downgrade"
fi

sample_runs='[
  {
    "databaseId": 101,
    "event": "push",
    "headBranch": "v1.2.3",
    "headSha": "commit123",
    "status": "completed",
    "conclusion": "failure"
  },
  {
    "databaseId": 102,
    "event": "push",
    "headBranch": "v1.2.3",
    "headSha": "commit123",
    "status": "completed",
    "conclusion": "success"
  },
  {
    "databaseId": 103,
    "event": "pull_request",
    "headBranch": "v1.2.3",
    "headSha": "commit123",
    "status": "completed",
    "conclusion": "success"
  },
  {
    "databaseId": 104,
    "event": "push",
    "headBranch": "v1.2.3",
    "headSha": "othercommit",
    "status": "completed",
    "conclusion": "success"
  }
]'

selected=$(select_release_run "$sample_runs" v1.2.3 commit123)
[[ $(jq -r .databaseId <<<"$selected") == "102" ]] ||
  fail "did not select the latest exact-tag push run"
verify_release_run "$selected" v1.2.3 commit123 ||
  fail "rejected the successful exact-tag push run"

git init --bare --initial-branch=main "$temporary/tap.git" >/dev/null
git clone "$temporary/tap.git" "$temporary/seed" >/dev/null 2>&1
git -C "$temporary/seed" config user.name "Test Committer"
git -C "$temporary/seed" config user.email "test@example.com"
mkdir -p "$temporary/seed/Formula"
cp "$temporary/current/avatars.rb" "$temporary/seed/Formula/avatars.rb"
git -C "$temporary/seed" add Formula/avatars.rb
git -C "$temporary/seed" commit -m "initial tap formula" >/dev/null
git -C "$temporary/seed" push origin HEAD:main >/dev/null 2>&1

git clone "$temporary/tap.git" "$temporary/publisher" >/dev/null 2>&1
git -C "$temporary/publisher" config user.name "Publisher"
git -C "$temporary/publisher" config user.email "publisher@example.com"

git -C "$temporary/seed" commit --allow-empty -m "racing tap update" >/dev/null
git -C "$temporary/seed" push origin HEAD:main >/dev/null 2>&1

cp "$temporary/expected/avatars.rb" "$temporary/publisher/Formula/avatars.rb"
git -C "$temporary/publisher" add Formula/avatars.rb
git -C "$temporary/publisher" commit -m "avatars v1.2.3" >/dev/null

remote_formula="$temporary/remote-formula.rb"
cleanup_remote_formula() {
  rm -f "$remote_formula"
}
trap cleanup_remote_formula EXIT
push_tap_commit "$temporary/publisher" "$temporary/expected/avatars.rb"
git --git-dir="$temporary/tap.git" show main:Formula/avatars.rb \
  >"$remote_formula"
cmp -s "$temporary/remote-formula.rb" "$temporary/expected/avatars.rb" ||
  fail "racing tap update did not preserve the expected formula"

echo "release publication unit tests passed"
