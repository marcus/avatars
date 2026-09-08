package library_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
	"github.com/marcus/avatars/pkg/avatar"
)

type memoryStore struct {
	collections []library.Collection
	saves       int
	err         error
}

func (m *memoryStore) Save(_ context.Context, collection library.Collection) error {
	m.saves++
	if m.err != nil {
		return m.err
	}
	m.collections = append(m.collections, collection)
	return nil
}

func (m *memoryStore) List(context.Context) ([]library.Collection, error) {
	return append([]library.Collection(nil), m.collections...), m.err
}

func TestCreateRestartAndRender(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "library.jsonl")
	persistent, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	service := library.New(avatar.New(), persistent)
	collection, err := service.Create(ctx, library.CreateRequest{Name: "My portraits", Count: 3})
	if err != nil {
		t.Fatal(err)
	}
	if collection.Name != "My portraits" || collection.Style != "gorey" || len(collection.Avatars) != 3 {
		t.Fatalf("unexpected collection: %+v", collection)
	}
	seeds := make(map[string]bool)
	for _, item := range collection.Avatars {
		if seeds[item.Seed] || len(item.Seed) != 32 {
			t.Fatalf("random seed missing or reused: %q", item.Seed)
		}
		seeds[item.Seed] = true
		if item.CollectionID != collection.ID || item.URL != "/?avatar="+item.ID || item.SVGURL != "/api/v1/avatars/"+item.ID+".svg" || item.PNGURL != "/api/v1/avatars/"+item.ID+".png" {
			t.Fatalf("broken avatar links: %+v", item)
		}
	}
	first := collection.Avatars[0]
	before, err := service.Render(ctx, first.ID, "svg", avatar.Options{})
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	restarted := library.New(avatar.New(), reopened)
	loaded, err := restarted.Collection(ctx, collection.ID)
	if err != nil || loaded.Name != collection.Name || len(loaded.Avatars) != 3 {
		t.Fatalf("collection did not survive restart: %+v, %v", loaded, err)
	}
	after, err := restarted.Render(ctx, first.ID, "svg", avatar.Options{})
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("render changed after restart: %v", err)
	}
	if _, err := restarted.Create(ctx, library.CreateRequest{}); err != nil {
		t.Fatal(err)
	}
	visible, err := service.List(ctx)
	if err != nil || len(visible) != 2 {
		t.Fatalf("live service did not see the other store's write: %d, %v", len(visible), err)
	}
}

func TestSeedReproduction(t *testing.T) {
	ctx := context.Background()
	service := library.New(avatar.New(), &memoryStore{})
	seed := " portrait <seed> 日本語 "
	single, err := service.Create(ctx, library.CreateRequest{Seed: seed})
	if err != nil {
		t.Fatal(err)
	}
	if single.Avatars[0].Seed != seed {
		t.Fatalf("seed was changed: %q", single.Avatars[0].Seed)
	}
	one, err := service.Create(ctx, library.CreateRequest{Seed: seed, Count: 3})
	if err != nil {
		t.Fatal(err)
	}
	two, err := service.Create(ctx, library.CreateRequest{Seed: seed, Count: 3})
	if err != nil {
		t.Fatal(err)
	}
	if one.ID == two.ID {
		t.Fatal("reproduction should create a separately shareable collection")
	}
	for i := range one.Avatars {
		want := fmt.Sprintf("%s:%d", seed, i)
		if one.Avatars[i].Seed != want || two.Avatars[i].Seed != want || one.Avatars[i].ID == two.Avatars[i].ID {
			t.Fatalf("incorrect deterministic batch at index %d", i)
		}
		a, err := service.Render(ctx, one.Avatars[i].ID, "svg", avatar.Options{})
		if err != nil {
			t.Fatal(err)
		}
		b, err := service.Render(ctx, two.Avatars[i].ID, "svg", avatar.Options{})
		if err != nil || !bytes.Equal(a, b) {
			t.Fatalf("batch reproduction failed at index %d: %v", i, err)
		}
	}
}

