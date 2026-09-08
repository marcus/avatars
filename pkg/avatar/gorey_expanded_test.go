package avatar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"image/png"
	"io"
	"strings"
	"testing"
)

func TestGoreyExpandedDeterminismAndVariety(t *testing.T) {
	generator := GoreyExpanded{}
	style := generator.Style()
	if style.ID != "gorey-expanded" || style.Name != "Gorey Expanded" {
		t.Fatalf("unexpected style: %+v", style)
	}
	seen := map[[32]byte]bool{}
	for i := 0; i < 100; i++ {
		seed := fmt.Sprintf("expanded-%d", i)
		first, err := generator.Generate(context.Background(), seed)
		if err != nil {
			t.Fatal(err)
		}
		again, err := generator.Generate(context.Background(), seed)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first.Data, again.Data) {
			t.Fatalf("seed %q is not deterministic", seed)
		}
		if first.MediaType != "image/svg+xml" || first.Width != 64 || first.Height != 72 {
			t.Fatal("unexpected native artwork contract")
		}
		hash := sha256.Sum256(first.Data)
		if seen[hash] {
			t.Fatalf("duplicate portrait at %q", seed)
		}
		seen[hash] = true
		dec := xml.NewDecoder(bytes.NewReader(first.Data))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("invalid SVG for %q: %v", seed, err)
			}
		}
	}
}

func TestGoreyExpandedSeedSafetyAndCancellation(t *testing.T) {
	generator := GoreyExpanded{}
	for _, seed := range []string{"", "日本語 🦉 é", "e\u0301", `</svg><script>seedMarker()</script><svg onload="seedMarker()">`, "a\x00b"} {
		art, err := generator.Generate(context.Background(), seed)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"script", "seedMarker", "onload", "日本語", "🦉", "\x00"} {
			if strings.Contains(string(art.Data), forbidden) {
				t.Fatalf("seed content leaked: %q", forbidden)
			}
		}
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := generator.Generate(canceled, "seed"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled generation: %v", err)
	}
}

func TestGoreyExpandedExports(t *testing.T) {
	generator := GoreyExpanded{}
	// Exercise varied hair, garments and facial silhouettes through the real
	// rasterizer, so malformed path commands cannot hide in valid XML.
	for i := 0; i < 30; i++ {
		art, err := generator.Generate(context.Background(), fmt.Sprintf("expanded-%d", i))
		if err != nil {
			t.Fatal(err)
		}
		options := Options{Width: 128, Height: 144}
		if i%2 == 0 {
			options = Options{Width: 96, Height: 96, Circle: true}
		}
		svg, err := (SVGExporter{}).Export(context.Background(), art, options)
		if err != nil {
			t.Fatal(err)
		}
		dec := xml.NewDecoder(bytes.NewReader(svg))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		data, err := (PNGExporter{}).Export(context.Background(), art, options)
		if err != nil {
			t.Fatalf("rasterizing portrait %d: %v", i, err)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Dx() != options.Width || img.Bounds().Dy() != options.Height {
			t.Fatalf("incorrect export dimensions: %v", img.Bounds())
		}
		_, _, _, center := img.At(options.Width/2, options.Height/2).RGBA()
		if center != 65535 {
			t.Fatalf("portrait %d has missing artwork", i)
		}
		_, _, _, corner := img.At(0, 0).RGBA()
		if options.Circle && corner != 0 {
			t.Fatal("circular export has opaque corners")
		}
		dark := 0
		for y := 0; y < options.Height; y++ {
			for x := 0; x < options.Width; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				if a > 60000 && r < 35000 && g < 35000 && b < 35000 {
					dark++
				}
			}
		}
		if dark < 100 {
			t.Fatalf("portrait %d lost its ink detail", i)
		}
	}
}
