package library_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
	"github.com/marcus/avatars/pkg/avatar"
)

func TestCompanionRecipesSurviveRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "library.jsonl")
	persistent, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	engine := avatar.New()
	service := library.New(engine, persistent)
	var saved []library.Avatar
	for _, animal := range []string{"", "dog", "cat"} {
		request := library.CreateRequest{Style: "companions", Seed: "restart-companion", Count: 2}
		if animal != "" {
			request.Inputs = &avatar.Inputs{Animal: animal}
		}
		collection, err := service.Create(ctx, request)
		if err != nil {
			t.Fatal(err)
		}
		want := animal
		if want == "" {
			want = "dog"
		}
		for _, item := range collection.Avatars {
			if item.Inputs == nil || *item.Inputs != (avatar.Inputs{Animal: want}) {
				t.Fatalf("wrong persisted input: %+v", item)
			}
			saved = append(saved, item)
		}
	}
	reopened, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	restarted := library.New(avatar.New(), reopened)
	for _, item := range saved {
		loaded, err := restarted.Avatar(ctx, item.ID)
		if err != nil || loaded.Inputs == nil || *loaded.Inputs != *item.Inputs {
			t.Fatalf("recipe changed after restart: %+v %v", loaded, err)
		}
		for _, format := range []string{"svg", "png"} {
			opts := avatar.Options{Width: 64, Height: 64, Circle: true}
			expected, err := engine.RenderWithInputs(ctx, item.Style, item.Seed, format, *item.Inputs, opts)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := restarted.Render(ctx, item.ID, format, opts)
			if err != nil || !bytes.Equal(actual, expected) {
				t.Fatalf("restarted %s differs from original recipe: %v", format, err)
			}
		}
	}
}
