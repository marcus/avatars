# Give Pebble more personality

Status: Astra implementation complete; awaiting coordinator review. Task: `td-c3c8ed`.

## Assignment

Work in `/Users/marcus/code/avatars-pebble-uplift`, branch `pebble-uplift`. Read `AGENTS.md`, the [style authoring guide](../../guides/active/creating-styles.md), the existing Pebble brief, and `pkg/avatar/pebble.go`. This is an uplift of Sol's committed implementation at `63ccbbe`, not a replacement project. Inspect the [original reference](../../references/pebble-reference.png) and render the existing implementation before editing.

The user finds Pebble acceptable but lacking personality and explicitly requested an Astra refinement of the existing work. Preserve the tiny pudgy character with only eyes, all 15 color values and fills, default Walnut, native 64 × 64 canvas, and the color/persistence/CLI/API contract. The user has authorized this pre-release visual revision under the existing `pebble` ID. Describe that decision in the handoff; existing input recipes remain valid. Do not change another style's output.

Own only Pebble geometry and focused tests, its reproducible art-proof script, and this plan. Add a short style-authoring lesson if the visual review exposes a useful one. Do not edit shared engine/input types, library, CLI, HTTP, studio, other style files, or the original reference image. The coordinator is separately adding Dark/Light/Gray inspector backgrounds, another Sol agent owns the card grid, and another Astra agent owns dogs/cats. Do not delegate, install, merge main, alter Tailscale, or touch the live server on port 7447. Commit coherent changes and push the private branch. Use `TD_CONTEXT_ID=avatars-pebble-uplift-astra` for progress and handoff.

## Visual direction

Build on the current soft body and two-eye composition. Give seeds meaningful character differences that survive at icon size: squat versus tall-but-pudgy silhouettes, gentle leaning, rounded cheeks or lopsided mass, and coherent eye placement. Keep the shape appealing and comfortably inside the circular crop.

Eye width, height, spacing, gaze, asymmetry, and angle can suggest curiosity, sleepiness, shyness, contentment, or mild mischief. Simple soft squints or a wink are welcome when both eye marks still read. Use the whole face arrangement to communicate mood. Avoid making every avatar the same face with a slightly different outline, and avoid random combinations that look angry, broken, or distressed.

Maintain exactly two eye features and a single plain body. No mouth, nose, eyebrows, blush, limbs, hair, accessories, texture, outlines, shadows, or gradients. Eye marks may use the simplest SVG geometry needed; adjust implementation-shaped tests when they unnecessarily prevent an intended expression. Keep high-signal checks for safety, deterministic input/output, two-eye simplicity, valid exports, and stable color semantics.

## Proof and handoff

1. Capture a before sheet from the existing code, then refine its geometry and expressions. Compare the same labeled seeds before and after, including `walnut-proof:0` through `walnut-proof:11` and a larger 24-seed cast. Keep the original source available through git.
2. Visually inspect actual renderer output for all 15 colors and at 16/24/32/64/256 pixels, plus circle exports. Check eye contrast and feature separation on dark, light, and gray surrounds. Do not generate a mockup as evidence for the procedural renderer.
3. Run focused Pebble/core race tests, the unchanged TypeScript fixture checks, and a fixed-seed explicit-color CLI render/export check. Run `make fmt-check vet test-race build`, required Node tests, and `git diff --check` before the final handoff. Pass changed public prose through `naturally`.
4. Commit and push. Record what improved, the visual recipe choices, tests, proof paths, and any concerns here. The coordinator independently reviews and lands the result, then creates the live sample collection and finishes task approval.

## Handoff

Implementation: the original Pebble now varies among five pudgy proportion families with smooth unequal cheeks, a broad rounded base, and gentle lean. Seven coordinated expression recipes use quiet horizontal eyes, curiosity, sleepiness, a lower close-set shy face, contented arches, a wink, or open eyes. Eyes stay paired in placement and gesture. The renderer still emits one plain filled body and exactly two filled eye features on its native 64 × 64 canvas.

Compatibility decision: the user explicitly authorized this pre-release visual revision under the existing `pebble` ID. Existing seed/color recipes remain valid and recreate the refined art; their earlier geometry is intentionally revised. The 15 color values, labels, body and eye fills, default Walnut, persistence, CLI, and HTTP contracts are unchanged. No other generator or shared implementation was edited. Sol's original source remains at `63ccbbe:pkg/avatar/pebble.go`.

Before/after evidence: `/tmp/avatars-pebble-uplift-proof/comparison.png` compares the original 12 labeled seeds; `comparison-24.png` compares `walnut-proof:0` through `walnut-proof:23`. `before/` and `after/` contain complete real PNG exports and `index.html`, `light.html`, `dark.html`, and `gray.html`. The original binary was captured at `before/avatars-baseline` before editing. `cast-{light,dark,gray}.png` and `palette-{light,dark,gray}.png` are derived inspection sheets. All six sheets plus both comparisons were visually inspected. No generated mockup was used.

Reproduction: build the desired revision with `make build`, then run `scripts/pebble-contact-sheet.sh NEW_OUTPUT_DIR`. Set `AVATARS_BIN` to a baseline binary to render the same seeds before the change. The script now captures 24 Walnut seeds, native-size 16/24/32/64 previews, circle framing, and all 15 colors at 16/24/32/64/256 pixels in both square and circle exports, on three surrounds.

Validation: `go test -race ./pkg/avatar -run TestPebble -count=1`, `make fmt-check vet test-race build`, `node --test internal/studio/*.test.mjs port/agent-portrait.test.ts`, and `git diff --check` passed. The TypeScript fixtures are unchanged. Focused tests cover deterministic SVG, one body plus two allowed eye shapes, palette-independent geometry across all colors, and actual PNG eye separation and byte-identical circle/square framing across 128 seeds at 24 and 64 pixels. Existing small and circle export checks cover 16/24/32/64/256 pixels. Changed prose was scanned with `naturally`.

Real CLI proof: an isolated saved Sage recipe with seed `uplift-contract` was created and reopened through a fresh export process. Both 64-pixel circle SVG and PNG saved exports exactly match direct rendering. Evidence is at `/var/folders/9z/_hxsyhcx59d_cbrbhxfk9j000000gn/T/avatars-pebble-uplift-cli.vxemoaba/`. SHA-256: SVG `39fe1b0e285c03d340023d378f1267f705625a7d13d720902a817837910b0018`; PNG `3c4fe8fb919dc0237d78e998adb3c71072cb7930e4634c905ffcf5c666acb999`.

Visual findings: the cast has visible squat/tall, lean, and expression differences at icon size. Eyes remain separate and contrast with every body color on all three surrounds. The full shape fits comfortably inside circle exports. At 16 pixels the specific squint/wink nuance softens, as expected, while the silhouette and two marks remain readable. Gray provides less body contrast for Moss and Slate because their fixed palette fills are close to the surround; their eye contrast remains clear.

Style-authoring lesson: define a few whole-face expression recipes before adding bounded variation. Independent random dimensions can produce many technically different outputs that feel like the same character, or make unrelated eye angles look upset. Compare the same labeled seeds before and after, and test visible two-eye separation instead of requiring every eye to be an ellipse. Raster eye checks should exclude translucent body-edge pixels, whose unpremultiplied color can differ through rounding.

No assignment clarification was needed. The independent td context could not start the coordinator-owned task because it was already in progress; progress and handoff use `TD_CONTEXT_ID=avatars-pebble-uplift-astra` as assigned.

Independent review and live delivery: pending with the coordinator. This implementation does not merge, install, change Tailscale, or touch the live studio on port 7447. The coordinator owns final live collection creation and task approval.
