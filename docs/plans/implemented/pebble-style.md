# Pebble: a fresh-agent style authoring trial

Status: implemented and independently reviewed; live on the private studio. Task: `td-bb42fd`.

## Assignment and ownership

You are the implementation agent. This committed document is your complete assignment; you receive no conversation history or additional design advice. Work in `/Users/marcus/code/avatars-pebble-style`, branch `pebble-style`. Read `AGENTS.md`, [Creating an avatar style](../../guides/active/creating-styles.md), and its linked API guide. Inspect the code and the reference image. Use an independent `TD_CONTEXT_ID` such as `avatars-pebble-sol` for task logs. Do not start another implementation agent; this trial evaluates one fresh agent.

Implement, test, inspect the art and studio, update the docs, commit coherent changes, and push this private feature branch. Do not merge to main, install over the active binary, alter Tailscale, or restart the live studio on port 7447. The coordinating agent owns independent review, landing, installation, live sample creation, and final task approval. Use an isolated temporary data directory and free loopback port for your proof. Keep code and proof self-contained in this checkout.

Make ordinary implementation decisions yourself. If guidance is insufficient, record the gap in the handoff below. Ask only if an unresolved issue prevents useful progress. Finish the full feature, including the input plumbing.

## Intended experience

The user selects **Pebble**, chooses a color from a compact dropdown, and generates a batch. Each avatar is a cute, pudgy, softly irregular round character with only two small eyes. Randomness gives the batch quiet differences in shape and gaze. Users and agents can reopen and export exactly the same colored character later.

Style ID: `pebble`. Display name: `Pebble`.

Reference: [user-provided screenshot](../../references/pebble-reference.png). The screenshot is the user's visual direction, committed unchanged for this implementation trial; it is not a runtime asset. It shows a squat warm-brown shape with two tiny light eye marks on a light background. Create original procedural geometry inspired by its simplicity. Keep the source attribution in `docs/references/README.md`.

## Visual contract

- A single solid-color body, round but slightly squashed or uneven, with generous breathing room on a transparent square canvas. Native artwork is 64 × 64.
- Exactly two eyes. Small ovals, dots, or short soft eye marks may vary subtly in width, height, spacing, gaze, and tilt. Keep both eyes visible and distinct at 24 and 32 pixels; at 16 pixels the silhouette must read and the eyes should not merge.
- No mouth, nose, limbs, ears, hair, clothing, eyebrows, blush, outlines, gradients, shadows, grain, or accessories. The silhouette and eye placement carry the character.
- Eye color may be warm ivory on darker colors and dark brown on light colors. Define the eye contrast in the shared palette/core.
- Seed-driven silhouette and eyes vary while color is fixed by the chosen input. Avoid extreme bulges, sharp corners, flat cutoffs, or expression changes that look angry or distressed.
- Both portrait and circle exports should retain the whole character; size the body with enough margin inside the shared circular mask.

Start with these 15 stable color values. Use these body fills so samples and downstream recipes are predictable:

| ID | Label | Body fill |
| --- | --- | --- |
| walnut | Walnut | `#92744F` |
| cocoa | Cocoa | `#655047` |
| clay | Clay | `#B66F56` |
| apricot | Apricot | `#E6AA78` |
| butter | Butter | `#E4CA78` |
| moss | Moss | `#71805A` |
| sage | Sage | `#A5B59A` |
| teal | Teal | `#4D8985` |
| sky | Sky | `#9DBFD1` |
| denim | Denim | `#607C9B` |
| lavender | Lavender | `#AAA0C6` |
| mauve | Mauve | `#A47B97` |
| rose | Rose | `#D6A0A4` |
| coral | Coral | `#D77F6D` |
| slate | Slate | `#6C777D` |

Default color is `walnut`. A batch shares the selected color; its silhouettes and eyes vary. No random-color option or custom hex input is needed.

## Shared color contract

This is the first style with user input. Add the smallest durable input seam; the code currently has only seed-based generators and export `Options`.

- The shared core owns color discovery, defaults, validation, and rendering. Keep the palette in one place, exposed through `styles --json` and `GET /api/v1/styles`. Give each color its stable value, readable label, and swatch hex, plus the default. A small typed color descriptor is sufficient; do not add a general form/schema framework.
- Use a typed generation input object with JSON shape `"inputs":{"color":"sage"}` on `CreateRequest` and each saved avatar recipe. Resolve and persist `walnut` when color is omitted for Pebble. Older records with no inputs remain valid and render identically. Existing styles must reject nonempty color input.
- Preserve the existing input-free Go generator/engine usage. Add a narrow optional adapter or input-aware method; exporters remain concerned with framing and encoding. Document the implemented Go extension path with an example.
- `avatars generate --style pebble --color sage --count 12 --json` saves Sage avatars. `avatars render --style pebble --color sage --seed sample --format png` renders without saving. Both local and `--url` calls must work. Omitting `--color` selects Walnut for Pebble. Other styles continue their current behavior when it is omitted.
- HTTP collection creation accepts `{"style":"pebble","inputs":{"color":"sage"},"count":12}`. Stateless rendering accepts `/api/v1/render?style=pebble&seed=sample&color=sage`. Unknown nested input fields, invalid colors, duplicate scalar query values, or color supplied for another style are invalid requests. Validate before saving anything.
- Saved export routes and `avatars export` use the stored input and refuse appearance overrides. Existing framing options still work. A saved avatar link does not need the color in the URL: its ID resolves the recipe. Copied links must continue to preserve portrait/circle mode and dimensions.
- Keep CLI help, `instructions`, `capabilities`, API documentation, README style list, and changelog accurate. Remove the blanket claim that the studio has no appearance steering. Do not claim all native artwork is 64 × 72 after adding this square style.

