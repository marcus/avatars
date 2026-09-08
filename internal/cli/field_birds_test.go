package cli_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcus/avatars/internal/httpapi"
	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
	"github.com/marcus/avatars/internal/studio"
	"github.com/marcus/avatars/pkg/avatar"
)

func TestFieldBirdCollectionJourney(t *testing.T) {
	cleanEnv(t)
	path := filepath.Join(t.TempDir(), "library.jsonl")
	storage, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	engine := avatar.New()
	service := library.New(engine, storage)
	server := httptest.NewServer(httpapi.New(service, engine, studio.Handler()))
	defer server.Close()
	code, data, stderr := invoke(t, "styles", "--url", server.URL, "--json")
	if code != 0 || !strings.Contains(data, `"field-birds"`) {
		t.Fatal("style discovery", stderr)
	}
	code, data, stderr = invoke(t, "generate", "--url", server.URL, "--style", "field-birds", "--seed", "field-notes", "--count", "8", "--json")
	if code != 0 {
		t.Fatal(stderr)
	}
	var col library.Collection
	if err := json.Unmarshal([]byte(data), &col); err != nil {
		t.Fatal(err)
	}
	if len(col.Avatars) != 8 {
		t.Fatal("wrong count")
	}
	reopened, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	restarted := library.New(avatar.New(), reopened)
	for _, item := range col.Avatars {
		loaded, err := restarted.Avatar(context.Background(), item.ID)
		if err != nil || loaded.Seed != item.Seed || loaded.Style != "field-birds" {
			t.Fatal("restart lost recipe", err)
		}
		for _, format := range []string{"svg", "png"} {
			code, local, stderr := invoke(t, "render", "--style", "field-birds", "--seed", item.Seed, "--format", format, "--size", "64", "--circle")
			if code != 0 {
				t.Fatal(stderr)
			}
			code, remote, stderr := invoke(t, "render", "--url", server.URL, "--style", "field-birds", "--seed", item.Seed, "--format", format, "--size", "64", "--circle")
			if code != 0 || local != remote {
				t.Fatal("local/HTTP mismatch", format, stderr)
			}
			code, saved, stderr := invoke(t, "export", item.ID, "--url", server.URL, "--format", format, "--size", "64", "--circle")
			if code != 0 || local != saved {
				t.Fatal("saved mismatch", format, stderr)
			}
			after, err := restarted.Render(context.Background(), item.ID, format, avatar.Options{Width: 64, Height: 64, Circle: true})
			if err != nil || string(after) != local {
				t.Fatal("restart export mismatch", err)
			}
		}
	}
	code, _, _ = invoke(t, "generate", "--url", server.URL, "--style", "field-birds", "--color", "sage", "--json")
	if code == 0 {
		t.Fatal("accepted unsupported color")
	}
	cols, err := service.List(context.Background())
	if err != nil || len(cols) != 1 {
		t.Fatal("rejected input wrote collection", err)
	}
}
