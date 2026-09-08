# Give Pebble more personality

Status: ready for Astra. Task: `td-c3c8ed`.

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

Implementation and before/after evidence: pending.

Validation and lessons: pending.

Independent review and live delivery: pending.
