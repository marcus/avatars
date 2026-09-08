# Companions: dogs and cats by a fresh Astra agent

Status: artwork implemented and visually checked; input integration waits for the Pebble foundation. Task: `td-98d7df`.

## Assignment and ownership

This committed brief is your complete assignment. Work in `/Users/marcus/code/avatars-companion-style`, branch `companion-style`. Read `AGENTS.md`, [Creating an avatar style](../../guides/active/creating-styles.md), and the API guide it links. Use your own `TD_CONTEXT_ID`, such as `avatars-companions-astra`. Do not delegate the implementation.

A Sol agent is independently adding Pebble and the first generation-input seam in another worktree. Preserve that trial: do not message or edit its worktree. Start with the artwork and its tests in your own new files. The shared input, CLI, HTTP, and studio changes depend on Pebble landing on `main`.

Commit and push your artwork slice, then inspect `main` for `InputGenerator` in `pkg/avatar/engine.go` and the completed Pebble plan in `docs/plans/implemented/`. Once both exist, merge `main` into this branch and complete integration against the actual code. If the dependency is not ready after your artwork proof, report the committed artwork and pause; the coordinator will resume you by pointing to this same brief after the dependency lands. This dependency signal is operational coordination, not extra design guidance.

Keep temporary preview servers on unused loopback ports with isolated libraries. Do not install over the live binary, alter Tailscale, merge into `main`, or restart the live studio on port 7447. The coordinator owns independent review, landing, installation, live samples, and final task approval. Commit coherent slices and push the private feature branch. Keep this plan active until final delivery.

## User outcome

The user selects **Companions**, chooses **Dogs** or **Cats** from an **Animal** dropdown, and generates a collection of playful pet portraits. Agents have the same choice through CLI and HTTP. Every saved avatar preserves its species and appearance through reopening, exports, and share links.

The user asked for roughly the detail and mood of *Calvin and Hobbes*, without requiring recognizable art from that comic. Use that direction for affectionate humor, curiosity, lively ink lines, and expressive poses. Create original dog and cat characters with their own visual language.

Style ID: `companions`. Display name: `Companions`. Default animal: `dog`.

## Visual contract

- Recognizable dogs and cats, framed as head-and-shoulders comic portraits. Choose native square or portrait dimensions based on the composition, and report them through the existing artwork and style metadata contracts.
- Organic ink contours, clear facial features, warm selective color, and a light paper-like background. Use a few controlled fur or motion strokes. Keep the visual density readable at 32 and 64 pixels and attractive at 256 pixels.
- Favor cheerful, curious, sleepy, proud, or gently mischievous expressions. Each character should feel alive and appealing. Avoid aggressive or distressed expressions.
- Dogs: vary ear silhouette, muzzle length/width, fur tufts, cheek shapes, spots or patches, gaze, and head tilt. Include both floppy and upright ears while keeping the animal clearly canine.
- Cats: clearly feline ears, cheek/fur silhouettes, small muzzle, whiskers, and varied markings. Vary coat color/pattern, gaze, cheek shape, and head tilt. Expressions can be sly or content without making every cat look angry.
- Use coherent seed-driven recipes. Meaningful silhouette and expression differences matter more than random decoration. Avoid a batch of nearly identical heads differentiated only by coat color.
- Keep characteristic ears, eyes, and muzzle visible in standard portrait and circular exports. Avoid geometry extending outside the artboard, floating features, accidental tangencies, and tiny clutter. Restrained collars or bandanas are optional; faces should carry the personality.
- No speech balloons, lettering, names, seed text, scenery, copied comic characters, external images, fonts, or online asset dependencies.

## Input and surface contract

Implement this after merging the completed Pebble foundation. Extend its actual narrow typed input seam. Do not create a second options framework or special studio endpoint.

