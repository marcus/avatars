# Tactile avatar preview cards

Status: ready for Sol implementation. Task: `td-988156`.

## Assignment and ownership

Work in `/Users/marcus/code/avatars-tactile-grid`, branch `tactile-grid`. Read `AGENTS.md` and this brief. Use a distinct `TD_CONTEXT_ID`, such as `avatars-grid-sol`. Implement this yourself, commit coherent slices, and push the private branch. The coordinator reviews, lands, installs, and restarts the live studio. Do not modify the two source repositories, other avatar worktrees, the installed binary, Tailscale, or the live server on port 7447.

Two independent style trials are running: Pebble introduces a Color dropdown and Companions introduces an Animal dropdown. Keep their work intact. Your scope is the preview grid, related presentation controls, and supporting modules/tests/docs under `internal/studio/`. The CLI, HTTP, storage, generation inputs, and image outputs remain unchanged.

Start by inspecting the reference implementations and preparing copied/adapted modules in new files. Before integrating with `app.js`, merge `main` once `docs/plans/implemented/pebble-style.md` exists there. If the dependency is not ready after the independent work, commit and report ready for integration; the coordinator can resume you with this same brief. Merge current `main` again before final handoff to incorporate any completed animal-style work. Resolve conflicts while retaining both style controls and this grid behavior.

## User outcome

Marcus requested the same card style, animations, sounds, and physics used in `~/code/marcuspictures` or `~/code/opentangle`, applied only to the avatar preview grid. The rest of the studio should continue to feel like its current compact creative app.

The preview area becomes a table of small, tactile avatar cards. Cards have matte stock, subtle grain, visible layered edges, quiet contact shadows, and a slight physical angle. The avatar remains the focus and retains its true colors and shape. Users can drag and gently toss cards, see them settle and nudge neighbors, and hear the familiar quiet paper movement. A click or keyboard activation still selects the portrait in the existing inspector.

## Reference code to reuse

Read the source repositories' `AGENTS.md` and design guides before adapting behavior. They are read-only references; do not start their full application stacks.

- `/Users/marcus/code/opentangle/DESIGN.md`, current source commit `78b0a86862770b695183f589f32b85d10b93cff9`.
- `/Users/marcus/code/opentangle/public/card-physics.js`: small DOM-free Matter adapter, bounded dragging, inertia, collisions, cancellation, and sleep.
- `/Users/marcus/code/opentangle/public/card-sound.js`: one procedural rubbing voice driven by current movement speed, gesture unlock, mute persistence, and visibility handling.
- `/Users/marcus/code/opentangle/public/card-table.js`: deal interpolation and pointer integration; adapt its assumptions about a full-page table to our scrolling preview panel.
- `/Users/marcus/code/opentangle/public/card-skin.css` and `cards.css`: inspect final cascade for the matte charcoal stock, 4-layer cut edges, restrained shadows, grain, and slight lift. Scope adopted rules to avatar cards and their preview surface.
- `/Users/marcus/code/opentangle/public/vendor/matter/`: engine and license. Its focused tests are in `/Users/marcus/code/opentangle/tests/card-physics.test.mjs`, `card-world.test.mjs`, and `card-sound.test.mjs`.
- `/Users/marcus/code/marcuspictures/docs/table-design.md`, source checkout at `7a5142d` when inspected. Its `web/static/js/table/` and `web/static/css/table/` modules extend the same system with more gallery-specific transitions. `tests/js/table-card-sound.test.js` covers timely shared sound and prevents delayed replay.

Prefer the smaller OpenTangle physics/audio modules as the base, consulting Marcus Pictures for useful refinements. Reuse their real constants and algorithms. Keep adapted source and vendor/license provenance in a short committed guide or README. Include only the dependencies this preview needs. The Matter adapter plus DOM transforms may be enough; a whole scene/rendering framework is optional only if needed to preserve the intended behavior. No CDN, frontend build pipeline, fonts, photos, or other source-site assets are needed.

## Interaction contract

- Keep a readable grid as the initial arrangement, with subtle deterministic per-card angles. Use the existing thumbnail-size control and collection navigation. Provide a compact, accessible Arrange/reset action to bring displaced cards back to their grid positions.
- Drag the whole card after a small movement threshold. Dragging must not select an avatar, trigger a link, or cause native image dragging/text selection. Ordinary click, tap, Enter, and Space keep selecting the portrait. Preserve focus indication and selected-state clarity. Do not add flipping or another detail view; the existing inspector is the card's action.
- Physics stays inside the preview's content bounds, away from the sidebar, toolbar, and inspector. Base bounds on layout, never transformed overflow. Preserve scrolling for long collections and on blank preview space. At 390-pixel mobile width, scrolling and tap selection remain comfortable; cancel a touch gesture cleanly when the browser takes over scrolling.
- Give movement the same deliberate, damped feel as the references. Animation frames stop when settled. Reset/cancel active drags safely during navigation, resize, pointer cancellation, visibility changes, and reduced-motion changes.
- Preserve keyed card elements and their physical positions during selection or an unchanged background library refresh. New agent-created avatars should appear without redealing the whole collection every five seconds. A real collection change or explicit Arrange may redeal.
- Add a compact sound toggle with an accessible name and pressed state. Reuse the quiet procedural paper voice and an Avatars-specific persisted preference key. Unlock audio only on a user gesture. Drive sound from visible current motion, share one voice for a batch, fade to silence when settled, and stop on mute/hidden page. Missing Web Audio or storage must not break the grid. No startup sound or late replay of expired motion.
- Honor `prefers-reduced-motion`: immediate layout, no inertia/deal flourish, and no motion sounds. Selection, accessible controls, and a useful grid remain available.
- Keep actual SVG/PNG pixels unmodified by the card treatment. Presentation texture belongs to stock and margins. Existing image loading/error recovery, crop/dimensions, copied links, inspector downloads, and direct navigation keep working.

This is presentation behavior; dragging and sound preferences do not change avatar records and do not need CLI/API counterparts.

## Implementation and proof

1. Inspect both source implementations. Adapt the relevant headless physics and sound modules and vendor files with provenance; retain or adapt their focused tests where behavior is reused.
2. After merging the Pebble foundation, connect a narrow card-grid controller to the studio. Keep domain operations and style inputs in the existing application/API path. Check any new embedded asset paths and MIME types.
3. Prove the real browser journey with an isolated data directory and unused loopback port: initial cards, selection, drag without selection, collisions and boundary containment, return to Arrange, resize, collection change, long-deck scroll, new CLI-created avatars, background refresh without a jump, keyboard selection, and inspector exports/share links. Inspect desktop and 390-pixel mobile screenshots.
4. Exercise sound unlock, mute persistence, settling, hidden-page silence, and reduced motion. Verify the audio lifecycle through focused tests and actual browser operation. Check there is no continuous idle animation loop or growing listener/observer population as collections change.
5. Run `make fmt-check vet test-race build`, all `internal/studio/*.test.mjs` tests, the TypeScript reference tests, JavaScript syntax checks, and `git diff --check`. Pass changed human-facing prose through `naturally`.
6. Commit and push. Update this plan with evidence, source provenance, decisions, and guidance gaps. The coordinator will independently review the integrated result, repair findings, and deliver a Tailscale link.

## Handoff

Source adaptation and decisions: pending.

Implementation and browser evidence: pending.

Validation and outstanding concerns: pending.

Independent review and live delivery: pending.
