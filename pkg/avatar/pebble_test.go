package avatar

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"image/png"
	"io"
	"strings"
	"testing"
)

func TestPebbleStyleInputsAndPalette(t *testing.T) {
	e := New()
	var style *Style
	for _, candidate := range e.Styles() {
		if candidate.ID == "pebble" {
			candidate := candidate
			style = &candidate
		}
	}
	if style == nil || style.Name != "Pebble" || style.Inputs == nil || style.Inputs.Color == nil {
		t.Fatalf("Pebble discovery is incomplete: %+v", style)
	}
	color := style.Inputs.Color
	if color.Default != "walnut" || len(color.Values) != 15 {
		t.Fatalf("unexpected color contract: %+v", color)
	}
	want := []string{
		"walnut:Walnut:#92744F", "cocoa:Cocoa:#655047", "clay:Clay:#B66F56",
		"apricot:Apricot:#E6AA78", "butter:Butter:#E4CA78", "moss:Moss:#71805A",
		"sage:Sage:#A5B59A", "teal:Teal:#4D8985", "sky:Sky:#9DBFD1",
		"denim:Denim:#607C9B", "lavender:Lavender:#AAA0C6", "mauve:Mauve:#A47B97",
		"rose:Rose:#D6A0A4", "coral:Coral:#D77F6D", "slate:Slate:#6C777D",
	}
	for i, choice := range color.Values {
		got := choice.Value + ":" + choice.Label + ":" + choice.Swatch
		if got != want[i] {
			t.Fatalf("palette entry %d: got %q, want %q", i, got, want[i])
		}
	}
	resolved, err := e.ResolveInputs("pebble", Inputs{})
	if err != nil || resolved.Color != "walnut" {
		t.Fatalf("default color: %+v, %v", resolved, err)
	}
	resolved, err = e.ResolveInputs("pebble", Inputs{Color: "sage"})
	if err != nil || resolved.Color != "sage" {
		t.Fatalf("explicit color: %+v, %v", resolved, err)
	}
	for _, test := range []struct {
		style string
		input Inputs
	}{
		{"pebble", Inputs{Color: "chartreuse"}},
		{"gorey", Inputs{Color: "sage"}},
	} {
		if _, err := e.ResolveInputs(test.style, test.input); !errors.Is(err, ErrInvalidInputs) {
			t.Fatalf("accepted unsupported inputs %+v for %s: %v", test.input, test.style, err)
		}
	}
}

func TestPebbleRenderingContract(t *testing.T) {
	ctx := context.Background()
	e := New()
	implicit, err := e.Render(ctx, "pebble", "quiet-character", "svg", Options{})
	if err != nil {
		t.Fatal(err)
	}
	explicit, err := e.RenderWithInputs(ctx, "pebble", "quiet-character", "svg", Inputs{Color: "walnut"}, Options{})
	if err != nil || !bytes.Equal(implicit, explicit) {
		t.Fatal("input-free rendering must use the Walnut default", err)
	}
	svg := string(explicit)
	for _, want := range []string{`viewBox="0 0 64 64"`, `fill="#92744F"`, `fill="#FFF4DC"`} {
		if !strings.Contains(svg, want) {
			t.Fatalf("missing %s in %s", want, svg)
		}
	}
	if strings.Count(svg, "<path ") != 1 || strings.Count(svg, "<ellipse ") != 2 {
		t.Fatalf("Pebble must contain one body and exactly two eyes: %s", svg)
	}
	seed := `<script>alert("seed leak")</script>`
	injected, err := e.RenderWithInputs(ctx, "pebble", seed, "svg", Inputs{Color: "sage"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(injected, []byte("script")) || bytes.Contains(injected, []byte("seed leak")) {
		t.Fatal("seed text leaked into Pebble SVG")
	}
	decoder := xml.NewDecoder(bytes.NewReader(injected))
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("invalid SVG: %v", err)
		}
	}
	other, err := e.RenderWithInputs(ctx, "pebble", "another-character", "svg", Inputs{Color: "walnut"}, Options{})
	if err != nil || bytes.Equal(implicit, other) {
		t.Fatal("different seeds must vary the silhouette and eyes", err)
	}
}

func TestPebbleSmallAndCirclePNGExports(t *testing.T) {
	e := New()
	for _, size := range []int{16, 24, 32, 64, 256} {
		for _, circle := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d-circle-%t", size, circle), func(t *testing.T) {
				data, err := e.RenderWithInputs(context.Background(), "pebble", "small-proof", "png", Inputs{Color: "cocoa"}, Options{Width: size, Height: size, Circle: circle})
				if err != nil {
					t.Fatal(err)
				}
				image, err := png.Decode(bytes.NewReader(data))
				if err != nil || image.Bounds().Dx() != size || image.Bounds().Dy() != size {
					t.Fatalf("invalid %dpx PNG: %v", size, err)
				}
				_, _, _, centerAlpha := image.At(size/2, size/2).RGBA()
				if centerAlpha == 0 {
					t.Fatal("character center is transparent")
				}
				_, _, _, cornerAlpha := image.At(0, 0).RGBA()
				if cornerAlpha != 0 {
					t.Fatal("transparent canvas margin is missing")
				}
			})
		}
	}
}
