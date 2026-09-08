# Avatars

A local avatar generator with a CLI, HTTP API, and creative studio. Generate a collection, pick a portrait, and export it as SVG or PNG. Avatars created by agents appear in the same library as those created in the studio.

Five styles are available:

| Style | Character |
| --- | --- |
| **Gorey** (`gorey`) | The original engraved pen-and-ink portraits, preserved exactly. |
| **Gorey Expanded** (`gorey-expanded`) | A wider cast with more face shapes, ages, hairstyles, clothes, and accessories. |
| **Picasso** (`picasso`) | Recognizable faces drawn with bold contours, muted color, and restrained cubist planes. |
| **Pebble** (`pebble`) | Pudgy, softly irregular characters with two tiny eyes and 15 palette colors and a Random choice. |
| **Companions** (`companions`) | Playful dogs and cats with expressive ears, warm coats, and lively ink contours. |

Generation is random by default. Programs can provide a seed to reproduce a portrait exactly. Each style uses the same SVG/PNG exporters and library workflow.

## Start the studio

Build with Go 1.27 or newer:

```sh
make build
./bin/avatars serve --open
```

The studio opens at `http://127.0.0.1:7447`. It includes a collection browser, portrait grid, and export inspector. Choose a style and batch size, then generate. Companions shows an Animal dropdown for Dogs, Cats, or Mixed; Pebble shows Color. Each saved portrait keeps its selected species or color. Select a portrait to adjust export dimensions, apply a circular crop, download SVG or PNG, or copy its direct link with the selected shape and dimensions.

The inspector offers Dark, Light, and Gray preview backgrounds. Copied portrait links retain the selected surround, shape, and dimensions. Background choices affect the preview only; SVG and PNG exports keep their original transparency.

The service runs in the foreground until you press Ctrl-C. It serves both the API and the embedded studio. No Node.js runtime or frontend build is needed.

## Generate from the CLI

In another terminal:

```sh
# Save random portraits and return links to the collection and each avatar.
./bin/avatars generate --count 12 --name "First cast" --json
./bin/avatars generate --style picasso --count 12 --name "Picasso studies" --json
./bin/avatars generate --style pebble --color random --count 12 --name "Mixed pebbles" --json
./bin/avatars generate --style companions --animal mixed --count 12 --name "Mixed companions" --json

# Browse the same library shown in the studio.
./bin/avatars list --json
./bin/avatars show AVATAR_ID --json

# Export a saved portrait as a circular PNG.
./bin/avatars export AVATAR_ID --format png --size 256 --circle --out icon.png

# Generate, save, and export a single avatar in one command.
./bin/avatars generate --out portrait.svg

# Produce a reproducible image without saving a collection.
./bin/avatars render --seed agent-42 --format svg --out agent.svg
./bin/avatars render --style companions --animal dog --seed sample --format png --out dog.png
```

Replace `AVATAR_ID` with an ID returned by `generate` or `list`. Each saved avatar and collection has a direct studio URL. The studio refreshes when another process creates a collection.

`generate` works while the service is stopped. To use a running service's library through HTTP, add `--url http://127.0.0.1:7447` or set `AVATARS_URL`. Otherwise the CLI opens the local library through the same application core used by the API.

Companions defaults to Dogs. `--animal mixed` chooses a dog or cat independently for each final seed, without promising an exact balance in a batch. Saved avatars retain the concrete species. Mixed stays selected for consecutive generation and refresh; opening a mixed-species collection suggests Mixed, while selecting a portrait restores Dogs or Cats.

Pebble defaults to Walnut. `--color random` picks from all 15 palette colors independently for each avatar; a fixed seed repeats its color. Saved recipes contain the chosen palette value, so later exports keep that color. In the studio, Random stays selected after generation and ordinary refresh. Opening a saved portrait selects its concrete color; opening a mixed-color collection selects Random.

SVG is the default export format. `--size 256` requests a square canvas; `--size 256x288` sets both dimensions. Gorey and Picasso portraits use a 64 × 72 native canvas; Pebble uses 64 × 64, and Companions uses 128 × 128. Style discovery reports these as `native_width` and `native_height`. Rectangular exports preserve the artwork's aspect ratio, with transparent margins when needed. Circular exports crop the center and leave transparent corners. Each dimension supports 1–2048 pixels.

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

To add a style, start with [Creating an avatar style](docs/guides/active/creating-styles.md) and the [style brief template](docs/guides/active/style-brief-template.md). The guide covers artwork, inputs, saved recipes, surface integration, and visual review.

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

The JSONL file stores one collection per line, including stable IDs, style, seeds, and resolved generation inputs. File locking coordinates CLI and HTTP writes. Back it up or inspect it with ordinary file tools; stop writers before editing or replacing it. An incomplete or corrupt record is reported with its location and is preserved for repair.

The service listens on loopback. To reach it from other devices on your tailnet, use Tailscale Serve and configure its HTTPS origin:

```sh
avatars serve --public-url https://YOUR_HOST.YOUR_TAILNET.ts.net:7447
# In another terminal, add this route without changing existing Tailscale services:
tailscale serve --bg --https=7447 http://127.0.0.1:7447
```

Use the Tailscale URL in the browser and with CLI `--url` when sharing links. `AVATARS_PUBLIC_URL` can supply the serve flag's default. The proxy host and browser origin are accepted only when configured explicitly. Pebble color and companion animal choices are saved appearance inputs. Omitting them resolves to Walnut or Dogs. Other styles refuse unsupported nonempty inputs, and saved exports refuse appearance overrides. Accounts, public hosting, and AI providers are outside this version.

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
