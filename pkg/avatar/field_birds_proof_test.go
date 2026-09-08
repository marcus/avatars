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

// Opt-in contact sheets from the production PNG and SVG exporters.
func TestFieldBirdsProof(t *testing.T) {
	dir := os.Getenv("AVATARS_BIRD_PROOF_DIR")
	if dir == "" {
		t.Skip("set AVATARS_BIRD_PROOF_DIR")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(name string, b []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), b, 0644); err != nil {
			t.Fatal(err)
		}
	}
	var html strings.Builder
	html.WriteString(`<!doctype html><meta charset="utf-8"><title>Field Birds proof</title><style>body{font:14px system-ui;background:#e6e2d8;color:#343b37;margin:24px}section{display:grid;grid-template-columns:repeat(6,1fr);gap:12px}figure{margin:0;background:#faf8ef}img{max-width:100%}figcaption{padding:8px;font-size:12px}.sizes{display:flex;gap:16px;align-items:end}</style><h1>Field Birds / actual renderer</h1><section>`)
	sheet := image.NewRGBA(image.Rect(0, 0, 6*200, 8*218))
	draw.Draw(sheet, sheet.Bounds(), image.NewUniform(color.RGBA{230, 226, 216, 255}), image.Point{}, draw.Src)
	engine := New()
	for i := 0; i < 48; i++ {
		seed := fmt.Sprintf("field-notes:%d", i)
		_, family, palette := fieldBirdRecipe(seed)
		for _, format := range []string{"png", "svg"} {
			data, err := engine.Render(context.Background(), "field-birds", seed, format, Options{Width: 192, Height: 192})
			if err != nil {
				t.Fatal(err)
			}
			write(fmt.Sprintf("bird-%02d.%s", i, format), data)
			if format == "png" {
				img, err := png.Decode(bytes.NewReader(data))
				if err != nil {
					t.Fatal(err)
				}
				x, y := i%6*200+4, i/6*218+4
				draw.Draw(sheet, image.Rect(x, y, x+192, y+192), img, image.Point{}, draw.Over)
				d := font.Drawer{Dst: sheet, Src: image.NewUniform(color.RGBA{50, 56, 51, 255}), Face: basicfont.Face7x13, Dot: fixed.P(x, y+208)}
				d.DrawString(fmt.Sprintf("%02d %s / %02d", i, birdForms[family].name, palette))
			}
		}
		fmt.Fprintf(&html, `<figure><img src="bird-%02d.svg"><figcaption>%02d / %s / palette %02d</figcaption></figure>`, i, i, birdForms[family].name, palette)
	}
	html.WriteString(`</section><h2>Size and circle checks</h2>`)
	for _, i := range []int{0, 4, 12, 20, 29, 35, 40, 47} {
		html.WriteString(`<div class="sizes">`)
		for _, circle := range []bool{false, true} {
			for _, size := range []int{32, 64, 160, 320} {
				data, err := engine.Render(context.Background(), "field-birds", fmt.Sprintf("field-notes:%d", i), "png", Options{Width: size, Height: size, Circle: circle})
				if err != nil {
					t.Fatal(err)
				}
				name := fmt.Sprintf("size-%d-%d-%t.png", i, size, circle)
				write(name, data)
				fmt.Fprintf(&html, `<figure><img src="%s"><figcaption>%d px / circle %t</figcaption></figure>`, name, size, circle)
			}
		}
		html.WriteString(`</div>`)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, sheet); err != nil {
		t.Fatal(err)
	}
	write("contact-sheet.png", buf.Bytes())
	write("index.html", []byte(html.String()))
}
