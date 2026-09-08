# avatars

Deterministic pen-and-ink avatar and profile icon generator in Go.

`avatars` generates reproducible, vintage-engraved portraits from arbitrary seed strings or agent identities. It produces SVG vectors or rasterized images (PNG and other formats), designed for local-first co-working tools, CLIs, web apps, and native interfaces.

## Features

- **Deterministic & Static**: The same seed identity always produces the exact same portrait.
- **Engraved Pen-and-Ink Aesthetic**: Intricate hand-drawn styling with varied outfits, hair, facial geometry, accessories, and engraved backdrop hatching.
- **Multiple Export Formats**: Native SVG vector output, rasterized PNG export at arbitrary resolutions, with support for circular masks.
- **Safe & Sanitized**: Input seed text never enters the generated markup; only seeded numeric geometry is rendered.
- **Fast & Self-Contained**: Compiled Go binary with no external runtime dependencies.

## Install

Homebrew is the supported macOS installation path:

```sh
brew install marcus/tap/avatars
avatars --version
```

To build from source:

```sh
make build
make install PREFIX="$HOME/.local"
```

## Development & Worktree Workflow

This repository follows the dev-install and worktree switching workflow used across our core Go toolchain:

```sh
make build               # build bin/avatars
make test                # run unit tests
make test-race           # run tests with race detector
make vet                 # run go vet
make fmt-check           # verify code formatting
```

### Machine-wide Dev Installation

To switch Homebrew's binary link to a local build:

```sh
make install-local       # canonical main checkout only: build and activate
make install-status      # inspect currently active build, commit, and source path
make use-homebrew        # switch back to the released Homebrew formula
make verify-homebrew     # upgrade and verify the Homebrew tap formula
```

When working in a Git worktree or feature branch:

```sh
make install-worktree    # activate current worktree build deliberately
```

## Porting & Reference

The original TypeScript and Svelte implementation from `comms-web` is preserved in `port/` along with test suites and porting specifications. See [port/README.md](port/README.md) for details.

## License

MIT License. See [LICENSE](LICENSE) for details.
