package cli_test

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/marcus/avatars/internal/httpapi"
	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
	"github.com/marcus/avatars/internal/studio"
	"github.com/marcus/avatars/pkg/avatar"
)

func TestRandomColorLocalRemoteAndSavedParity(t *testing.T) {
	cleanEnv(t)
	storage, err := store.New(filepath.Join(t.TempDir(), "collections.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	engine := avatar.New()
	server := httptest.NewServer(httpapi.New(library.New(engine, storage), engine, studio.Handler()))
	defer server.Close()
	code, output, stderr := invoke(t, "generate", "--url", server.URL, "--style", "pebble", "--color", "random", "--seed", "random-parity", "--json")
	if code != 0 {
		t.Fatal(stderr)
	}
	var collection library.Collection
	if err := json.Unmarshal([]byte(output), &collection); err != nil {
		t.Fatal(err)
	}
	item := collection.Avatars[0]
	if item.Inputs == nil || item.Inputs.Color == "random" || item.Inputs.Color == "" {
		t.Fatalf("unresolved saved color: %+v", item)
	}
	for _, format := range []string{"svg", "png"} {
		code, local, stderr := invoke(t, "render", "--style", "pebble", "--color", "random", "--seed", "random-parity", "--format", format, "--size", "64", "--circle")
		if code != 0 {
			t.Fatal(stderr)
		}
		code, remote, stderr := invoke(t, "render", "--url", server.URL, "--style", "pebble", "--color", "random", "--seed", "random-parity", "--format", format, "--size", "64", "--circle")
		if code != 0 || local != remote {
			t.Fatal("local/HTTP random rendering differs", format, stderr)
		}
		code, saved, stderr := invoke(t, "export", item.ID, "--url", server.URL, "--format", format, "--size", "64", "--circle")
		if code != 0 || local != saved {
			t.Fatal("saved random export differs", format, stderr)
		}
	}
	code, _, stderr = invoke(t, "export", item.ID, "--url", server.URL, "--color", "random", "--json")
	if code != 2 {
		t.Fatal("saved export accepted Random override", code, stderr)
	}
}
