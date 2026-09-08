# Reference Implementation & Porting Guide

This folder contains the original TypeScript/Svelte implementation extracted from `comms-web` (`~/code/comms-web`), to serve as the reference specification for the Go port.

## Files

- `agent-portrait.ts`: The complete deterministic pen-and-ink portrait generator.
- `agent-portrait.test.ts`: Node.js test suite proving determinism, entropy/variety across 100 seeds, and injection safety.
- `AgentPortrait.svelte`: Svelte UI component demonstrating rendering options (circular crop, borders, preserveAspectRatio slice).

## Algorithm Details

1. **Hash Initialization**:
   - 32-bit FNV-1a hash over `seed` string:
     - `hash = 2166136261`
     - For each char: `hash = Math.imul(hash ^ codePoint, 16777619) >>> 0`
   - Store `initial = hash` for wardrobe/backdrop derivations.

2. **PRNG**:
   - 32-bit xorshift:
     - `hash ^= hash << 13`
     - `hash ^= hash >>> 17`
     - `hash ^= hash << 5`
     - `next(max) = (hash >>> 0) % max`

3. **Palette & Attributes**:
   - Ink: `#292820`
   - Paper: selected from 4 cream/stone tones (`#d6d0bb`, `#ded7c5`, `#cfcbb8`, `#d8d0be`) via `initial % 4`.
   - Wardrobe: derived independently using `Math.imul(initial ^ 0x6d2b79f5, 0x45d9f3b) >>> 0`.
   - Outfit: 6 styles (evening shirt & bow tie, tie & waistcoat, ribbed pullover, open shirt & braces, knit sweater, shawl collar cardigan).
   - Backdrop: 6 warm paper tones.
   - Cloth: 4 dark vintage cloth tones.
   - Face geometry: `faceWidth` (11..14), `chin` (44..49), `eyeY` (27..29), `nose` (34..37), `hair` (0..4), `accessory` (0..5), `coat` (0..2).

4. **Stroke & Geometry Generation**:
   - ViewBox: `0 0 64 72`.
   - Backdrop hatching: 18 dry engraved strokes.
   - Coat & torso hatching: 16 vertical/diagonal strokes.
   - Collar & shirt / outfit details.
   - Ears, jawline, and chin outline.
   - Facial features: hooded eyes, angular noses, restrained expressions.
   - Accessories: round spectacles, mustache, chin beard, monocle.
   - Hair styles: 5 distinct engraved pen-and-ink hairstyles.

5. **Security**:
   - Pure numeric geometry: no input characters are ever interpolated into the SVG markup.

## Go Port Requirements

1. **Core Package (`pkg/avatar` or `internal/avatar`)**:
   - Implement the numeric geometry generation with 100% visual and structural fidelity to the reference.
   - Provide an API to generate SVG strings/bytes for any seed.
   - Unit tests matching `agent-portrait.test.ts`:
     - Determinism: identical seed produces identical output.
     - Variety: 100 seeds produce 100 distinct outputs.
     - Safety: seeds containing HTML/SVG injection tags produce valid, clean SVGs without injected markup.

2. **Multi-Format Export**:
   - **SVG**: Default vector format.
   - **PNG**: Raster export at requested dimensions (e.g. 64x72, 128x144, 256x288, 512x576) via a pure Go SVG rasterizer or headless renderer without CGO dependencies.
   - Extensible design allowing additional formats (JPEG, WebP, etc.).

3. **CLI Commands (`cmd/avatars`)**:
   - `avatars generate [seed]` or `avatars generate --seed <string>`
   - Flags:
     - `--format`, `-f`: `svg` (default), `png`
     - `--size`, `-s`: target pixel size (e.g. 72, 144, 288, or WxH)
     - `--out`, `-o`: destination file path (default: stdout for SVG, or automatic filename `<seed>.<format>`)
     - `--circle`, `-c`: apply circular mask / crop
