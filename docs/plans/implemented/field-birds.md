# Field Birds style brief

Status: implemented and delivered. Task: td-08c5c2.

## Outcome

Generate and save a varied set of invented bird species, then reopen and export them in the studio. Use a field-guide illustration approach inspired by the user's Sibley reference: believable anatomy, clear profile poses, restrained painted color, and fine feather detail.

## Visual contract

- Style ID `field-birds`, display name Field Birds. Native 160 × 160; useful at 64 px, with silhouette recognition at 32 px.
- Reference: [Sibley field-guide comparison spread](https://www.birderslibrary.com/quick_picks/first-look-sibley-eastern-western-guides-2nd-edition.htm). Study the isolated specimens, natural posture, small eyes, layered flight feathers, and quiet backgrounds. No reproduced illustrations ship in the generator. A small original renderer sample will live at `docs/references/field-birds.svg` as the repository visual reference.
- Eight anatomically distinct families: songbird, finch, longtail, groundbird, wader, shorebird, raptor, and nectarbird. These are construction families, not claims of real species identity.
- Twelve coherent palettes: earth, olive, slate, ochre, rust, teal, plum, and pale combinations. Seed controls family, palette, bill proportions, crest, wing bars, breast pattern, face marks, and facing direction. Complete seed identity stays independent of export options.
- Fine warm contours, shaded color washes, layered feathers, quiet ivory background, small natural eyes and thin jointed legs. No caricature faces, accessories, labels, real species names, habitat scenes, or copied compositions.
- All bird geometry must fit the square and circle. Tail length and leg length must contribute real silhouette variety. Inspect the actual SVG and PNG at small and large sizes.

## Input and surface contract

No appearance inputs. Every seed selects a complete fictional species. Existing shared seed, count, naming, persistence, export size, and circle controls apply. Independent seed sampling has no general batch-balance guarantee; inspect the delivered 48-bird collection for broad coverage. CLI: `avatars generate --style field-birds --seed field-notes --count 48`. HTTP creation: `{"style":"field-birds","seed":"field-notes","count":48}`. Save style and seed using the existing store; no persistence migration or new backend. Unsupported color/animal inputs fail through shared validation. Studio discovers the style and hides irrelevant controls. Existing styles must retain exact output.

## Scope and ownership

Worktree `/Users/marcus/code/avatars-field-birds`, branch `feat/field-birds`. Root owns implementation, proof, landing, installation, and replacement of the existing Avatars HTTP process only. Preserve its library and the existing Tailscale Serve routes. Never touch the default tmux server. Request independent review after artwork and journey proof, with a separate td context. No changes to the unrelated active service plan.

## Acceptance and handoff

Follow `docs/guides/active/creating-styles.md`. Commit this brief before implementation. Add focused determinism, cancellation, SVG safety, variation, PNG/circle, and CLI/HTTP save/reopen parity checks. Produce repeatable contact sheets and check desktop/mobile studio behavior. Run `make fmt-check vet test-race build`, Node studio/reference tests, and `git diff --check`. Record proof paths, commands, review, and the delivered collection here before moving this brief to implemented.

### Completed evidence

- Added the registered Field Birds generator, original SVG reference, reusable proof command, and discovery/integration documentation. No UI or persistence changes were needed.
- `make fmt-check vet test-race build` passes. Node studio/reference suite: 30 tests pass. `git diff --check` passes. CLI/HTTP SVG and PNG agree before and after store reopening. Unsupported appearance input is rejected without writing a collection.
- `scripts/prove-field-birds.sh /tmp/avatars-field-birds-proof` produces 48 labeled specimens plus 32/64/160/320 px square/circle examples. Fixed collection covers all eight families and twelve palettes. Inspected actual PNG contact sheet and enlarged details, corrected stray tail/crown strokes, and replaced a soft cheek spot with feathered markings. The visual result is a vector field-guide interpretation; it does not reproduce watercolor paper granulation.
- Isolated studio at port 17447 with library `/tmp/avatars-bird-studio.HsjC43`: generated 12 birds through the UI, selected a saved bird, changed circle/160 px exports, copied and reopened its link, and completed a PNG export. Inspected desktop cards and 390 × 844 mobile cards/inspector. Field Birds selected and irrelevant controls hidden throughout.
- Independent review: `review_field_birds`, `TD_CONTEXT_ID=field-birds-independent-review`, session `ses_22e826`. Focused tests and 48-bird / 32 / 64 / 320 px visual review passed. Fixed formatted-JSON test assertion and README style count/section placement. No outstanding renderer, safety, or parity findings. Reviewer evidence: `/tmp/avatars-birds-review`.
- Landed implementation `2d6592a` on main and installed through `make install-local`. The live service reports that commit. Replaced only the Avatars process; existing Tailscale routes were preserved. Verified all ten prior collections remain.
- Delivered [Birds of Elsewhere](https://aerie.tail53fd54.ts.net:7447/?collection=col_79264b370b74872181600da807807844), 48 birds from `field-notes`. Verified the actual browser displays all 48 cards and live saved SVG/PNG circle exports match the installed CLI. Local operational evidence, previous binary path, process ID, and logs are in `/tmp/avatars-field-birds-proof`.
- All requested work is complete. No remaining review findings.
