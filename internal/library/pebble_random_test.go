package library_test

import (
	"context"
	"testing"

	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/pkg/avatar"
)

func TestRandomBatchPersistsConcretePerAvatarColors(t *testing.T) {
	ctx := context.Background()
	storage := &memoryStore{}
	engine := avatar.New()
	service := library.New(engine, storage)
	requested := &avatar.Inputs{Color: "random"}
	collection, err := service.Create(ctx, library.CreateRequest{Style: "pebble", Seed: "batch-proof", Count: 100, Inputs: requested})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, item := range collection.Avatars {
		want, err := engine.ResolveRecipeInputs("pebble", item.Seed, *requested)
		if err != nil || item.Inputs == nil || *item.Inputs != want || item.Inputs.Color == "random" {
			t.Fatalf("invalid persisted recipe: %+v, %v", item, err)
		}
		seen[item.Inputs.Color] = true
	}
	if storage.saves != 1 || len(seen) < 10 || requested.Color != "random" {
		t.Fatalf("batch resolution: %d saves, %d colors, request %+v", storage.saves, len(seen), requested)
	}
	for _, req := range []library.CreateRequest{
		{Style: "pebble", Inputs: &avatar.Inputs{Color: "random", Animal: "cat"}},
		{Style: "companions", Inputs: &avatar.Inputs{Color: "random", Animal: "dog"}},
		{Style: "pebble", Inputs: &avatar.Inputs{Color: "Random"}},
	} {
		if _, err := service.Create(ctx, req); err == nil {
			t.Fatalf("accepted invalid input %+v", req)
		}
	}
	if storage.saves != 1 {
		t.Fatal("invalid request wrote to the store")
	}
}