## Studio behavior

Show one compact **Color** dropdown when the selected generator supports it. Populate it from API metadata, with readable labels and a nearby swatch if useful. Keep the control hidden for other styles and do not leak a stale color into their requests. Defaults should be evident.

Changing the dropdown affects the next batch only. Selecting or opening a saved Pebble avatar should show its saved color in the inspector and initialize the generation control to that value. Opening a Pebble collection may use its first avatar's saved color; batches are uniform. A background library refresh must not overwrite an in-progress manual color choice. Existing navigation, crop, sizing, downloads, and copied links remain intact. Check the toolbar and inspector at 390-pixel mobile width.

## Work sequence and acceptance

1. Implement one colored recipe through core, stateless render, persistence, and saved export. Cover default resolution, rejection before writes, old records, and input-free Go compatibility.
2. Connect CLI and HTTP, update discoverability, and prove local/HTTP SVG and PNG parity for a fixed seed and explicit color. Exercise actual commands with isolated state.
3. Implement all palette values and the studio control. Prove generation, direct-link restoration, saved-color metadata, refresh behavior, style switching, and downloads in a real browser.
4. Produce and visually inspect a reproducible contact sheet: all 15 colors with the same seed, at least 12 Walnut seeds showing shape/eye variety, and 16/24/32/64/256-pixel samples plus circle framing. Use the real renderer. Keep a small recipe/script in the repo, generated artifacts outside tracked source, and report proof paths.
5. Run `make fmt-check vet test-race build`, `node --test internal/studio/*.test.mjs port/agent-portrait.test.ts`, and `git diff --check`. Pass changed human-facing prose through `naturally`. Commit and push. Update the reusable guide and template with lessons that help the next style author.

Independent review checks this acceptance contract; live sample collections follow review and installation. Keep this plan active until that is complete.

## Handoff

Implementation: complete on branch `pebble-style`. Pebble now renders original deterministic 64 × 64 SVG geometry with one softly irregular body, exactly two eyes, and all 15 specified colors. A typed core input descriptor and optional input-aware generator seam provide discovery, defaulting, validation, rendering, saved recipes, CLI/HTTP access, and input-free Go compatibility. The studio shows the API-provided Color control only for Pebble, sends resolved recipes, restores saved colors, preserves manual choices across refresh and style switching, and displays saved color in the inspector. README, changelog, agent discovery, API documentation, this authoring guide, and the brief template are updated. `scripts/pebble-contact-sheet.sh` reproduces the visual proof.

Evidence: focused and full automated checks pass. The final required commands were `make fmt-check vet test-race build`, `node --test internal/studio/*.test.mjs port/agent-portrait.test.ts`, and `git diff --check`. Gated `naturally scan` runs passed for every changed human-facing Markdown file. Local and HTTP render commands produced identical fixed-seed Sage SVG and PNG files; SHA-256 was `2a6e9ebf4e98dcfbdf7404a840e6d6c8adef4ad46d914ad4a52dfffce4f311ff` for SVG and `cd29ca7a44e31b73937d38bb5b595a472067d58f9f9793124dfaca280281be99` for PNG.

Visual proof is at `/tmp/avatars-pebble-proof.woh8vU/index.html`, with derived inspection sheets at `palette-sheet.png` and `variety-sheet.png`. It contains all 15 colors for `palette-proof`, 12 Walnut seeds, 16/24/32/64/256 pixel exports, and a 256-pixel circle export. Inspection found good light/dark eye contrast, distinct eyes through 24 pixels and still readable at 16 pixels, comfortable canvas and circle margins, and quiet silhouette/gaze variation without extra features. Real browser proof used an isolated library at `/tmp/avatars-pebble-studio.oLYHD1` and a free loopback port. It covered default Walnut discovery, manual Sage preservation after refresh and style switching, Coral restoration from a saved collection and avatar, studio generation, inspector metadata, SVG/PNG downloads, direct-link circle and dimension restoration, and desktop plus 390-pixel layouts. The first narrow screenshot exposed a clipped Generate button; the responsive grid was corrected and visually rechecked at 390 pixels.

Guidance gaps and decisions made without extra instructions: no product clarification was needed. The task was already `in_progress` under the coordinator's td session, so the independent `avatars-pebble-sol` context could not start it; that context was still used successfully for implementation logs. Eye colors are stored beside body fills in the shared core palette but omitted from public discovery because callers need only the selectable value, label, and swatch. An empty color value resolves like omission; any nonempty unknown value is rejected.

Independent review and live delivery: the coordinator reviewed the code, contact sheets, and real browser journeys. Independent CLI/HTTP checks covered all 15 palettes in SVG and PNG, persistence, invalid requests before writes, defaults, and 36 unchanged exports from the original styles. Desktop and 390-pixel browser checks passed for generation, saved-color restoration, manual choices surviving refresh, and circle/dimension links. The implementation was merged to main and installed at commit `63ccbbe`; the private Tailscale studio serves the color metadata and a generated Walnut collection. The initial Sol-only documentation trial is complete without extra product guidance. The user subsequently requested more personality; that separate Astra refinement is tracked in `pebble-uplift.md` and may revise this pre-release artwork while retaining its input contract.
