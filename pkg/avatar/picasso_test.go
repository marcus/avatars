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

func TestPicassoDeterminismVarietyAndSafety(t *testing.T) {
	generator := Picasso{}
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		seed := fmt.Sprintf("picasso-%d", i)
		one, err := generator.Generate(context.Background(), seed)
		if err != nil {
			t.Fatal(err)
		}
		two, err := generator.Generate(context.Background(), seed)
		if err != nil || !bytes.Equal(one.Data, two.Data) {
			t.Fatalf("unstable recipe for %q: %v", seed, err)
		}
		if seen[string(one.Data)] {
			t.Fatalf("duplicate artwork for %q", seed)
		}
		seen[string(one.Data)] = true
		if one.Width != 64 || one.Height != 72 || one.MediaType != "image/svg+xml" {
			t.Fatalf("invalid native artwork: %+v", one)
		}
	}
	for _, seed := range []string{"", "🦉你好", `<script>attack()</script><svg onload="evil()">`} {
		art, err := generator.Generate(context.Background(), seed)
		if err != nil {
			t.Fatal(err)
		}
		for _, token := range []string{"script", "attack", "onload", "evil", "NaN", "%!"} {
			if strings.Contains(string(art.Data), token) {
				t.Fatalf("invalid or leaked SVG token %q", token)
			}
		}
		decoder := xml.NewDecoder(bytes.NewReader(art.Data))
		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("invalid SVG: %v", err)
			}
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := generator.Generate(ctx, "seed"); !errors.Is(err, context.Canceled) {
		t.Fatalf("want canceled, got %v", err)
	}
}

func TestPicassoPNGExports(t *testing.T) {
	engine := &Engine{}
	if err := engine.RegisterGenerator(Picasso{}); err != nil {
		t.Fatal(err)
	}
	if err := engine.RegisterExporter(PNGExporter{}); err != nil {
		t.Fatal(err)
	}
	// Exercise varied clothing, hair, and color paths through the real rasterizer.
	for i := 0; i < 24; i++ {
		options := Options{Width: 128, Height: 144}
		if i%2 == 0 {
			options = Options{Width: 72, Height: 72, Circle: true}
		}
		data, err := engine.Render(context.Background(), "picasso", fmt.Sprintf("picasso-%d", i), "png", options)
		if err != nil {
			t.Fatalf("render %d: %v", i, err)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil || img.Bounds().Dx() != options.Width || img.Bounds().Dy() != options.Height {
			t.Fatalf("bad PNG dimensions for %d: %v", i, err)
		}
		_, _, _, center := img.At(options.Width/2, options.Height/2).RGBA()
		if center != 65535 {
			t.Fatal("portrait center should be opaque")
		}
		_, _, _, corner := img.At(0, 0).RGBA()
		if options.Circle && corner != 0 {
			t.Fatal("circular portrait must have transparent corners")
		}
	}
}
