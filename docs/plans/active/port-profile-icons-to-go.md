# Port Profile Icon Generator to Go

**Status:** active  
**Goal:** Extract and port the pen-and-ink portrait/avatar generator from `comms-web` to Go with multi-format export (SVG, PNG, etc.) and standalone distribution.

## Context

The profile icon generator in `comms-web` (`src/lib/agent-portrait.ts`) creates deterministic, static pen-and-ink portraits from an arbitrary seed or agent identity. To enable wider use across CLIs, services, and native apps, we are creating a dedicated Go project (`avatars`).

## Core Requirements

1. **Go Core Engine (`pkg/avatar` or `internal/avatar`)**:
   - Deterministic 32-bit FNV-1a hash and 32-bit xorshift PRNG matching the reference implementation.
   - Exact recreation of the pen-and-ink aesthetic: paper palettes, ink tones, cloth textures, engraved backdrop hatching, 6 wardrobe styles, facial geometry, accessories, and hair styles.
   - Zero input string leakage into generated markup (pure numeric geometry).
   - Direct SVG generation matching the reference structure and viewBox (`0 0 64 72`).

2. **Multi-Format Export**:
   - **SVG**: Native vector output.
   - **PNG**: Rasterized export with configurable resolution (e.g., 64x72, 128x144, 256x288, 512x576) and optional circular masking.
   - Clean seam/adapter design so additional raster or vector formats can be plugged in without refactoring the generation core.

3. **CLI (`cmd/avatars`)**:
   - `avatars generate <seed> [flags]`
   - Flags:
     - `-s, --seed`: identity seed string (or positional argument)
     - `-f, --format`: output format (`svg`, `png`; default `svg`)
     - `--size`: dimensions or scale factor
     - `-o, --out`: output destination file (defaults to stdout or `<seed>.<format>`)
     - `-c, --circle`: render with circular crop/mask
   - `avatars --version`, `avatars --help`

4. **Testing Parity**:
   - Unit tests covering determinism (same seed produces identical output).
   - Entropy check (100 distinct seeds produce 100 distinct outputs).
   - Injection safety test (seeds with `<script>`, `<image>`, etc. never bleed into SVG markup).
   - Visual parity verification against reference SVG outputs.

5. **Release & Packaging**:
   - Local dev install and worktree switching via `make install-local` and `make install-worktree`.
   - Homebrew tap formula distribution (`marcus/tap/avatars`).
   - Cross-platform release binaries for Darwin and Linux (amd64/arm64) via GoReleaser and GitHub Actions.

## Steps

- [x] Create repository structure, git initialization, and private GitHub repo push.
- [x] Set up Go build, test, lint, and worktree tooling matching `~/code/tasks`.
- [x] Copy reference files (`agent-portrait.ts`, `agent-portrait.test.ts`, `AgentPortrait.svelte`) into `port/`.
- [x] Author `AGENTS.md` extracting SDLC guidelines from `agentic-sdlc-project-standards.md`.
- [ ] Implement core Go generator package.
- [ ] Implement SVG rendering engine and verify against reference tests.
- [ ] Implement PNG export rasterization.
- [ ] Wire up CLI commands and flags.
- [ ] End-to-end integration and smoke testing.
