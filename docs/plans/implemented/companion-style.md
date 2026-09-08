# Companions: dogs and cats by a fresh Astra agent

Status: implementation and browser proof complete; ready for coordinator review and live delivery. Task: `td-98d7df`.

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

Artwork and visual evidence: implemented in `pkg/avatar/companions.go`, with native 128 × 128 SVG. Six dog ear/head recipes and three cat cheek silhouettes use eight warm coat palettes, five marking recipes, relaxed/curious/happy/winking expressions, gaze, and head tilt. The registered `Companions` generator implements `InputGenerator`; input-free `Generate` draws a dog.

Run `scripts/prove-companions.sh /tmp/avatars-companions-astra-proof` to reproduce the artwork evidence. Visually inspected `dog-portrait-sheet.png`, `dog-circle-sheet.png`, `cat-portrait-sheet.png`, `cat-circle-sheet.png`, `dog-sizes.png`, and `cat-sizes.png` from the actual Go PNG exporter. Each contact sheet contains the same 24 labeled seeds (`companion-00` through `companion-23`); size sheets show seeds 00, 07, and 16 at actual 32/64/256 px in both shapes. HTML also loads the actual SVG exporter output. The species, expression, and silhouette remain legible at 32 px. Ear tips, whiskers, eyes, and muzzles retain clear circle margins. Quiet paper flecks and broad color areas avoid small-size clutter. The proof generator and shell entry point are committed; generated images stay in `/tmp`.

Focused checks: deterministic output, 100-seed variety per species, safe and well-formed SVG, cancellation, 24 PNG recipes per species, native size, and circle transparency. `go test -race ./pkg/avatar` and `git diff --check` passed.

Input integration and parity evidence: complete. Merged the Pebble foundation from `main`, then merged the reviewed Pebble artwork uplift and Dark/Light/Gray inspector backgrounds. The shared typed input now includes `Animal`, with dog/cat choices published by Companions metadata. Shared defaulting and refusal cover both fields, including direct generator calls. Library persistence uses the existing recipe pointer; CLI and HTTP forward the typed inputs. The studio reads enum choices from API descriptors, remembers unsubmitted choices by style, filters unsupported fields, and restores saved recipes on navigation.

Committed tests cover default and explicit species, direct/engine rendering agreement, invalid cross-style combinations, HTTP parsing and duplicate scalar rejection, defaults persisted on each avatar, SVG/PNG saved/stateless parity, local/remote CLI behavior, restart persistence, and native-size metadata. Existing Pebble and input-free fixtures pass.

Real-process proof: `/var/folders/9z/_hxsyhcx59d_cbrbhxfk9j000000gn/T/avatars-companions-integration.l9nk8l5c/runtime-proof.json`. It checks default/dog/cat local and HTTP generation, native and circular SVG/PNG byte parity through local render, HTTP render, local saved export, and remote saved export; it terminates and restarts its own isolated service and rechecks the saved cat. Invalid requests leave the collection count unchanged. Pebble still persists Sage.

Browser proof used a separate headless Chrome process and an isolated service on loopback port 57778. No existing user tab or live service was touched. `browser-proof.json` in the same evidence directory records successful Cat generation, Color/Animal/input-free switching, saved Cat restoration, a manual Dog selection surviving both explicit and automatic refresh, and a copied Cat link retaining circle shape, 192 × 192 dimensions, and Gray background. Downloaded `browser-cat.svg` and `browser-cat.png` have the requested frame. Visually inspected `desktop-cats-grid.png`, `desktop-cat-inspector.png`, `mobile-cat-inspector.png`, `mobile-animal-toolbar.png`, `mobile-color-toolbar.png`, `mobile-input-free-toolbar.png`, `mobile-dogs-grid.png`, and `mobile-dog-export.png`. The mobile viewport is 390 × 844; controls fit without horizontal overflow, and the inspector scrolls to export controls. Both dog and cat batches were created through the real UI. The proof server was stopped after verification.

Final checks passed: `make fmt-check vet test-race build`, `node --test internal/studio/*.test.mjs port/agent-portrait.test.ts` (12 tests), and `git diff --check`. README, changelog, API guide, style guide, and generated instructions passed `naturally scan` with scores 97–100. The coordinator independently reported old-style byte parity (36 exports), all 15 Pebble colors through local/HTTP/saved SVG+PNG, companion dog/cat parity, and invalid-input refusal; its evidence is `/var/folders/9z/_hxsyhcx59d_cbrbhxfk9j000000gn/T/avatars-final-review-2aexayio/result.json`.

Guidance gaps and decisions: the brief provides enough guidance for the artwork slice. Native square framing leaves ears and whiskers inside the circle without style-specific export logic. Preview uses an opt-in Go test so the temporary pre-integration animal selector does not become a public API. Seed identity is deterministic but not promised to be collision-free in a finite visual recipe space. The artwork descriptions and proof labels passed `naturally scan`.

The merged foundation had no native-size fields on style discovery, despite the brief referring to a style metadata contract. Added `native_width`/`native_height` for built-ins and documented the fields. This changes discovery metadata without changing existing artwork. Browser proof found that first grid selection carried the old 8:9 control defaults into a square style; first selection now derives defaults from the target style, while explicit shared-link dimensions still win. The reusable guide now calls out validating all typed inputs before return, using the same validator for direct and engine calls, filtering stale studio controls, and checking both supported and hidden-control mobile layouts. No additional design instruction was needed.

Independent artwork review: the coordinating agent inspected the actual 24-dog and 24-cat portrait sheets, the cat circle sheet, and the dog size sheet on September 7, 2026. Species, silhouette variety, expressions, and small-size readability meet the brief; circle margins preserve defining features. Read the generator and focused tests with no blocking artwork finding. Input integration and implementer browser proof are now complete. The coordinator owns independent integrated UI review, landing, installation, live samples, and final approval.

Live delivery: merged and installed from main `31f3412`. The coordinator independently inspected the live dog and cat collections, species selection, and Cat inspector with Circle and Gray; no browser console warnings or errors were reported. Sample collections are `col_c84bd979c73c5e2c909d01971e4a0236` (Good dogs) and `col_29deee27f78ff7de5553f579a4b55078` (Curious cats), served through the verified private Tailscale studio. The core/input review and real CLI/API/export proof above passed.
