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

For one avatar, a supplied seed is used verbatim. For a batch, seeds are `SEED:0`, `SEED:1`, and so on. A batch is saved as one record; callers never see a partially saved collection. A successful retry creates another collection with new IDs, even when it reuses a seed.

The response shape is:

```json
{
  "id": "col_...",
  "name": "First cast",
  "style": "gorey",
  "created_at": "2026-09-08T00:00:00Z",
  "url": "/?collection=col_...",
  "avatars": [
    {
      "id": "av_...",
      "collection_id": "col_...",
      "style": "gorey",
      "seed": "generated-seed",
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

Saved image routes accept `width`, `height`, and `circle`. Omitted dimensions preserve the native 64 × 72 portrait ratio. If only one dimension is provided, the other is derived from the native ratio. A circular export defaults to 72 × 72, and one supplied dimension sets both sides. With both dimensions supplied, the circle is inscribed in the requested canvas. Each dimension must be 1–2048.

```text
/api/v1/avatars/av_ID.png?width=256&height=256&circle=true
```

Stateless rendering additionally accepts required `seed` and optional `style` and `format` (`svg` by default). It never writes a collection. An empty seed is valid for stateless rendering and has the same deterministic meaning as in the TypeScript reference. URL-encode seeds with your HTTP client's query parameter support.

```text
/api/v1/render?seed=agent-42&style=gorey&format=svg
```

SVG uses `image/svg+xml`; PNG uses `image/png`. The filename appears in Content-Disposition. The service does not cache mutable library queries.

## Errors

API errors have this shape:

```json
{"error":{"code":"invalid_request","message":"count must be 1..100"}}
```

Invalid input returns 400, absent records return 404, and persistence or unexpected failures return 500. Rejected host or cross-origin mutations return 403. With `--json`, CLI errors use the same envelope on stderr. CLI exit codes are 0 for success, 1 for an operation failure, and 2 for invalid invocation.

## Extension boundaries

`pkg/avatar.Engine` owns separate registries for `Generator` and `Exporter`. Generators receive context and a seed, and return `Artwork` with bytes, media type, and native dimensions. Exporters receive that artwork and normalized size/crop options. Adapters must support concurrent calls and keep artwork self-contained. The built-in exporters accept SVG artwork; a future raster-native provider will need a compatible exporter.

The saved library lives behind `library.Store`. The JSONL adapter handles paths, locks, append ordering, and sync. The application handles creation rules and lookup. The HTTP adapter and CLI call that application; the embedded studio calls the public HTTP API.

Saved records contain rendering recipes. Preserve the output contract for an existing style ID, and introduce a new ID if a style changes its geometry. Before adding a nondeterministic provider, extend persistence to retain the generated image itself. No AI model is invoked by this version.

## Proof commands

```sh
make fmt-check vet test-race build
node --test port/agent-portrait.test.ts
```

The Go suite verifies TypeScript fixture parity, image exports, circular transparency, shared CLI/HTTP behavior, persistence across restart, and concurrent subprocess writes. Node is needed only to run or regenerate the reference fixtures, not to build or use Avatars.
