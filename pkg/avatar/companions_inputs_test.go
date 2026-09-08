package avatar

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestCompanionInputContract(t *testing.T) {
	engine := New()
	style := (Companions{}).Style()
	if style.Inputs.Animal.Default != "dog" || len(style.Inputs.Animal.Values) != 2 || style.Inputs.Color != nil {
		t.Fatalf("unexpected metadata: %+v", style)
	}
	if style.Inputs.Animal.Values[0] != (AnimalChoice{Value: "dog", Label: "Dogs"}) || style.Inputs.Animal.Values[1] != (AnimalChoice{Value: "cat", Label: "Cats"}) {
		t.Fatal("invalid species descriptors")
	}
	for _, requested := range []Inputs{{}, {Animal: "dog"}, {Animal: "cat"}} {
		want := requested.Animal
		if want == "" {
			want = "dog"
		}
		resolved, err := engine.ResolveInputs("companions", requested)
		if err != nil || resolved != (Inputs{Animal: want}) {
			t.Fatalf("resolve %+v: %+v %v", requested, resolved, err)
		}
		direct, err := (Companions{}).GenerateWithInputs(context.Background(), "same", requested)
		if err != nil {
			t.Fatal(err)
		}
		rendered, err := engine.RenderWithInputs(context.Background(), "companions", "same", "svg", resolved, Options{})
		if err != nil || !bytes.Equal(direct.Data, rendered) {
			t.Fatal("direct and engine rendering differ", err)
		}
	}
	implicit, err := engine.Render(context.Background(), "companions", "same", "svg", Options{})
	if err != nil {
		t.Fatal(err)
	}
	explicit, err := engine.RenderWithInputs(context.Background(), "companions", "same", "svg", Inputs{Animal: "dog"}, Options{})
	if err != nil || !bytes.Equal(implicit, explicit) {
		t.Fatal("input-free path must preserve default dog", err)
	}
	for _, test := range []struct {
		style string
		input Inputs
	}{
		{"companions", Inputs{Animal: "fox"}}, {"companions", Inputs{Animal: "Cat"}},
		{"companions", Inputs{Color: "sage"}}, {"companions", Inputs{Animal: "cat", Color: "sage"}},
		{"pebble", Inputs{Animal: "cat"}}, {"pebble", Inputs{Animal: "cat", Color: "sage"}},
		{"gorey", Inputs{Animal: "dog"}}, {"picasso", Inputs{Animal: "cat"}},
	} {
		if _, err := engine.ResolveInputs(test.style, test.input); !errors.Is(err, ErrInvalidInputs) {
			t.Fatalf("accepted invalid %s %+v: %v", test.style, test.input, err)
		}
		var g InputGenerator
		if test.style == "companions" {
			g = Companions{}
		}
		if test.style == "pebble" {
			g = Pebble{}
		}
		if g != nil {
			if _, err := g.GenerateWithInputs(context.Background(), "same", test.input); !errors.Is(err, ErrInvalidInputs) {
				t.Fatalf("direct generator accepted %+v: %v", test.input, err)
			}
		}
	}
}

func TestBuiltInNativeSizeMetadata(t *testing.T) {
	engine := New()
	for _, g := range []Generator{Gorey{}, GoreyExpanded{}, Picasso{}, Pebble{}, Companions{}} {
		style := g.Style()
		art, err := g.Generate(context.Background(), "native-metadata")
		if err != nil {
			t.Fatal(err)
		}
		if style.NativeWidth != art.Width || style.NativeHeight != art.Height {
			t.Fatalf("%s discovery size differs from artwork", style.ID)
		}
		registered := false
		for _, s := range engine.Styles() {
			if s.ID == style.ID {
				registered = true
			}
		}
		if !registered {
			t.Fatalf("%s not registered", style.ID)
		}
	}
}
