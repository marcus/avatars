# Random palette colors for Pebble

Status: active. Task: `td-2734a4`.

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

Pending.
