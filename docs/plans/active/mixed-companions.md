# Mixed dogs and cats

Status: active. Task: `td-58936f` (also includes a parent-owned future-service plan).

## Assignment

Work in `/Users/marcus/code/avatars-mixed-companions`, branch `mixed-companions`. Read AGENTS.md and the style authoring guide. Add **Mixed** to Companions so a batch can contain dogs and cats. Use value `mixed` through discovery, studio, CLI `--animal mixed`, and HTTP `inputs.animal` / query. Keep default Dog and explicit Dog/Cat behavior, artwork, Pebble Random, and card/background/link behavior intact.

Use RecipeInputResolver to choose deterministically per final seed and persist concrete `dog` or `cat` on every avatar. Sample either species per avatar; no hybrids or promised exact batch balance. Direct, local, remote, and saved SVG/PNG must agree. Unsupported inputs and appearance overrides on saved exports remain refused.

Retain Mixed for consecutive generation and refresh. Selecting a saved avatar shows its concrete species; reopening a mixed collection suggests Mixed. Replace the current collection helper assumption that every mixed choice is called `random` with small metadata if needed. Avoid style-specific view branches or an elaborate schema. Update help/discovery, relevant docs, and the reusable guide/template with the convention: offer a mixed/random option for selectable appearance choices where meaningful, resolving per avatar and persisting concrete values. Keep presentation choices separate.

Own scoped implementation, tests, related docs, and this plan. Parent owns `docs/plans/active/requested-styles-service.md` and hosting/agent/auth research. Do not implement jobs, auth, plugins, or runners. Do not delegate, install, restart the live server, change Tailscale, or merge main. Use an independent TD_CONTEXT_ID. Commit and push coherent changes; parent reviews, lands, installs, creates a sample, and approves the shared task after the plan is complete.

## Proof

Check Mixed discovery, deterministic selection reaching both species, concrete persistence, fresh-process saved export and CLI/API parity, invalid requests, and unchanged explicit Dog/Cat output. Browser-check consecutive generation, refresh, mixed collection restoration, individual species metadata, and Pebble Random coexistence on desktop/mobile. Use isolated state, a free loopback port, and the CUA browser tools.

Run `make fmt-check vet test-race build`, required Node/reference tests, changed JavaScript syntax checks, `git diff --check`, and naturally for changed public prose. Record evidence, decisions, and limits here. Keep this plan active until parent delivery.

## Handoff

Pending.
