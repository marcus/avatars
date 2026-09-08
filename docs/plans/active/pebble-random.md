# Random palette colors for Pebble

Status: implementation and isolated browser proof complete; awaiting final card integration and checks. Task: `td-2734a4`.

## Assignment

Work in `/Users/marcus/code/avatars-pebble-random`, branch `pebble-random`. Read AGENTS.md and the style authoring guide. The user likes the Astra uplift and now requests a Random color option that generates using all existing palette colors randomly. Preserve the refined artwork and all 15 palette fills. Build on the existing typed color input and committed Companions foundation in this checkout.

Implement one discoverable `random` choice for Pebble through studio, CLI (`--color random`), and HTTP (`inputs.color` / stateless `color` query). Random selects from the full existing 15-color palette, independently for each generated avatar. Do not generate arbitrary hex colors or add an AI provider. Explicit colors and default Walnut retain their current behavior.

Resolve an actual palette color for each saved avatar and persist that concrete value in its recipe. Reopening/exporting must reproduce its color, and the inspector must display it. A fixed seed and Random request must remain deterministic across local, remote, saved, SVG, and PNG routes; a normal random-seed batch should mix colors. Use one shared core resolution path so the studio only passes the request. Keep validation before writes and saved appearance overrides refused. Expose Random through metadata, accurate help/instructions/capabilities, and docs. Other styles reject color input as before.

Retain studio draft behavior: Random should be usable for consecutive batches, ordinary refresh must preserve a manual choice, explicit navigation to a saved avatar should expose its concrete color. Mixed-color collection navigation may use a sensible generation default; document the choice. Preserve dogs/cats controls, inspector backgrounds, copied crop/dimension/background links, and the tactile card integration when you merge current main before handoff.

Own the required core/library/transport/studio input logic, focused tests, documentation, and this plan in your worktree. Do not edit Pebble geometry, other artwork, or card physics/sound. Do not delegate, merge to main, install, change Tailscale, or restart the live studio. Parent owns independent review and delivery. Use `TD_CONTEXT_ID=avatars-random-astra`. Commit coherent changes and push the private branch. This task is not a model comparison trial; ask parent for operational coordination if needed.

## Proof

Prove deterministic Random rendering across CLI and HTTP for both formats, explicit palette output stability, varied colors in a seeded batch, coverage of the palette over a reasonable fixed-seed sample, persisted concrete colors, fresh-process saved export, invalid requests refused before writes, and compatibility with animal inputs. Browser-check Random generation, mixed inspector colors, draft refresh, direct-link restoration, and desktop/mobile controls using isolated state and a free loopback port.

Run `make fmt-check vet test-race build`, `node --test internal/studio/*.test.mjs port/agent-portrait.test.ts`, syntax checks for changed JavaScript, `git diff --check`, and naturally on changed public prose. Record evidence and decisions here. Merge current main before final checks so the completed pet/card UI is retained. Hand off after pushing; parent will create a live mixed-color collection.

## Handoff

Implementation: Random is appended to Pebble discovery after the unchanged 15 palette entries. It has no swatch because it requests a palette selection rather than a single fill. `Engine.ResolveInputs` validates a batch request; `Engine.ResolveRecipeInputs` invokes an optional style resolver for each final avatar seed. Pebble hashes `pebble:color:` plus the UTF-8 seed with FNV-1a and indexes the complete existing palette. Geometry randomness and the refined artwork are untouched. The library persists each resolved concrete color before its single batch write; direct generator, engine, CLI, HTTP, and saved exports agree.

Studio decisions: successful generation preserves the draft, including Random for a one-avatar batch. Ordinary refresh preserves it too. Explicit avatar navigation restores the saved concrete color. Explicit navigation to a mixed-color collection selects Random; uniform collections restore their concrete color, including legacy Walnut defaults. Animal controls and copied shape/dimension/background state remain separate. There is no store migration and no new request field.

Validation so far: focused race checks for Pebble/core, library, and CLI pass; studio Node tests and changed JavaScript syntax checks pass. The fixed 256-seed core sample reaches all 15 colors; saved seeded batches contain concrete colors. Changed README, changelog, API guide, style-authoring guide, and generated instructions all pass gated `naturally scan`.

Independent review: `companions_astra` found no blockers in the merged implementation or documentation. Its real-process evidence at `/var/folders/9z/_hxsyhcx59d_cbrbhxfk9j000000gn/T/avatars-random-independent.bk28aqwi/result.json` records 30 unchanged explicit-color SVG/PNG comparisons, all 15 colors in a seeded 100-avatar batch, 40 Random/concrete/saved export comparisons, HTTP creation, local/HTTP parity, fresh server restart, refusal before writes, saved override refusal, and malicious/Unicode/empty seed safety.

Browser evidence: isolated library `/tmp/avatars-pebble-random-proof/library`, loopback port 60239. A 12-avatar Random batch followed by a one-avatar Random batch both retained Random. Manual refresh retained the choice; opening the mixed collection selected Random; opening two saved avatars restored Sage and Teal respectively. A copied Teal portrait link restored its concrete color, Light surround, circle crop, and 128 × 128 dimensions. Desktop and 390 × 844 screenshots were visually inspected. The mobile Random control and Generate button fit, and switching from Random to Cats generated a valid one-cat collection without stale color input. The proof tab viewport override was reset. Collections: `col_0993b372837b1775ec095d21bfd19bc3`, `col_15a6db305cf5593018ddc450e1fa38bc`, and `col_db83a18d3e1ddb3994812d89bb4f012c`.

Final integration and full required checks: pending the coordinator's card-grid landing. The Companions completion is merged through `da4c078`. Git inferred an active-to-implemented directory rename during that merge; the Random brief was deliberately kept active. No operational clarification was required. No live service, installed binary, or Tailscale route was changed.
