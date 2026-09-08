# Mixed dogs and cats

Status: implementation complete, ready for parent review and delivery. Task: `td-58936f` (also includes a parent-owned future-service plan).

## Assignment

Work in `/Users/marcus/code/avatars-mixed-companions`, branch `mixed-companions`. Read AGENTS.md and the style authoring guide. Add **Mixed** to Companions so a batch can contain dogs and cats. Use value `mixed` through discovery, studio, CLI `--animal mixed`, and HTTP `inputs.animal` / query. Keep default Dog and explicit Dog/Cat behavior, artwork, Pebble Random, and card/background/link behavior intact.

Use RecipeInputResolver to choose deterministically per final seed and persist concrete `dog` or `cat` on every avatar. Sample either species per avatar; no hybrids or promised exact batch balance. Direct, local, remote, and saved SVG/PNG must agree. Unsupported inputs and appearance overrides on saved exports remain refused.

Retain Mixed for consecutive generation and refresh. Selecting a saved avatar shows its concrete species; reopening a mixed collection suggests Mixed. Replace the current collection helper assumption that every mixed choice is called `random` with small metadata if needed. Avoid style-specific view branches or an elaborate schema. Update help/discovery, relevant docs, and the reusable guide/template with the convention: offer a mixed/random option for selectable appearance choices where meaningful, resolving per avatar and persisting concrete values. Keep presentation choices separate.

Own scoped implementation, tests, related docs, and this plan. Parent owns `docs/plans/active/requested-styles-service.md` and hosting/agent/auth research. Do not implement jobs, auth, plugins, or runners. Do not delegate, install, restart the live server, change Tailscale, or merge main. Use an independent TD_CONTEXT_ID. Commit and push coherent changes; parent reviews, lands, installs, creates a sample, and approves the shared task after the plan is complete.

## Proof

Check Mixed discovery, deterministic selection reaching both species, concrete persistence, fresh-process saved export and CLI/API parity, invalid requests, and unchanged explicit Dog/Cat output. Browser-check consecutive generation, refresh, mixed collection restoration, individual species metadata, and Pebble Random coexistence on desktop/mobile. Use isolated state, a free loopback port, and the CUA browser tools.

Run `make fmt-check vet test-race build`, required Node/reference tests, changed JavaScript syntax checks, `git diff --check`, and naturally for changed public prose. Record evidence, decisions, and limits here. Keep this plan active until parent delivery.

## Handoff

Implementation commit: `b097019`. Mixed is available through discovery, the studio, local and remote CLI, direct generation, and HTTP. Companions resolves `mixed` through `RecipeInputResolver` using the UTF-8 final seed and an independent FNV-1a hash domain (`companions:animal:`), then saves only `dog` or `cat`. Existing artwork randomness and explicit/default recipes are unchanged. Both appearance descriptors now expose optional `mixed_value`; collection restoration reads that metadata for Companions Mixed and Pebble Random without style-specific view logic. The authoring guide and brief template carry this convention forward.

Validation passed:

- `make fmt-check vet test-race build`, all 30 required Node/reference tests, both changed JavaScript syntax checks, and `git diff --check`. Log: `/tmp/avatars-mixed-final-checks.log`.
- Real CLI/HTTP processes: 48 explicit Dog/Cat SVG/PNG exports unchanged against the pre-change binary; Pebble Walnut/Sage/Random PNG unchanged; both species persisted concretely; 24 local/HTTP/saved export comparisons; direct HTTP Mixed collection creation; identical saved export after process restart; invalid inputs refused before writes and saved appearance overrides refused.
- CUA desktop and 390-by-844 mobile proof in owned Chrome tab `727446955`: Mixed persisted after a 12-avatar batch, consecutive single-avatar generation, and refresh; mixed collection navigation restored Mixed; individual dog and cat navigation restored concrete species. A copied cat link retained Circle, 128-by-128 dimensions, and Light. Mobile controls fit without horizontal overflow. Pebble Random and Mixed generated valid concrete recipes when switching between them. Screenshots were visually inspected through CUA; the browser reported no console errors.
- Browser-created Mixed batches contained 8 dogs/4 cats, one cat, and 1 dog/3 cats. This demonstrates independent selection rather than any promised batch balance.
- Naturally gated scans passed for README, changelog, API guide, authoring guide, brief template, and generated instructions (scores 97–100; no hard-gate findings).

Evidence: `/var/folders/9z/_hxsyhcx59d_cbrbhxfk9j000000gn/T/avatars-mixed-proof.ccr2q62b/`, including `runtime-proof.json`, `browser-proof.json`, and the isolated saved library. The owned proof service on port 65501 was stopped and its browser tab closed; the mobile viewport override was reset. Live port 7447, installed binaries, main, and Tailscale were not changed.

Parent still owns independent review, landing/install/live samples, the future-service plan, and approval of shared task `td-58936f`. Keep this brief active until that delivery.

## Decisions and limits

- Mixed is an input policy, not saved appearance or a new artwork style. It offers either existing species per avatar, without hybrids or guaranteed batch balance.
- `mixed_value` is optional discovery metadata. A collection without it keeps the existing first-concrete-value fallback; presentation choices remain separate.
