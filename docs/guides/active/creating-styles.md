# Creating an avatar style

Use this guide to add a style that works in the Go library, CLI, HTTP API, and studio. Start with a committed brief based on [the style brief template](style-brief-template.md). Keep the reference art in the repository so a new agent can work without conversation history.

## Read and inspect

Read `AGENTS.md`, the style brief, this guide, and [API and integration](api-and-integration.md). Inspect the current code before editing; the paths below are starting points, not a second specification. Run `td usage --new-session -q`, `sidecar --agents`, and `git status --short --branch`.

| Responsibility | Starting point |
| --- | --- |
| Native artwork, style registration, export framing | `pkg/avatar/engine.go` |
| A standalone procedural style | `pkg/avatar/picasso.go` |
| SVG framing and PNG conversion | `pkg/avatar/svg.go`, `pkg/avatar/png.go` |
| Saved recipes and creation rules | `internal/library/library.go` |
| Persistence adapter | `internal/store/jsonl.go` |
| CLI arguments and local/remote calls | `internal/cli/commands.go`, `cli.go`, `client.go` |
| Public HTTP routes and input parsing | `internal/httpapi/httpapi.go` |
| Agent help and capability discovery | `internal/discovery/discovery.go` |
| Studio controls, navigation, and previews | `internal/studio/assets/` |

## Establish the visual contract

Describe the silhouette, features, palette, background, variation, and details to exclude. Specify the smallest useful display size. A style is successful when its character reads at that size; a detailed large preview alone is not sufficient.

Choose a stable lowercase style ID and a short display name. Existing saved avatars are recipes. Changing geometry, palette meanings, defaults, or random-number ordering under a shipped style ID can change every saved avatar. Preserve existing outputs; use a new style ID for incompatible revisions. Keep seed text out of SVG markup and metadata.

Implement a complete composition from each seed. Vary a few meaningful characteristics rather than adding arbitrary noise. Bound every geometric choice so features remain inside the silhouette and crop. Do not change a shared random helper in a way that alters another style.

Return self-contained `Artwork` with its native width, height, media type, and bytes. Use the shared exporters for sizing and circle masking. Native dimensions belong to the artwork and may differ by style. Keep generator calls deterministic and safe for concurrent use, with all mutable random state local to one call. Handle cancellation. Avoid fonts, network resources, external images, scripts, and new rendering dependencies for procedural SVG styles.

## Add only the inputs the style needs

Generation inputs change the identity of the artwork. Export options change its framing or encoding. Keep those concepts separate: body color belongs to the saved recipe; width, height, and circle belong to export options.

For each new input:

1. Specify a stable machine value, human label, default, accepted values, and behavior when omitted.
2. Publish its supported values through style discovery. The studio gets its choices from the API. Do not maintain another palette or validation table in JavaScript.
3. Resolve defaults and validate in the shared core before persistence or rendering. Reject values and inputs unsupported by the selected style. CLI and HTTP translate the same error into their normal envelopes.
4. Save the resolved input with every avatar, so retrieval, export, restart, and copied links recreate its appearance. Do not fold an input into the seed or store it only in browser state.
5. Expose generation through CLI and HTTP as well as the UI. Update human help, JSON discovery, integration examples, and the guide when the contract changes.
6. Preserve records that omit new fields, and keep input-free calls compatible. Saved exports use the saved recipe; they must refuse attempts to override its appearance.

Use a narrow typed input and a small adapter extension when required. A schema engine, plugin loader, or arbitrary options map is unnecessary for one enum. Keep the existing input-free `Generator` and `Engine.Render` usable where practical; put optional input support behind a documented seam. The core must not switch on a style ID to implement that style's rules.

The current Go seam uses `Style.Inputs` for discovery, `Engine.ResolveInputs` for shared defaulting and validation, and the optional `InputGenerator.GenerateWithInputs` method for rendering. Callers with an input use `Engine.RenderWithInputs`; ordinary generators and callers continue to use `Generate` and `Engine.Render`. Add another typed descriptor only when a shipped style needs it.

## Build and prove one journey

First connect one seed and one input value through native artwork, rendering, saving, and re-export. Then connect CLI, HTTP, and the studio. Expand the visual palette after that path works.

Meaningful evidence includes:

- A fixed seed and inputs produce identical SVG/PNG bytes through local and HTTP calls.
- Saved recipes survive a new store/service instance, and saved exports match the stateless recipe.
- Unknown input values, unsupported style/input combinations, and malformed requests fail before a record is written.
- Existing style fixtures and records still render unchanged.
- PNG dimensions, transparency, and circle clipping are correct; SVG is well formed and contains no seed text or external resources.
- The studio shows only supported controls, creates the requested result, and can reopen, export, and share it with the selected crop and dimensions.
- A background library refresh preserves a manual input choice, while opening a saved avatar or collection restores its resolved recipe.

Do not duplicate every assertion across all surfaces. Cover domain rules at the core and add focused adapter and real-process parity checks.

Run the required checks from the worktree:

```sh
make fmt-check vet test-race build
node --test internal/studio/*.test.mjs port/agent-portrait.test.ts
git diff --check
```

Use `mktemp -d` and an unused loopback port for proof libraries and servers. Build the embedded studio before browser testing it. Never modify the user's saved records to manufacture test fixtures, and never stop or replace the default tmux server. Stop only proof processes you own.

## Review the actual artwork

Produce a contact sheet from the real generator, with fixed seeds and labeled input values. An HTML sheet containing the rendered SVGs is sufficient. Include native-size and enlarged examples, each palette choice, several silhouettes or expressions, and portrait/circle exports. Inspect the result as an image or in a browser; reading SVG paths is not visual review.

Check the smallest intended sizes, eye contrast, spacing, silhouette margins, and whether randomness changes the intended character. Inspect the studio at desktop and narrow mobile widths. Do not use AI-generated mockups as evidence for a procedural renderer.

Record the commands, proof paths, visual findings, and any remaining concerns in the controlling plan. Generated samples may stay in an ignored local proof directory; commit a small deterministic recipe or script if it makes future review reproducible. Avoid large generated image dumps.

## Handoff and completion

Commit coherent changes and request an independent review. Give the reviewer the brief and acceptance evidence. Repair findings, rerun affected checks, and update the guide to describe the resulting extension seam accurately.

For a handoff experiment, give the implementer only the path to the committed brief in a fresh context. Put clarification in that document, commit it, and record that the experiment needed another instruction. Do not quietly supply implementation advice in chat. The implementer records unclear or missing guidance even if they resolve it independently.

The coordinating agent lands and installs the reviewed change, verifies the running service, and presents a generated collection using a verified Tailscale HTTPS link. Keep operational machine details out of reusable style code. The controlling brief assigns who may restart the service; an implementation agent should not assume it owns the live studio.
