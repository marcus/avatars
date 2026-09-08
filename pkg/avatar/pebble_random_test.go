package avatar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestPebbleRandomResolvesFullPaletteWithoutChangingArtwork(t *testing.T) {
	e := New()
	seen := map[string]bool{}
	request := Inputs{Color: "random"}
	for i := 0; i < 256; i++ {
		seed := fmt.Sprintf("random-proof:%d", i)
		resolved, err := e.ResolveRecipeInputs("pebble", seed, request)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := pebbleColor(resolved.Color); !ok {
			t.Fatalf("Random did not resolve a concrete color: %+v", resolved)
		}
		seen[resolved.Color] = true
		repeated, err := e.ResolveRecipeInputs("pebble", seed, request)
		if err != nil || repeated != resolved {
			t.Fatalf("resolution is not deterministic: %+v, %v", repeated, err)
		}
		if i >= 16 {
			continue
		}
		for _, format := range []string{"svg", "png"} {
			random, err := e.RenderWithInputs(context.Background(), "pebble", seed, format, request, Options{})
			if err != nil {
				t.Fatal(err)
			}
			concrete, err := e.RenderWithInputs(context.Background(), "pebble", seed, format, resolved, Options{})
			if err != nil || !bytes.Equal(random, concrete) {
				t.Fatalf("Random changed %s artwork for seed %q: %v", format, seed, err)
			}
		}
		direct, err := (Pebble{}).GenerateWithInputs(context.Background(), seed, request)
		if err != nil {
			t.Fatal(err)
		}
		concrete, err := (Pebble{}).GenerateWithInputs(context.Background(), seed, resolved)
		if err != nil || !bytes.Equal(direct.Data, concrete.Data) {
			t.Fatal("direct generator disagrees with recipe resolution", err)
		}
	}
	if len(seen) != 15 {
		t.Fatalf("fixed sample covered %d of 15 colors: %v", len(seen), seen)
	}
	if request.Color != "random" {
		t.Fatal("resolution changed the request")
	}
	for _, style := range []string{"gorey", "picasso", "companions"} {
		if _, err := e.ResolveRecipeInputs(style, "seed", request); !errors.Is(err, ErrInvalidInputs) {
			t.Fatalf("%s accepted Random color: %v", style, err)
		}
	}
	if _, err := e.ResolveRecipeInputs("pebble", "seed", Inputs{Color: "random", Animal: "cat"}); !errors.Is(err, ErrInvalidInputs) {
		t.Fatal("Pebble accepted animal input", err)
	}
}
