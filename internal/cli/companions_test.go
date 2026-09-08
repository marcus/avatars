package cli_test

import (
	"encoding/json"
	"github.com/marcus/avatars/internal/httpapi"
	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
	"github.com/marcus/avatars/internal/studio"
	"github.com/marcus/avatars/pkg/avatar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompanionAnimalCLIParityAndRefusals(t *testing.T) {
	cleanEnv(t)
	dir := t.TempDir()
	s, err := store.New(filepath.Join(dir, "collections.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	engine := avatar.New()
	server := httptest.NewServer(httpapi.New(library.New(engine, s), engine, studio.Handler()))
	defer server.Close()

	code, out, stderr := invoke(t, "generate", "--url", server.URL, "--style", "companions", "--animal", "cat", "--seed", "cli-companion", "--json")
	if code != 0 {
		t.Fatal(stderr)
	}
	var collection library.Collection
	if err := json.Unmarshal([]byte(out), &collection); err != nil {
		t.Fatal(err)
	}
	if collection.Avatars[0].Inputs == nil || collection.Avatars[0].Inputs.Animal != "cat" {
		t.Fatalf("CLI response lost saved animal: %s", out)
	}
	code, saved, stderr := invoke(t, "export", collection.Avatars[0].ID, "--url", server.URL)
	if code != 0 {
		t.Fatal(stderr)
	}
	code, local, stderr := invoke(t, "render", "--style", "companions", "--animal", "cat", "--seed", "cli-companion")
	if code != 0 || local != saved {
		t.Fatal("local render and remote saved export differ", code, stderr)
	}
	code, remote, stderr := invoke(t, "render", "--url", server.URL, "--style", "companions", "--animal", "cat", "--seed", "cli-companion")
	if code != 0 || remote != saved {
		t.Fatal("remote render and saved export differ", code, stderr)
	}
	for _, args := range [][]string{
		{"generate", "--data-dir", t.TempDir(), "--style", "gorey", "--animal", "cat", "--json"},
		{"render", "--style", "companions", "--animal", "missing", "--seed", "x", "--json"},
		{"export", collection.Avatars[0].ID, "--url", server.URL, "--animal", "dog", "--json"},
	} {
		code, _, stderr := invoke(t, args...)
		if code != 2 || !strings.Contains(stderr, `"code":"invalid_request"`) {
			t.Fatalf("accepted invalid CLI request %v: %d %s", args, code, stderr)
		}
	}
}
