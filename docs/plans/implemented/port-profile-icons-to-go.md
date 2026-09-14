# Avatar generator and studio

Status: implemented. Tracking: `td-7c3910`.

## Outcome

Generate an avatar or a batch from the CLI or studio, see the same saved collection in the studio, link directly to an avatar or collection, and export SVG or PNG. The original Gorey style reproduces `port/agent-portrait.ts`. Gorey Expanded adds more varied people in the same engraved language. Picasso adds recognizable portraits with restrained cubist planes. Generation is random by default; a seed remains available to programs for reproducibility. The studio has no prompt or appearance controls.

## Architecture and decisions

- `pkg/avatar`: public, deterministic generator library with separate generator and export adapters. Gorey emits SVG; SVG and pure Go PNG exporters consume an artwork value. Additional style providers and output codecs can register without changing callers.
- `internal/library`: shared creation, validation, saved collection queries, and export behavior. Every batch is a durable collection with stable avatar IDs and links.
- `internal/store`: inspectable JSONL behind a narrow repository interface. Cross-process file locking coordinates direct CLI calls and the HTTP service; each batch is one append and sync. No database server or model runtime.
- `internal/httpapi`: versioned loopback HTTP interface over the same library. The studio uses these public endpoints exclusively.
- `internal/studio`: embedded HTML, CSS, and JavaScript. The Go binary serves the entire app without a frontend build or network dependencies.
- `internal/cli`: human and JSON output, command help, agent instructions, and machine-readable capabilities. Local calls use the library; `--url` selects HTTP access to a running service.
- The first service runs explicitly with `avatars serve --open`. Bind to loopback by default, refuse non-loopback exposure, and protect browser mutations against cross-origin requests. Tailscale Serve supplies private HTTPS through an explicitly configured `--public-url`; user-facing links use that verified origin. Automatic daemon lifecycle, accounts, public hosting, AI providers, and prompt steering are outside this slice.
- Saved records retain style and seed; the initial procedural style is a stable rendering contract. A future nondeterministic provider must persist its generated artwork through a storage extension before shipping.

## Capability parity

| Capability | Studio | CLI | HTTP |
| --- | --- | --- | --- |
| Discover styles and formats | Style selector | `styles`, `capabilities` | `GET /api/v1/styles` |
| Generate and save batch | Generate | `generate` | `POST /api/v1/collections` |
| Browse collections | Sidebar and grid | `list`, `show` | `GET /api/v1/collections[/{id}]` |
| Inspect an avatar | Inspector | `show` | `GET /api/v1/avatars/{id}` |
| Export SVG/PNG | Download | `export` | `GET /api/v1/avatars/{id}.{format}` |
| Reproducible stateless render | No steering UI by design | `render --seed` | `GET /api/v1/render` |
| Open exact selection | Deep links | Returned URLs | `/?avatar=ID`, `/?collection=ID` |

## Work and acceptance

- [x] Inspect reference generator, scaffold, local runtime, and Comms conventions.
- [x] Create isolated `avatar-studio` worktree and agree parallel file ownership.
- [x] Port generator and check TypeScript fixture parity including Unicode, entropy, and injection safety.
- [x] Implement SVG and PNG exports with dimensions and optional circular crop.
- [x] Implement shared collection library and JSONL persistence with restart and concurrency proof.
- [x] Implement CLI, HTTP, structured discovery, and consistent error mapping.
- [x] Build compact studio with grid, collection navigation, inspector, generation, downloads, and deep links.
- [x] Run focused tests, race suite, vet, formatting, build, and actual CLI/API/browser journeys.
- [x] Independently review meaningful changes and repair findings.
- [x] Update usage docs, run external prose through `naturally`, land on main, push private backup, install, and leave studio running.

- [x] Add Gorey Expanded and Picasso adapters, visually inspect collections, and verify all surfaces.
- [x] Preserve portrait/circle mode and dimensions in copied portrait URLs.
- [x] Add Tailscale link guidance to AGENTS.md, configure private HTTPS, and verify the remote journey.

## Completion evidence

- `make fmt-check vet test-race build` passes. Node reference and URL-state tests pass. The original Gorey SVG matches 108 TypeScript fixtures exactly; both new styles retain deterministic numeric SVG and use the same PNG/SVG exporters.
- CLI-to-HTTP tests create and export each of the three styles, verify shared metadata, and compare saved images with stateless render output. Store tests cover concurrent subprocesses, restarts, corrupt records, and cancellation.
- Independent reviews by the generator, library, and studio agents covered code they did not implement. Findings about export framing, cancellation, literal seeds, proxy origins, image recovery, and URL state were repaired and rechecked.
- Browser proof covers desktop and 390-pixel layouts, generation, live collection refresh, SVG/PNG downloads, copy links, browser history, failed-image recovery, and shape/dimension restoration. Opening a collection selects its style for subsequent generation; refresh preserves an explicit manual choice.
- Original Gorey, Gorey Expanded, and Picasso each have a 24-portrait collection in the local library. Both new contact sheets received author and independent visual review.
- The canonical main checkout is installed at `/opt/homebrew/bin/avatars`; the serving code build is `fe8c3cb`. The studio runs in Sidecar shell `sidecar-sh-avatars-2`, using `avatars serve --public-url https://YOUR_HOST.YOUR_TAILNET.ts.net:7447`.
- Tailscale HTTPS health, CLI collection creation, browser loading, image rendering, and circle links are verified at `https://YOUR_HOST.YOUR_TAILNET.ts.net:7447`. Existing Tailscale routes were preserved. A separate laptop SSH probe timed out, so no second-machine proof is claimed.
- Before/after Tailscale Serve snapshots are retained in `~/.local/state/avatars/rollback/`. Remove only this proxy with `tailscale serve --https=7447 off` if rollback is needed.

## Handoff

The implementation is complete and merged into main. The repository remains private, with no release tag. Agent usage and integration contracts live in `docs/guides/active/api-and-integration.md` and `avatars instructions`. Shared links use Tailscale as directed in AGENTS.md. Future nondeterministic providers must persist their artwork before they join the saved library; existing style output contracts remain stable.
