# Avatars

A local avatar generator with a CLI, HTTP API, and creative studio. Generate a collection, pick a portrait, and export it as SVG or PNG. Avatars created by agents appear in the same library as those created in the studio.

The first style, **Gorey**, draws engraved pen-and-ink characters with varied faces, hair, clothes, and paper tones. Generation is random by default. Programs can provide a seed to reproduce a portrait exactly.

## Start the studio

Build with Go 1.27 or newer:

```sh
make build
./bin/avatars serve --open
```

The studio opens at `http://127.0.0.1:7447`. It includes a collection browser, portrait grid, and export inspector. Choose a style and batch size, then generate. Select a portrait to adjust export dimensions, apply a circular crop, download SVG or PNG, or copy its direct link.

The service runs in the foreground until you press Ctrl-C. It serves both the API and the embedded studio. No Node.js runtime or frontend build is needed.

## Generate from the CLI

In another terminal:

```sh
# Save random portraits and return links to the collection and each avatar.
./bin/avatars generate --count 12 --name "First cast" --json

# Browse the same library shown in the studio.
./bin/avatars list --json
./bin/avatars show AVATAR_ID --json

# Export a saved portrait as a circular PNG.
./bin/avatars export AVATAR_ID --format png --size 256 --circle --out icon.png

# Generate, save, and export a single avatar in one command.
./bin/avatars generate --out portrait.svg

# Produce a reproducible image without saving a collection.
./bin/avatars render --seed agent-42 --format svg --out agent.svg
```

Replace `AVATAR_ID` with an ID returned by `generate` or `list`. Each saved avatar and collection has a direct studio URL. The studio refreshes when another process creates a collection.

`generate` works while the service is stopped. To use a running service's library through HTTP, add `--url http://127.0.0.1:7447` or set `AVATARS_URL`. Otherwise the CLI opens the local library through the same application core used by the API.

SVG is the default export format. `--size 256` requests a square canvas; `--size 256x288` sets both dimensions. Native portraits are 64 × 72. Rectangular exports preserve the artwork's aspect ratio, with transparent margins when needed. Circular exports crop the center and leave transparent corners. Each dimension supports 1–2048 pixels.

`render` and `export` write image bytes to stdout when `--out` is omitted. With `--json`, they return base64 data and its media type instead. `--out FILE` creates a new file and refuses to overwrite an existing one.

## Agent help and HTTP

```sh
./bin/avatars help generate
./bin/avatars instructions
./bin/avatars capabilities
./bin/avatars styles --json
```

Create a saved collection through HTTP:

```sh
curl -sS http://127.0.0.1:7447/api/v1/collections \
  -H 'Content-Type: application/json' \
  -d '{"style":"gorey","count":12,"name":"First cast"}'
```

For a stateless image URL:

```text
http://127.0.0.1:7447/api/v1/render?seed=agent-42&format=png&width=256&height=256&circle=true
```

See the [API and integration guide](docs/guides/active/api-and-integration.md) for endpoints, response shapes, storage, and extension seams.

## Use the Go library

```go
package main

import (
    "context"
    "os"

    "github.com/marcus/avatars/pkg/avatar"
)

func main() {
    engine := avatar.New()
    image, err := engine.Render(context.Background(), "gorey", "agent-42", "png",
        avatar.Options{Width: 256, Height: 256, Circle: true})
    if err != nil {
        panic(err)
    }
    if err := os.WriteFile("icon.png", image, 0644); err != nil {
        panic(err)
    }
}
```

Generator adapters produce native artwork; exporter adapters convert it to the requested format. Register additional implementations with `RegisterGenerator` and `RegisterExporter`. The built-in SVG and PNG exporters are pure Go and require no CGO or external rasterizer. The native Gorey SVG matches the preserved TypeScript reference byte-for-byte.

## Local library

Saved collections live in `$XDG_DATA_HOME/avatars/collections.jsonl`, or `~/.local/share/avatars/collections.jsonl` when `XDG_DATA_HOME` is unset. Use `--data-dir PATH` or `AVATARS_DATA_DIR` to select another library, and use the same setting when starting the studio. This setting cannot be combined with `--url`, because the server owns the selected library.

The JSONL file stores one collection per line, including stable IDs, style, and seeds. File locking coordinates CLI and HTTP writes. Back it up or inspect it with ordinary file tools; stop writers before editing or replacing it. An incomplete or corrupt record is reported with its location and is preserved for repair.

The service listens on loopback. To reach it from other devices on your tailnet, use Tailscale Serve and configure its HTTPS origin:

```sh
avatars serve --public-url https://YOUR_HOST.YOUR_TAILNET.ts.net:7447
# In another terminal, add this route without changing existing Tailscale services:
tailscale serve --bg --https=7447 http://127.0.0.1:7447
```

Use the Tailscale URL in the browser and with CLI `--url` when sharing links. `AVATARS_PUBLIC_URL` can supply the serve flag's default. The proxy host and browser origin are accepted only when configured explicitly. Accounts, public hosting, AI providers, and appearance steering are outside this version.

## Development

```sh
make fmt-check vet test test-race build
make install PREFIX="$HOME/.local"
```

The repository also supports the shared Homebrew dev-install workflow:

```sh
make install-local       # build and activate the canonical main checkout
make install-worktree    # deliberately activate the current worktree
make install-status      # inspect the active binary and source checkout
make use-homebrew        # return to an installed released formula
```

Release and Homebrew packaging are prepared in `scripts/` and `packaging/`; this development version has no published release. Reference TypeScript and Svelte code lives in [port/](port/README.md).

MIT License. See [LICENSE](LICENSE).