func TestInvalidRequestsNeverSave(t *testing.T) {
	for _, test := range []struct {
		name    string
		request library.CreateRequest
	}{
		{"negative count", library.CreateRequest{Count: -1}},
		{"too many avatars", library.CreateRequest{Count: 101}},
		{"unknown style", library.CreateRequest{Style: "missing"}},
		{"long name", library.CreateRequest{Name: strings.Repeat("界", 121)}},
		{"invalid name", library.CreateRequest{Name: "\xff"}},
		{"long seed", library.CreateRequest{Seed: strings.Repeat("s", 4097)}},
		{"invalid seed", library.CreateRequest{Seed: "\xff"}},
		{"invalid Pebble color", library.CreateRequest{Style: "pebble", Inputs: &avatar.Inputs{Color: "missing"}}},
		{"color unsupported by style", library.CreateRequest{Style: "gorey", Inputs: &avatar.Inputs{Color: "sage"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			storage := &memoryStore{}
			service := library.New(avatar.New(), storage)
			_, err := service.Create(context.Background(), test.request)
			var failure *library.Error
			if !errors.As(err, &failure) || failure.Code != "invalid_request" {
				t.Fatalf("want invalid_request, got %v", err)
			}
			if storage.saves != 0 {
				t.Fatal("invalid request reached persistence")
			}
		})
	}
}

func TestPebbleInputsPersistAndOldRecipesUseDefault(t *testing.T) {
	ctx := context.Background()
	storage := &memoryStore{}
	service := library.New(avatar.New(), storage)
	collection, err := service.Create(ctx, library.CreateRequest{Style: "pebble", Seed: "saved", Count: 2, Inputs: &avatar.Inputs{Color: "sage"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range collection.Avatars {
		if item.Inputs == nil || item.Inputs.Color != "sage" {
			t.Fatalf("resolved color missing from saved recipe: %+v", item)
		}
	}
	defaults, err := service.Create(ctx, library.CreateRequest{Style: "pebble", Seed: "default"})
	if err != nil || defaults.Avatars[0].Inputs == nil || defaults.Avatars[0].Inputs.Color != "walnut" {
		t.Fatalf("Pebble default was not persisted: %+v, %v", defaults, err)
	}

	now := time.Now().UTC()
	old := library.Avatar{ID: "av_old", CollectionID: "col_old", Style: "pebble", Seed: "legacy", CreatedAt: now}
	storage.collections = append(storage.collections, library.Collection{ID: "col_old", Style: "pebble", CreatedAt: now, Avatars: []library.Avatar{old}})
	fromOldRecipe, err := service.Render(ctx, old.ID, "svg", avatar.Options{})
	if err != nil {
		t.Fatal(err)
	}
	explicitDefault, err := avatar.New().RenderWithInputs(ctx, "pebble", old.Seed, "svg", avatar.Inputs{Color: "walnut"}, avatar.Options{})
	if err != nil || !bytes.Equal(fromOldRecipe, explicitDefault) {
		t.Fatal("old input-free Pebble recipe did not resolve to Walnut", err)
	}
}

func TestPebbleRecipeSurvivesStoreRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "library.jsonl")
	persistent, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	service := library.New(avatar.New(), persistent)
	collection, err := service.Create(ctx, library.CreateRequest{Style: "pebble", Seed: "restart", Inputs: &avatar.Inputs{Color: "lavender"}})
	if err != nil {
		t.Fatal(err)
	}
	before, err := service.Render(ctx, collection.Avatars[0].ID, "png", avatar.Options{Width: 96, Height: 96})
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	restarted := library.New(avatar.New(), reopened)
	loaded, err := restarted.Avatar(ctx, collection.Avatars[0].ID)
	if err != nil || loaded.Inputs == nil || loaded.Inputs.Color != "lavender" {
		t.Fatalf("saved inputs did not survive restart: %+v, %v", loaded, err)
	}
	after, err := restarted.Render(ctx, loaded.ID, "png", avatar.Options{Width: 96, Height: 96})
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("appearance changed after restart", err)
	}
}

func TestBatchSavesOnceAndListsNewestFirst(t *testing.T) {
	storage := &memoryStore{}
	service := library.New(avatar.New(), storage)
	collection, err := service.Create(context.Background(), library.CreateRequest{Count: 100, Name: strings.Repeat("界", 120)})
	if err != nil {
		t.Fatal(err)
	}
	if storage.saves != 1 || len(collection.Avatars) != 100 {
		t.Fatalf("batch must save in one operation: %d saves, %d avatars", storage.saves, len(collection.Avatars))
	}
	storage.collections = []library.Collection{
		{ID: "old", CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "new", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	list, err := service.List(context.Background())
	if err != nil || list[0].ID != "new" {
		t.Fatalf("want newest first: %+v, %v", list, err)
	}
}

func TestMissingRecordsAndStoreFailure(t *testing.T) {
	ctx := context.Background()
	storage := &memoryStore{}
	service := library.New(avatar.New(), storage)
	_, avatarErr := service.Avatar(ctx, "unknown")
	_, collectionErr := service.Collection(ctx, "unknown")
	for _, err := range []error{avatarErr, collectionErr} {
		var failure *library.Error
		if !errors.As(err, &failure) || failure.Code != "not_found" {
			t.Fatalf("want not_found, got %v", err)
		}
	}
	storage.err = errors.New("disk is full")
	collection, err := service.Create(ctx, library.CreateRequest{})
	var failure *library.Error
	if collection.ID != "" || !errors.As(err, &failure) || failure.Code != "internal" {
		t.Fatalf("failed save appeared successful: %+v, %v", collection, err)
	}
}
