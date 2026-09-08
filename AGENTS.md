# AGENTS.md — Avatars Development Standards

Operational and architectural guidelines for coding agents working in this repository, extracted and consolidated from project SDLC standards.

---

## 1. Core Operating Principles

- **Harness- and model-agnostic**: Standards and tools must never couple to a single agent runner, harness, or model.
- **Local-first & inspectable**: Prefer local tools, reproducible CLI workflows, and inspectable file formats. Avoid proprietary control planes.
- **Commit as you go**: Commit small, coherent units of work. Never leave large uncommitted piles.
- **Push freely & private by default**: Push to GitHub regularly as a working backup. All repositories are private by default unless explicitly directed otherwise.
- **Land the branch**: Feature-branch work should be merged back to `main`. Do not leave repositories parked on an unmerged feature branch unless requested.
- **Branching**: Use Git worktrees for parallel or substantive feature work; committing directly to `main` is acceptable early on or for small fixes.
- **Finish in line**: Complete obvious next work in the current change. Never close an issue by spawning a pile of "should-do" follow-ups. Use `td` items only for non-obvious, optional, or genuinely deferred work.
- **Clean as you go**: Codebase hygiene is owned by every agent. If you encounter noisy logs, flaky tests, awkward build steps, worktree friction, linter warnings, or overgrown files: **surface it and fix it**.

---

## 2. Task Tracking (`td`)

Task tracking in this repository is managed with `td`.

<!-- td-agent-instructions:start -->
<!-- td-agent-instructions:version=3 -->

### Working with td

td keeps task context durable across sessions. In a new context, run `td usage --new-session -q` to see current work.

Use your judgment about how much tracking a task needs. For substantive work: `td start <id>`, record progress with `td log`, hand off with `td handoff <id>`, then `td review <id>`.

Closing needs a review. Say who did it (default trusted mode; delegated/strict allow only the first):

- independent session: `td approve <id> --reason "..."`
- a sub-agent: `td approve <id> --reviewed-by "<who>"`
- you: `td approve <id> --self-review --reason "..."`

Prefer a reviewer with its own `TD_CONTEXT_ID`; never name one who did not review.

Run `td usage` or `td <command> --help`.

<!-- td-agent-instructions:end -->

---

## 3. Documentation & Living Plans

- **Layout**: Follow the Clara/td standard:
  ```
  docs/
  ├── plans/
  │   ├── active/
  │   ├── implemented/
  │   └── deprecated/
  └── guides/
      ├── active/
      ├── implemented/
      └── deprecated/
  ```
- **Living documents**: Plans are living documents, not static post-mortems. Keep them updated informally as work proceeds.
- Move plans across `active/`, `implemented/`, and `deprecated/` as lifecycle progresses.
- State handoffs belong in **both `td` and the active plan**.
- Keep decision logs brief at the bottom; do not clutter plan bodies with historical decision archaeology.

---

## 4. Go Toolchain, Build & Worktree Standards

Matches the conventions established in `tasks`, `td`, `Sidecar`, and `Recall`.

### Build & Verification Commands
- `make build`: Compile `bin/avatars`.
- `make test`: Run unit tests (`go test ./...`).
- `make test-race`: Run race-detector suite (`go test -race ./...`).
- `make vet`: Run Go static analysis (`go vet ./...`).
- `make fmt-check`: Ensure formatting compliance (`gofmt -l`).
- `make clean`: Remove build and distribution artifacts.

### Dev Installation & Worktree Switching
Homebrew binary links in `/opt/homebrew/bin` (or `brew --prefix/bin`) are safely switched between the released formula and local checkout builds:
- `make install-local`: Compile and activate the canonical `main` checkout.
- `make install-worktree`: Deliberately activate the current worktree checkout.
- `make install-status`: Inspect currently active binary, version, commit, and source checkout path.
- `make use-homebrew`: Relink the installed Homebrew formula.
- `make verify-homebrew`: Update, upgrade, and test the Homebrew tap formula.

### Releases & CI/CD
- Versioning uses strict SemVer (`vX.Y.Z`) driven by `CHANGELOG.md` (`## [X.Y.Z] - YYYY-MM-DD`).
- `make release-dry-run` and `make release`: Run preflight checks, tag, and publish via `scripts/publish-release.sh`.
- Homebrew formula is published to `marcus/homebrew-tap` (`Formula/avatars.rb`).
- GitHub Actions CI verifies Linux & macOS builds on push and pull requests (`.github/workflows/ci.yml`).

---

## 5. Architecture & Code Structure

- **Creating styles**: Start with [Creating an avatar style](docs/guides/active/creating-styles.md) and commit a brief using [the style brief template](docs/guides/active/style-brief-template.md). Include reference art, supported inputs, persistence and surface contracts, and visual acceptance evidence so another agent can work from the repository alone.
- **Core generator (`pkg/avatar` or `internal/avatar`)**:
  - Deterministic pen-and-ink vector portrait generation.
  - Zero input string leakage into markup (pure numeric geometry derived via 32-bit FNV-1a hash and xorshift PRNG).
  - Port reference logic lives in `port/agent-portrait.ts` and `port/README.md`.
- **Modularity & Adapters**:
  - Use the adapter pattern for export formats (`svg`, `png`, and future codecs). The generation core must not depend directly on concrete rasterizers.
  - Prefer pure Go implementations without CGO dependencies for easy cross-platform compilation and distribution.
- **CLI (`cmd/avatars`)**:
  - Thin command wrapper over core library packages.

---

## 6. Runtime Environment Awareness

- **Sidecar**: Always check if operating inside Sidecar (`sidecar --agents`). Use Sidecar commands to open files, manage worktrees, message other agents, and present work to Marcus.
- **Comms**: Use `comms` for system-wide cross-agent messaging.
- **TUI design studio**: If introducing interactive terminal interfaces, consult the TUI design studio before designing in-app UI.
- **Clippy**: Aware that Marcus may send screenshots from the laptop via Clippy.

---

## 7. Communication & Decision Framing

When communicating decisions and choices to Marcus:
- **Lead with user & product impact**: Explain how the change affects the experience or capabilities.
- **Frame technical choices as trade-offs**: State the real trade across maintainability, performance, flexibility, security, and operability.
- **Jargon last**: Do not lead with internal function names, file paths, or pattern labels.
- **Links for Marcus**: Send studio, avatar, and collection links using the machine's verified Tailscale HTTPS address. Marcus is usually on another machine, so localhost links are not a useful handoff. Use `avatars --url https://HOST:PORT` for shareable CLI output. Inspect the current Tailscale Serve configuration before configuring a route; preserve other services and never enable public Funnel exposure by implication.
- **External prose**: Any user-facing copy or documentation meant for people other than Marcus must be passed through the `naturally` CLI.
