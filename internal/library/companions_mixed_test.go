package library_test

import (
	"bytes"
	"context"
	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
	"github.com/marcus/avatars/pkg/avatar"
	"path/filepath"
	"testing"
)

func TestMixedCompanionBatchPersistsAndReopensConcreteRecipes(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "library.jsonl")
	storage, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	engine := avatar.New()
	service := library.New(engine, storage)
	req := library.CreateRequest{Style: "companions", Seed: "mixed-batch", Count: 12, Inputs: &avatar.Inputs{Animal: "mixed"}}
	col, err := service.Create(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	restarted := library.New(avatar.New(), reopened)
	seen := map[string]bool{}
	for _, item := range col.Avatars {
		resolved, err := engine.ResolveRecipeInputs(item.Style, item.Seed, *req.Inputs)
		if err != nil || item.Inputs == nil || *item.Inputs != resolved || item.Inputs.Animal == "mixed" {
			t.Fatalf("bad saved recipe %+v %v", item, err)
		}
		seen[item.Inputs.Animal] = true
		loaded, err := restarted.Avatar(ctx, item.ID)
		if err != nil || loaded.Inputs == nil || *loaded.Inputs != resolved {
			t.Fatal("restart lost concrete species", err)
		}
		for _, format := range []string{"svg", "png"} {
			opts := avatar.Options{Width: 64, Height: 64, Circle: true}
			expected, err := engine.RenderWithInputs(ctx, item.Style, item.Seed, format, *req.Inputs, opts)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := restarted.Render(ctx, item.ID, format, opts)
			if err != nil || !bytes.Equal(actual, expected) {
				t.Fatal("saved export changed species", format, err)
			}
		}
	}
	if len(seen) != 2 || req.Inputs.Animal != "mixed" {
		t.Fatal("batch did not retain concrete species", seen)
	}
	for _, invalid := range []library.CreateRequest{
		{Style: "companions", Inputs: &avatar.Inputs{Animal: "Mixed"}},
		{Style: "companions", Inputs: &avatar.Inputs{Animal: "mixed", Color: "random"}},
		{Style: "pebble", Inputs: &avatar.Inputs{Animal: "mixed"}},
	} {
		if _, err := restarted.Create(ctx, invalid); err == nil {
			t.Fatal("accepted invalid Mixed request")
		}
	}
	cols, err := restarted.List(ctx)
	if err != nil || len(cols) != 1 {
		t.Fatal("invalid input wrote a record", err)
	}
}
