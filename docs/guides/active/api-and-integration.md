# API and integration

Run `avatars serve` to serve the studio and API at `http://127.0.0.1:7447`. Use `--listen 127.0.0.1:PORT` to select another port. The service accepts loopback hosts and rejects browser requests that mutate another origin's library. Behind Tailscale Serve, set `--public-url https://HOST:PORT` or `AVATARS_PUBLIC_URL` to accept that exact HTTPS host and origin. The listener remains on loopback, and forwarded headers do not grant trust.

## Capabilities

| Method | Path | Result |
| --- | --- | --- |
| GET | `/api/v1/health` | Service name, API version, build version, and commit |
| GET | `/api/v1/styles` | `{styles: [...], formats: [...]}` |
| GET | `/api/v1/capabilities` | Command syntax, HTTP operations, global flags, and limits |
| GET | `/api/v1/instructions` | `{instructions: "..."}` with the agent workflow |
| GET | `/api/v1/collections` | `{collections: [...]}`, newest first |
| POST | `/api/v1/collections` | A saved collection; status 201 and a Location header |
| GET | `/api/v1/collections/{id}` | One collection |
| GET | `/api/v1/avatars/{id}` | One avatar |
| GET | `/api/v1/avatars/{id}.svg` | SVG image bytes |
| GET | `/api/v1/avatars/{id}.png` | PNG image bytes |
| GET | `/api/v1/render` | Stateless image bytes from a seed |

## Create a collection

Send `Content-Type: application/json` and an object:

```json
{"style":"gorey","count":12,"name":"First cast"}
```

All fields are optional. Defaults are style `gorey`, count 1, and a name derived from the style. Count must be 1–100. A name supports up to 120 Unicode characters. The optional `seed` supports up to 4096 UTF-8 bytes. Omit it or leave it empty to generate random seeds. Unknown fields are refused.

Pebble accepts one style-owned input. The default is Walnut, and saved recipes always contain the resolved value:

```json
{"style":"pebble","inputs":{"color":"sage"},"count":12}
```

Companions accepts `animal`, with values `dog`, `cat`, and `mixed`. Its default is `dog`. Every saved companion recipe contains `"inputs":{"animal":"dog"}` or `"inputs":{"animal":"cat"}`, including when the input was omitted or requested as Mixed:

```json
{"style":"companions","inputs":{"animal":"mixed"},"count":12}
```

Mixed chooses one dog or cat per final avatar seed, without promising exact batch balance. The same seed repeats the species through direct, CLI, HTTP, and saved rendering. Use `--animal mixed` in the CLI or `animal=mixed` on the stateless render route. This request changes neither the explicit Dog/Cat artwork nor the default Dog choice.

Pebble also accepts `"inputs":{"color":"random"}` or stateless `color=random`. This request selects independently from all 15 palette colors using each avatar's final seed. The same seed produces the same color across direct rendering, saving, and export. Each saved avatar contains its concrete palette value, never `random`. Random has no single swatch; its discovery entry contains a value and label without `swatch`.

Use `GET /api/v1/styles` or `avatars styles --json` to discover supported inputs. Companions publishes `inputs.animal` with `default`, and `values` containing `value` and `label` (Dogs, Cats, or Mixed). Pebble publishes `inputs.color` with the same fields plus color swatches. Styles reject nonempty inputs they do not support; Companions refuses color and Pebble refuses animal. Unknown nested fields and values are refused before persistence. Existing records without inputs remain valid.

An input descriptor can publish `mixed_value` to name its mixed request choice: `mixed` for Companions animal and `random` for Pebble color. Clients use that value when restoring a collection containing different concrete values. Individual avatar recipes and metadata still show the resolved value. Successful generation and ordinary refresh preserve the unsubmitted choice.

Style discovery also reports `native_width` and `native_height` for built-in styles. These describe artwork geometry, independent of export dimensions.

For one avatar, a supplied seed is used verbatim. For a batch, seeds are `SEED:0`, `SEED:1`, and so on. A batch is saved as one record; callers never see a partially saved collection. A successful retry creates another collection with new IDs, even when it reuses a seed.

The response shape is:

```json
{
  "id": "col_...",
  "name": "Sage pebbles",
  "style": "pebble",
  "created_at": "2026-09-08T00:00:00Z",
  "url": "/?collection=col_...",
  "avatars": [
    {
      "id": "av_...",
      "collection_id": "col_...",
      "style": "pebble",
      "seed": "generated-seed",
      "inputs": {"color":"sage"},
      "created_at": "2026-09-08T00:00:00Z",
      "url": "/?avatar=av_...",
      "svg_url": "/api/v1/avatars/av_....svg",
      "png_url": "/api/v1/avatars/av_....png"
    }
  ]
}
```

