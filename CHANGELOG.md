# Changelog

## [Unreleased]

### Features

- Companions Mixed samples dogs and cats per avatar, preserves concrete saved species, and retains the Mixed choice for consecutive batches.
- Pebble Random selects from all 15 palette colors per avatar, with deterministic seeds and concrete saved colors across the studio, CLI, and HTTP API.
- Go Gorey generator with byte-for-byte TypeScript parity and extensible style and export adapters.
- SVG and pure Go PNG export with custom dimensions and circular cropping.
- Random saved collections, optional reproducible seeds, and a shared JSONL library.
- CLI generation, browsing, export, stateless rendering, structured help, and HTTP access.
- Embedded avatar studio with a portrait grid, collections, direct links, and export inspector.
- Loopback HTTP API over the same application core used by local CLI commands.
- Gorey Expanded adds a wider cast of engraved characters; Picasso adds restrained cubist portraits through the same generator interface.
- Portrait links preserve export shape and dimensions, and the service supports an explicit Tailscale HTTPS proxy origin.
- Pebble adds softly irregular two-eye characters with 15 stable colors across the Go library, CLI, HTTP API, saved recipes, and studio.
- Companions adds playful dog and cat portraits with saved Animal choices across Go, CLI, HTTP, and the studio.
- Field Birds adds invented bird species with eight body types, twelve plumage palettes, and field-guide proportions on a 160 × 160 canvas.
- Style discovery reports native dimensions so square portraits open with square export defaults.
- `avatars service ensure|status|stop` runs the studio as an on-demand local service that other tools can start, reuse, and stop safely.
- Homebrew tap, goreleaser archives for macOS and Linux, and a tag-triggered release workflow verified against `main`.
