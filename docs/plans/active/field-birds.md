# Field Birds style brief

Status: implementing. Task: td-08c5c2.

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