API URLs are relative to the service origin. CLI JSON expands them into absolute URLs. An avatar link selects that avatar in the inspector; a collection link opens its grid. Copied portrait links include `shape=portrait|circle` and export `width`/`height`, restoring that view when opened. A link without these options uses the portrait default.

## Export images

Saved image routes accept `width`, `height`, and `circle`. Omitted dimensions preserve the artwork's native ratio, which is 64 × 72 for Gorey, Gorey Expanded, and Picasso, 64 × 64 for Pebble, and 128 × 128 for Companions. If only one dimension is provided, the other is derived from the native ratio. A circular export defaults to the larger native dimension, and one supplied dimension sets both sides. With both dimensions supplied, the circle is inscribed in the requested canvas. Each dimension must be 1–2048.

```text
/api/v1/avatars/av_ID.png?width=256&height=256&circle=true
```

Stateless rendering additionally accepts required `seed` and optional `style`, `color`, `animal`, and `format` (`svg` by default). It never writes a collection. An empty seed is valid for stateless rendering and has the same deterministic meaning as in the TypeScript reference. URL-encode seeds with your HTTP client's query parameter support.

```text
/api/v1/render?seed=agent-42&style=gorey&format=svg
/api/v1/render?seed=agent-42&style=pebble&color=sage&format=svg
/api/v1/render?seed=sample&style=companions&animal=dog&format=png
```

Each scalar query parameter may appear only once. Saved image routes refuse `animal` and `color` overrides and render the stored recipe.

SVG uses `image/svg+xml`; PNG uses `image/png`. The filename appears in Content-Disposition. The service does not cache mutable library queries.

## Errors

API errors have this shape:

```json
{"error":{"code":"invalid_request","message":"count must be 1..100"}}
```

Invalid input returns 400, absent records return 404, and persistence or unexpected failures return 500. Rejected host or cross-origin mutations return 403. With `--json`, CLI errors use the same envelope on stderr. CLI exit codes are 0 for success, 1 for an operation failure, and 2 for invalid invocation.

## Extension boundaries

`pkg/avatar.Engine` owns separate registries for `Generator` and `Exporter`. Generators receive context and a seed, and return `Artwork` with bytes, media type, and native dimensions. Exporters receive that artwork and normalized size/crop options. Adapters must support concurrent calls and keep artwork self-contained. The built-in exporters accept SVG artwork; a future raster-native provider will need a compatible exporter.

A generator that supports a typed appearance input publishes its choices through `Style.Inputs` and implements the optional `InputGenerator` extension. Existing input-free generators and `Engine.Render` remain unchanged. Input-aware Go callers use the same validation and defaults as the other surfaces:

```go
image, err := engine.RenderWithInputs(ctx, "pebble", "agent-42", "png",
    avatar.Inputs{Color: "sage"},
    avatar.Options{Width: 256, Height: 256})

pets, err := engine.RenderWithInputs(ctx, "companions", "sample", "svg",
    avatar.Inputs{Animal: "cat"}, avatar.Options{})
```

`Engine.ResolveInputs` validates request choices and applies style defaults. `Engine.ResolveRecipeInputs(style, seed, inputs)` resolves one avatar's final appearance before saving or rendering. Styles with seed-dependent choices implement the optional `RecipeInputResolver`; Pebble uses it to turn Random into a concrete color; Companions turns Mixed into a concrete species. Other generators keep the existing defaulting and validation path. These methods reject inputs unsupported by the selected style. Saved exports pass the stored recipe to `RenderWithInputs` and expose no appearance override.

The saved library lives behind `library.Store`. The JSONL adapter handles paths, locks, append ordering, and sync. The application handles creation rules and lookup. The HTTP adapter and CLI call that application; the embedded studio calls the public HTTP API.

Saved records contain rendering recipes. Preserve the output contract for an existing style ID, and introduce a new ID if a style changes its geometry. Before adding a nondeterministic provider, extend persistence to retain the generated image itself. No AI model is invoked by this version.

## Proof commands

```sh
make fmt-check vet test-race build
node --test internal/studio/*.test.mjs port/agent-portrait.test.ts
```

The Go suite verifies TypeScript fixture parity, image exports, circular transparency, shared CLI/HTTP behavior, persistence across restart, and concurrent subprocess writes. Node is needed only to run or regenerate the reference fixtures, not to build or use Avatars.
