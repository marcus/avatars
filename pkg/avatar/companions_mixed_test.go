package avatar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestCompanionsMixedResolvesConcreteSpecies(t *testing.T) {
	engine := New()
	request := Inputs{Animal: "mixed"}
	seen := map[string]bool{}
	for i := 0; i < 64; i++ {
		seed := fmt.Sprintf("mixed-proof:%d", i)
		resolved, err := engine.ResolveRecipeInputs("companions", seed, request)
		if err != nil || resolved.Color != "" || (resolved.Animal != "dog" && resolved.Animal != "cat") {
			t.Fatalf("not a concrete species: %+v %v", resolved, err)
		}
		seen[resolved.Animal] = true
		again, err := engine.ResolveRecipeInputs("companions", seed, request)
		if err != nil || again != resolved {
			t.Fatal("unstable Mixed resolution", err)
		}
		if i >= 12 {
			continue
		}
		direct, err := (Companions{}).GenerateWithInputs(context.Background(), seed, request)
		if err != nil {
			t.Fatal(err)
		}
		concrete, err := (Companions{}).GenerateWithInputs(context.Background(), seed, resolved)
		if err != nil || !bytes.Equal(direct.Data, concrete.Data) {
			t.Fatal("direct generator changed the resolved artwork", err)
		}
		for _, format := range []string{"svg", "png"} {
			opts := Options{Width: 64, Height: 64, Circle: true}
			mixed, err := engine.RenderWithInputs(context.Background(), "companions", seed, format, request, opts)
			if err != nil {
				t.Fatal(err)
			}
			concrete, err := engine.RenderWithInputs(context.Background(), "companions", seed, format, resolved, opts)
			if err != nil || !bytes.Equal(mixed, concrete) {
				t.Fatal("Mixed and concrete render differ", format, err)
			}
		}
	}
	if len(seen) != 2 || request.Animal != "mixed" {
		t.Fatal("species coverage or mutated request", seen, request)
	}
	for _, test := range []struct {
		style string
		input Inputs
	}{
		{"companions", Inputs{Animal: "Mixed"}}, {"companions", Inputs{Animal: "mixed", Color: "random"}},
		{"pebble", Inputs{Animal: "mixed", Color: "random"}}, {"gorey", Inputs{Animal: "mixed"}},
	} {
		if _, err := engine.ResolveRecipeInputs(test.style, "seed", test.input); !errors.Is(err, ErrInvalidInputs) {
			t.Fatalf("accepted invalid request %+v: %v", test, err)
		}
	}
}
