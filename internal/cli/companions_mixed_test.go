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

func TestMixedCompanionLocalHTTPAndSavedParity(t *testing.T) {
	cleanEnv(t)
	storage, err := store.New(filepath.Join(t.TempDir(), "library.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	engine := avatar.New()
	server := httptest.NewServer(httpapi.New(library.New(engine, storage), engine, studio.Handler()))
	defer server.Close()
	code, data, stderr := invoke(t, "generate", "--url", server.URL, "--style", "companions", "--animal", "mixed", "--seed", "mixed-cli", "--count", "4", "--json")
	if code != 0 {
		t.Fatal(stderr)
	}
	var col library.Collection
	if err := json.Unmarshal([]byte(data), &col); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, item := range col.Avatars {
		if item.Inputs == nil || (item.Inputs.Animal != "dog" && item.Inputs.Animal != "cat") {
			t.Fatal("unresolved species", item.Inputs)
		}
		seen[item.Inputs.Animal] = true
		for _, format := range []string{"svg", "png"} {
			code, local, stderr := invoke(t, "render", "--style", "companions", "--animal", "mixed", "--seed", item.Seed, "--format", format)
			if code != 0 {
				t.Fatal(stderr)
			}
			code, remote, stderr := invoke(t, "render", "--url", server.URL, "--style", "companions", "--animal", "mixed", "--seed", item.Seed, "--format", format)
			if code != 0 || local != remote {
				t.Fatal("local/HTTP Mixed mismatch", format, stderr)
			}
			code, saved, stderr := invoke(t, "export", item.ID, "--url", server.URL, "--format", format)
			if code != 0 || local != saved {
				t.Fatal("saved Mixed mismatch", format, stderr)
			}
		}
	}
	if len(seen) != 2 {
		t.Fatal("seeded CLI batch did not reach both species")
	}
	code, _, _ = invoke(t, "export", col.Avatars[0].ID, "--url", server.URL, "--animal", "mixed", "--json")
	if code != 2 {
		t.Fatal("saved export accepted Mixed override")
	}
}
