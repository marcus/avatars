#!/bin/sh
set -eu

output=${1:-}
if [ -z "$output" ]; then
  echo "usage: scripts/pebble-contact-sheet.sh OUTPUT_DIR" >&2
  exit 2
fi

binary=${AVATARS_BIN:-./bin/avatars}
mkdir -p "$output/palette" "$output/variety"
palette='walnut:Walnut cocoa:Cocoa clay:Clay apricot:Apricot butter:Butter moss:Moss sage:Sage teal:Teal sky:Sky denim:Denim lavender:Lavender mauve:Mauve rose:Rose coral:Coral slate:Slate'

for entry in $palette; do
  color=${entry%%:*}
  for size in 16 24 32 64 256; do
    "$binary" render --style pebble --color "$color" --seed palette-proof --format png --size "$size" --out "$output/palette/$color-$size.png"
    "$binary" render --style pebble --color "$color" --seed palette-proof --format png --size "$size" --circle --out "$output/palette/$color-$size-circle.png"
  done
done

index=0
while [ "$index" -lt 24 ]; do
  for size in 16 24 32 64 128; do
    "$binary" render --style pebble --color walnut --seed "walnut-proof:$index" --format png --size "$size" --out "$output/variety/$index-$size.png"
  done
  "$binary" render --style pebble --color walnut --seed "walnut-proof:$index" --format png --size 128 --circle --out "$output/variety/$index-circle.png"
  index=$((index + 1))
done

for theme in light dark gray; do
  case "$theme" in
    light) background='#F4F2ED'; foreground='#302C27' ;;
    dark) background='#202124'; foreground='#F4F2ED' ;;
    gray) background='#808080'; foreground='#FFFFFF' ;;
  esac
  {
    printf '%s\n' '<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Pebble visual proof</title>'
    printf '<style>body{margin:0;padding:24px;background:%s;color:%s;font:14px system-ui,sans-serif}a{color:inherit}h2{margin-top:32px}.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(170px,1fr));gap:16px}.card{text-align:center;padding:8px}.label{display:block;font-size:12px;margin:8px 0}.sizes{display:flex;align-items:center;gap:16px;flex-wrap:wrap}img{object-fit:contain}</style></head><body>\n' "$background" "$foreground"
    printf '%s\n' '<h1>Pebble visual proof</h1><p>Actual PNG exports. <a href="light.html">Light</a> · <a href="dark.html">Dark</a> · <a href="gray.html">Gray</a></p><h2>24 fixed seeds, square and circle</h2><div class="grid">'
    index=0
    while [ "$index" -lt 24 ]; do
      printf '<div class="card"><img src="variety/%s-128.png" width="128" height="128"><img src="variety/%s-circle.png" width="64" height="64"><span class="label">walnut-proof:%s</span><div>' "$index" "$index" "$index"
      for size in 16 24 32 64; do
        printf '<img src="variety/%s-%s.png" width="%s" height="%s" alt="%spx">' "$index" "$size" "$size" "$size" "$size"
      done
      printf '%s\n' '</div></div>'
      index=$((index + 1))
    done
    printf '%s\n' '</div><h2>15 colors at 16 / 24 / 32 / 64 / 256 pixels</h2><p>Seed: palette-proof. Square exports followed by circle exports.</p>'
    for entry in $palette; do
      color=${entry%%:*}
      printf '<h3>%s · %s</h3>\n' "${entry#*:}" "$color"
      for suffix in '' '-circle'; do
        printf '%s\n' '<div class="sizes">'
        for size in 16 24 32 64 256; do
          printf '<div class="card"><img src="palette/%s-%s%s.png" width="%s" height="%s"><span class="label">%spx%s</span></div>\n' "$color" "$size" "$suffix" "$size" "$size" "$size" "$suffix"
        done
        printf '%s\n' '</div>'
      done
    done
    printf '%s\n' '</body></html>'
  } > "$output/$theme.html"
done
cp "$output/light.html" "$output/index.html"
printf '%s\n' "$output/index.html"
