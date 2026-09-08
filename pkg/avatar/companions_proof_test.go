package avatar

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// TestCompanionProof is opt-in visual evidence from the real exporters. Use
// scripts/prove-companions.sh; generated files stay outside tracked source.
func TestCompanionProof(t *testing.T) {
	dir := os.Getenv("AVATARS_COMPANION_PROOF_DIR")
	if dir == "" {
		t.Skip("set AVATARS_COMPANION_PROOF_DIR to write contact sheets")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(name string, data []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	label := func(dst draw.Image, x, y int, text string) {
		d := font.Drawer{Dst: dst, Src: image.NewUniform(color.RGBA{52, 46, 42, 255}), Face: basicfont.Face7x13, Dot: fixed.P(x, y)}
		d.DrawString(text)
	}
	var html strings.Builder
	html.WriteString(`<!doctype html><meta charset="utf-8"><title>Companions renderer proof</title><style>body{font:15px system-ui;background:#e9e5da;margin:24px;color:#342e2a}section{display:grid;grid-template-columns:repeat(6,1fr);gap:12px}figure{margin:0;background:#fff9eb;padding:8px}img{max-width:100%;height:auto}figcaption{font-size:12px}.sizes{display:flex;align-items:end;gap:12px;flex-wrap:wrap}</style><h1>Companions renderer proof</h1>`)
	for _, animal := range []string{"dog", "cat"} {
		for _, circle := range []bool{false, true} {
			mode := "portrait"
			if circle {
				mode = "circle"
			}
			sheet := image.NewRGBA(image.Rect(0, 0, 6*176, 4*194))
			draw.Draw(sheet, sheet.Bounds(), image.NewUniform(color.RGBA{233, 229, 218, 255}), image.Point{}, draw.Src)
			fmt.Fprintf(&html, "<h2>%s / %s</h2><section>", animal, mode)
			for i := 0; i < 24; i++ {
				seed := fmt.Sprintf("companion-%02d", i)
				art, err := generateCompanion(context.Background(), seed, animal)
				if err != nil {
					t.Fatal(err)
				}
				opts := Options{Width: 160, Height: 160, Circle: circle}
				data, err := (PNGExporter{}).Export(context.Background(), art, opts)
				if err != nil {
					t.Fatal(err)
				}
				img, err := png.Decode(bytes.NewReader(data))
				if err != nil {
					t.Fatal(err)
				}
				x, y := (i%6)*176+8, (i/6)*194+8
				draw.Draw(sheet, image.Rect(x, y, x+160, y+160), img, image.Point{}, draw.Over)
				label(sheet, x, y+178, seed)
				name := fmt.Sprintf("%s-%s-%02d", animal, mode, i)
				write(name+".png", data)
				svg, err := (SVGExporter{}).Export(context.Background(), art, opts)
				if err != nil {
					t.Fatal(err)
				}
				write(name+".svg", svg)
				fmt.Fprintf(&html, `<figure><img src="%s.svg"><figcaption>%s</figcaption></figure>`, name, seed)
			}
			html.WriteString("</section>")
			var buf bytes.Buffer
			if err := png.Encode(&buf, sheet); err != nil {
				t.Fatal(err)
			}
			write(animal+"-"+mode+"-sheet.png", buf.Bytes())
		}
		// Three recipes shown at actual 32/64/256 px, square and circle.
		sizes := image.NewRGBA(image.Rect(0, 0, 800, 3*300))
		draw.Draw(sizes, sizes.Bounds(), image.NewUniform(color.RGBA{233, 229, 218, 255}), image.Point{}, draw.Src)
		fmt.Fprintf(&html, "<h2>%s / icon sizes</h2>", animal)
		for row, i := range []int{0, 7, 16} {
			seed := fmt.Sprintf("companion-%02d", i)
			art, err := generateCompanion(context.Background(), seed, animal)
			if err != nil {
				t.Fatal(err)
			}
			x := 12
			html.WriteString(`<div class="sizes">`)
			for _, circle := range []bool{false, true} {
				for _, size := range []int{32, 64, 256} {
					data, err := (PNGExporter{}).Export(context.Background(), art, Options{Width: size, Height: size, Circle: circle})
					if err != nil {
						t.Fatal(err)
					}
					img, err := png.Decode(bytes.NewReader(data))
					if err != nil {
						t.Fatal(err)
					}
					y := row*300 + 268 - size
					draw.Draw(sizes, image.Rect(x, y, x+size, y+size), img, image.Point{}, draw.Over)
					label(sizes, x, row*300+286, fmt.Sprintf("%dpx", size))
					x += size + 12
					name := fmt.Sprintf("%s-%02d-%d-circle-%t.png", animal, i, size, circle)
					write(name, data)
					fmt.Fprintf(&html, `<figure><img width="%d" src="%s"><figcaption>%s / %dpx</figcaption></figure>`, size, name, seed, size)
				}
			}
			html.WriteString(`</div>`)
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, sizes); err != nil {
			t.Fatal(err)
		}
		write(animal+"-sizes.png", buf.Bytes())
	}
	write("index.html", []byte(html.String()))
	t.Logf("visual proof: %s", dir)
}
