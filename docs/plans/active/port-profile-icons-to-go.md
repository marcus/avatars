# Avatar generator and studio

Status: active. Tracking: `td-7c3910`.

## Outcome

Generate an avatar or a batch from the CLI or studio, see the same saved collection in the studio, link directly to an avatar or collection, and export SVG or PNG. The initial Gorey style reproduces `port/agent-portrait.ts`. Generation is random by default; a seed remains available to programs for reproducibility. The studio has no prompt or appearance controls.

## Architecture and decisions

- `pkg/avatar`: public, deterministic generator library with separate generator and export adapters. Gorey emits SVG; SVG and pure Go PNG exporters consume an artwork value. Additional style providers and output codecs can register without changing callers.
- `internal/library`: shared creation, validation, saved collection queries, and export behavior. Every batch is a durable collection with stable avatar IDs and links.
- `internal/store`: inspectable JSONL behind a narrow repository interface. Cross-process file locking coordinates direct CLI calls and the HTTP service; each batch is one append and sync. No database server or model runtime.
- `internal/httpapi`: versioned loopback HTTP interface over the same library. The studio uses these public endpoints exclusively.
- `internal/studio`: embedded HTML, CSS, and JavaScript. The Go binary serves the entire app without a frontend build or network dependencies.
- `internal/cli`: human and JSON output, command help, agent instructions, and machine-readable capabilities. Local calls use the library; `--url` selects HTTP access to a running service.
- The first service runs explicitly with `avatars serve --open`. Bind to loopback by default, refuse non-loopback exposure, and protect browser mutations against cross-origin requests. Automatic daemon lifecycle, accounts, public hosting, AI providers, prompt steering, and new styles are outside this slice.
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
- [ ] Port generator and check TypeScript fixture parity including Unicode, entropy, and injection safety.
- [ ] Implement SVG and PNG exports with dimensions and optional circular crop.
- [ ] Implement shared collection library and JSONL persistence with restart and concurrency proof.
- [ ] Implement CLI, HTTP, structured discovery, and consistent error mapping.
- [ ] Build compact studio with grid, collection navigation, inspector, generation, downloads, and deep links.
- [ ] Run focused tests, race suite, vet, formatting, build, and actual CLI/API/browser journeys.
- [ ] Independently review meaningful changes and repair findings.
- [ ] Update usage docs, run external prose through `naturally`, land on main, push private backup, install, and leave studio running.

## Current handoff

Generator agent owns `pkg/avatar`, fixtures, and Go dependencies. Library agent owns `internal/library` and `internal/store`. Studio agent owns `internal/studio`. Primary agent owns integration, CLI, HTTP, discovery, docs, and operational proof. Shared worktree: `/Users/marcus/code/avatars-avatar-studio`. Each contributor commits only owned files. No release tag or change to repository visibility is part of this work.
