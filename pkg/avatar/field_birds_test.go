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

func TestFieldBirds(t *testing.T) {
	ctx := context.Background()
	engine := New()
	seen, families, palettes := map[string]bool{}, map[int]bool{}, map[int]bool{}
	for i := 0; i < 100; i++ {
		seed := fmt.Sprintf("field-notes:%d", i)
		art, err := (FieldBirds{}).Generate(ctx, seed)
		if err != nil {
			t.Fatal(err)
		}
		repeat, err := engine.Render(ctx, "field-birds", seed, "svg", Options{})
		if err != nil || !bytes.Equal(art.Data, repeat) {
			t.Fatal("unstable direct/engine rendering", err)
		}
		if art.Width != 160 || art.Height != 160 {
			t.Fatal("native dimensions")
		}
		seen[string(art.Data)] = true
		_, family, palette := fieldBirdRecipe(seed)
		if i < 48 {
			families[family] = true
			palettes[palette] = true
		}
		decoder := xml.NewDecoder(bytes.NewReader(art.Data))
		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		if i < 48 {
			data, err := engine.Render(ctx, "field-birds", seed, "png", Options{Width: 64, Height: 64, Circle: i%2 == 0})
			if err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			if img.Bounds().Dx() != 64 || img.Bounds().Dy() != 64 {
				t.Fatal("PNG size")
			}
			_, _, _, center := img.At(32, 32).RGBA()
			_, _, _, corner := img.At(0, 0).RGBA()
			if center != 65535 || (i%2 == 0 && corner != 0) {
				t.Fatal("opacity/circle")
			}
		}
	}
	if len(seen) != 100 || len(families) != 8 || len(palettes) != 12 {
		t.Fatalf("variation: %d images, %d families, %d palettes", len(seen), len(families), len(palettes))
	}
	for _, seed := range []string{"", "鳥🪶", `<script onload="bad()">`} {
		data, err := engine.Render(ctx, "field-birds", seed, "svg", Options{})
		if err != nil {
			t.Fatal(err)
		}
		for _, token := range []string{"script", "onload", "NaN", "%!", "href=", "<image", "<text"} {
			if strings.Contains(string(data), token) {
				t.Fatal("unsafe artwork", token)
			}
		}
	}
	for _, input := range []Inputs{{Color: "sage"}, {Animal: "dog"}} {
		if _, err := engine.RenderWithInputs(ctx, "field-birds", "seed", "svg", input, Options{}); !errors.Is(err, ErrInvalidInputs) {
			t.Fatal("unsupported input", err)
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := (FieldBirds{}).Generate(canceled, "seed"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
