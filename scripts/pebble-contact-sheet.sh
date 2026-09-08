#!/bin/sh
set -eu

output=${1:-}
if [ -z "$output" ]; then
  echo "usage: scripts/pebble-contact-sheet.sh OUTPUT_DIR" >&2
  exit 2
fi

binary=${AVATARS_BIN:-./bin/avatars}
mkdir -p "$output/palette" "$output/variety" "$output/sizes"

palette='walnut:Walnut cocoa:Cocoa clay:Clay apricot:Apricot butter:Butter moss:Moss sage:Sage teal:Teal sky:Sky denim:Denim lavender:Lavender mauve:Mauve rose:Rose coral:Coral slate:Slate'
for entry in $palette; do
  color=${entry%%:*}
  "$binary" render --style pebble --color "$color" --seed palette-proof --format png --size 160 --out "$output/palette/$color.png"
done

index=0
while [ "$index" -lt 12 ]; do
  "$binary" render --style pebble --color walnut --seed "walnut-proof:$index" --format png --size 160 --out "$output/variety/$index.png"
  index=$((index + 1))
done

for size in 16 24 32 64 256; do
  "$binary" render --style pebble --color walnut --seed scale-proof --format png --size "$size" --out "$output/sizes/$size.png"
done
"$binary" render --style pebble --color walnut --seed scale-proof --format png --size 256 --circle --out "$output/sizes/circle-256.png"

{
  printf '%s\n' '<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Pebble visual proof</title>'
  printf '%s\n' '<style>body{margin:0;padding:32px;background:#f4f2ed;color:#302c27;font:14px system-ui,sans-serif}h1{margin:0 0 8px}h2{margin-top:34px}.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(150px,1fr));gap:16px}.card{background:white;border:1px solid #ddd8ce;border-radius:12px;padding:12px;text-align:center}.card img{display:block;max-width:100%;height:auto;margin:auto}.label{display:block;margin-top:8px;color:#655f57;font-size:12px}.sizes{display:flex;align-items:flex-end;gap:24px;flex-wrap:wrap}.sizes .card{min-width:90px}.checker{background:repeating-conic-gradient(#eee 0 25%,#fff 0 50%) 50%/16px 16px}</style></head><body>'
  printf '%s\n' '<h1>Pebble visual proof</h1><p>Real PNG exports from fixed seeds.</p><h2>15 colors, seed: palette-proof</h2><div class="grid">'
  for entry in $palette; do
    color=${entry%%:*}
    label=${entry#*:}
    printf '<div class="card checker"><img src="palette/%s.png" width="160" height="160"><span class="label">%s · %s</span></div>\n' "$color" "$label" "$color"
  done
  printf '%s\n' '</div><h2>Walnut variety, seeds: walnut-proof:0–11</h2><div class="grid">'
  index=0
  while [ "$index" -lt 12 ]; do
    printf '<div class="card checker"><img src="variety/%s.png" width="160" height="160"><span class="label">walnut-proof:%s</span></div>\n' "$index" "$index"
    index=$((index + 1))
  done
  printf '%s\n' '</div><h2>Scale and circle framing, seed: scale-proof</h2><div class="sizes">'
  for size in 16 24 32 64 256; do
    printf '<div class="card checker"><img src="sizes/%s.png" width="%s" height="%s"><span class="label">%s × %s</span></div>\n' "$size" "$size" "$size" "$size" "$size"
  done
  printf '%s\n' '<div class="card checker"><img src="sizes/circle-256.png" width="256" height="256"><span class="label">256 × 256 circle</span></div></div></body></html>'
} > "$output/index.html"

printf '%s\n' "$output/index.html"