- Generation input: `animal`, values `dog` and `cat`. UI labels: Dogs and Cats. Default: `dog`.
- Publish supported values and the default through the style metadata returned by `avatars styles --json` and `GET /api/v1/styles`. The studio uses those descriptors; the core owns defaults and refusal rules.
- Save the resolved value as `"inputs":{"animal":"dog"}` or `"inputs":{"animal":"cat"}` on each avatar recipe. Omitting the input for Companions resolves and persists `dog`. Existing input-free records and Pebble color recipes remain valid and unchanged.
- CLI: `avatars generate --style companions --animal cat --count 12 --json`; `avatars render --style companions --animal dog --seed sample --format png`. Support local and `--url` calls. `export` uses the saved input and refuses appearance overrides.
- HTTP create: `{"style":"companions","inputs":{"animal":"cat"},"count":12}`. Stateless route: `/api/v1/render?style=companions&seed=sample&animal=dog`. Reject unknown species, unknown nested fields, duplicate scalar query values, and animal inputs unsupported by another style. Companions does not accept Pebble's color input.
- Preserve `Generator` and the input-free Go rendering path. Defaults, validation, and rendering must not live only in transport or UI code. Avoid a core switch on style IDs.
- In the studio, show the compact Animal dropdown for Companions. Show Color for Pebble. Hide unsupported controls and exclude stale inputs from requests when switching styles. Selecting a saved avatar restores the correct species for generation and shows the saved choice in its metadata. A background refresh must preserve the user's unsubmitted selection.
- Selecting an animal affects the next batch. Existing saved portraits keep their appearance. Copied links continue to preserve portrait/circle mode and dimensions; the saved ID supplies the animal recipe.
- Update CLI help, instructions, capabilities, the API guide, README, changelog, and the reusable style guide where the second input exposes missing guidance.

## Sequence and acceptance

1. Build both animal compositions with deterministic native SVG output in new style-specific files. Use the existing exporter adapters for preview and tests. Keep this first slice independent of the unfinished input plumbing, then commit it.
2. Produce and visually inspect contact sheets from the real renderer: at least 24 dogs and 24 cats, labeled deterministic seeds, several 32/64/256-pixel examples, and circular exports. Inspect silhouettes, species recognition, expression, color, and feature placement. Commit a small reproducible proof script; keep generated sheets outside tracked source.
3. Merge the completed Pebble foundation from `main`, resolve local conflicts, and implement the full species choice through core, persistence, CLI, HTTP, and studio. Preserve existing styles and their SVG/PNG output contracts.
4. Test default and explicit species, invalid/unsupported input rejection before writes, deterministic local/HTTP parity, saved export and restart behavior, and compatibility with Pebble's color choices. Prove the actual CLI and browser journeys, including switching between animal/color/input-free styles, reopening a direct link, and exporting PNG/SVG. Check desktop and 390-pixel mobile layouts.
5. Run `make fmt-check vet test-race build`, `node --test internal/studio/*.test.mjs port/agent-portrait.test.ts`, and `git diff --check`. Pass changed human-facing prose through `naturally`. Commit and push the complete branch and record the handoff below.

Make routine choices independently and document guidance gaps you encounter. If a missing decision blocks useful progress, name it clearly. The trial should leave the next style author with better committed instructions.

## Handoff

Artwork and visual evidence: implemented in `pkg/avatar/companions.go`, with native 128 × 128 SVG. Six dog ear/head recipes and three cat cheek silhouettes use eight warm coat palettes, five marking recipes, relaxed/curious/happy/winking expressions, gaze, and head tilt. Default input-free `Companions.Generate` draws a dog; registration and typed animal inputs belong to the integration slice.

Run `scripts/prove-companions.sh /tmp/avatars-companions-astra-proof` to reproduce the artwork evidence. Visually inspected `dog-portrait-sheet.png`, `dog-circle-sheet.png`, `cat-portrait-sheet.png`, `cat-circle-sheet.png`, `dog-sizes.png`, and `cat-sizes.png` from the actual Go PNG exporter. Each contact sheet contains the same 24 labeled seeds (`companion-00` through `companion-23`); size sheets show seeds 00, 07, and 16 at actual 32/64/256 px in both shapes. HTML also loads the actual SVG exporter output. The species, expression, and silhouette remain legible at 32 px. Ear tips, whiskers, eyes, and muzzles retain clear circle margins. Quiet paper flecks and broad color areas avoid small-size clutter. The proof generator and shell entry point are committed; generated images stay in `/tmp`.

Focused checks: deterministic output, 100-seed variety per species, safe and well-formed SVG, cancellation, 24 PNG recipes per species, native size, and circle transparency. `go test -race ./pkg/avatar` and `git diff --check` passed.

Input integration and parity evidence: pending.

Guidance gaps and decisions: the brief provides enough guidance for the artwork slice. Native square framing leaves ears and whiskers inside the circle without style-specific export logic. Preview uses an opt-in Go test so the temporary pre-integration animal selector does not become a public API. Seed identity is deterministic but not promised to be collision-free in a finite visual recipe space. The artwork descriptions and proof labels passed `naturally scan`.

Independent review and live delivery: pending.
