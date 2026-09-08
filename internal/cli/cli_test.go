package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image/png"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcus/avatars/internal/cli"
	"github.com/marcus/avatars/internal/httpapi"
	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
	"github.com/marcus/avatars/internal/studio"
	"github.com/marcus/avatars/pkg/avatar"
)

func invoke(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, err bytes.Buffer
	code := cli.Run(context.Background(), args, &out, &err)
	return code, out.String(), err.String()
}
func cleanEnv(t *testing.T) {
	t.Helper()
	t.Setenv("AVATARS_DATA_DIR", "")
	t.Setenv("AVATARS_URL", "")
}
func TestSavedCLIHTTPStudioJourney(t *testing.T) {
	cleanEnv(t)
	dir := t.TempDir()
	s, e := store.New(filepath.Join(dir, "collections.jsonl"))
	if e != nil {
		t.Fatal(e)
	}
	engine := avatar.New()
	server := httptest.NewServer(httpapi.New(library.New(engine, s), engine, studio.Handler()))
	defer server.Close()
	code, out, err := invoke(t, "generate", "--data-dir", dir, "--count", "4", "--seed", "stable", "--name", "Review", "--json")
	if code != 0 {
		t.Fatal(err)
	}
	var c library.Collection
	if e := json.Unmarshal([]byte(out), &c); e != nil {
		t.Fatal(e)
	}
	if len(c.Avatars) != 4 || !strings.HasPrefix(c.URL, "http://127.0.0.1:7447/?collection=") {
		t.Fatal(out)
	}
	code, out, err = invoke(t, "--url", server.URL, "list", "--json")
	if code != 0 || !strings.Contains(out, c.ID) {
		t.Fatal(code, out, err)
	}
	code, out, err = invoke(t, "generate", "--url", server.URL, "--seed", "remote", "--json")
	if code != 0 {
		t.Fatal(err)
	}
	var remote library.Collection
	_ = json.Unmarshal([]byte(out), &remote)
	code, out, err = invoke(t, "show", remote.ID, "--data-dir", dir, "--json")
	if code != 0 || !strings.Contains(out, remote.ID) {
		t.Fatal(code, out, err)
	}
	file := filepath.Join(t.TempDir(), "icon.png")
	code, out, err = invoke(t, "export", c.Avatars[0].ID, "--url", server.URL, "--format", "png", "--size", "96", "--circle", "--out", file, "--json")
	if code != 0 {
		t.Fatal(err)
	}
	f, e := os.Open(file)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	im, e := png.Decode(f)
	if e != nil || im.Bounds().Dx() != 96 {
		t.Fatal(e)
	}
	_, _, _, alpha := im.At(0, 0).RGBA()
	if alpha != 0 {
		t.Fatal("circle corners must be transparent")
	}
	code, _, err = invoke(t, "export", c.Avatars[0].ID, "--data-dir", dir, "--out", file)
	if code == 0 {
		t.Fatal("overwrote output", err)
	}
	code, out, err = invoke(t, "render", "stable:0", "--format", "svg")
	if code != 0 {
		t.Fatal(err)
	}
	code, saved, err := invoke(t, "export", c.Avatars[0].ID, "--data-dir", dir)
	if code != 0 || saved != out {
		t.Fatal("saved rendering differs from stateless rendering", err)
	}
}
func TestCLIHelpAndRefusals(t *testing.T) {
	cleanEnv(t)
	for _, args := range [][]string{{"--help"}, {"help", "generate"}, {"generate", "--help"}, {"--json", "capabilities"}, {"instructions"}} {
		code, out, err := invoke(t, args...)
		if code != 0 || out == "" {
			t.Fatal(args, code, err)
		}
	}
	for _, args := range [][]string{{"wat"}, {"generate", "--count", "0"}, {"render"}, {"render", "s", "--size", "-2"}, {"serve", "--listen", "0.0.0.0:7447"}, {"generate", "--typo"}, {"--url", "bogus", "list"}, {"--url", "http://localhost:7447", "--data-dir", "/tmp/a", "list"}} {
		code, _, err := invoke(t, append(args, "--json")...)
		var envelope map[string]any
		if code != 2 || json.Unmarshal([]byte(err), &envelope) != nil {
			t.Fatal(args, code, err)
		}
	}
	code, out, err := invoke(t, "render", "--seed", "--json")
	if code != 0 || !strings.HasPrefix(out, "<svg") {
		t.Fatal("global flag consumed seed", code, err)
	}
	code, out, err = invoke(t, "render", "--", "--url")
	if code != 0 || !strings.HasPrefix(out, "<svg") {
		t.Fatal("literal seed failed", code, err)
	}
}

func TestLiteralHelpSeeds(t *testing.T) {
	cleanEnv(t)
	for _, args := range [][]string{{"render", "--", "--help"}, {"render", "--seed", "--help"}, {"render", "--seed", "-h"}} {
		code, out, err := invoke(t, args...)
		if code != 0 || !strings.HasPrefix(out, "<svg") {
			t.Fatal(args, code, err, out)
		}
	}
}

func TestFailedExportReportsSavedCollection(t *testing.T) {
	cleanEnv(t)
	dir := t.TempDir()
	file := filepath.Join(dir, "missing", "avatar.png")
	code, _, err := invoke(t, "generate", "--data-dir", dir, "--out", file, "--json")
	if code != 1 || !strings.Contains(err, "was saved; export failed") {
		t.Fatal(code, err)
	}
	code, out, err := invoke(t, "list", "--data-dir", dir, "--json")
	if code != 0 || !strings.Contains(out, "col_") {
		t.Fatal(code, out, err)
	}
}

func TestStylesGeneralizeAcrossCLIAndHTTP(t *testing.T) {
	cleanEnv(t)
	s, e := store.New(filepath.Join(t.TempDir(), "collections.jsonl"))
	if e != nil {
		t.Fatal(e)
	}
	engine := avatar.New()
	server := httptest.NewServer(httpapi.New(library.New(engine, s), engine, studio.Handler()))
	defer server.Close()
	code, out, err := invoke(t, "styles", "--url", server.URL, "--json")
	if code != 0 {
		t.Fatal(err)
	}
	for _, style := range []string{"gorey", "gorey-expanded", "pebble", "picasso"} {
		if !strings.Contains(out, `"id": "`+style+`"`) {
			t.Fatal("missing style", style, out)
		}
		code, data, err := invoke(t, "generate", "--url", server.URL, "--style", style, "--seed", "integration", "--json")
		if code != 0 {
			t.Fatal(style, err)
		}
		var c library.Collection
		if e := json.Unmarshal([]byte(data), &c); e != nil {
			t.Fatal(e)
		}
		if c.Style != style || len(c.Avatars) != 1 || c.Avatars[0].Style != style {
			t.Fatal(c)
		}
		code, saved, err := invoke(t, "export", c.Avatars[0].ID, "--url", server.URL)
		if code != 0 {
			t.Fatal(err)
		}
		code, rendered, err := invoke(t, "render", "--style", style, "--seed", "integration")
		if code != 0 || saved != rendered {
			t.Fatal("cross-surface rendering mismatch", style, err)
		}
	}
}

func TestPebbleColorCLIParityAndRefusals(t *testing.T) {
	cleanEnv(t)
	dir := t.TempDir()
	s, err := store.New(filepath.Join(dir, "collections.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	engine := avatar.New()
	server := httptest.NewServer(httpapi.New(library.New(engine, s), engine, studio.Handler()))
	defer server.Close()

	code, out, stderr := invoke(t, "generate", "--url", server.URL, "--style", "pebble", "--color", "sage", "--seed", "cli-pebble", "--json")
	if code != 0 {
		t.Fatal(stderr)
	}
	var collection library.Collection
	if err := json.Unmarshal([]byte(out), &collection); err != nil {
		t.Fatal(err)
	}
	if collection.Avatars[0].Inputs == nil || collection.Avatars[0].Inputs.Color != "sage" {
		t.Fatalf("CLI response lost saved color: %s", out)
	}
	code, saved, stderr := invoke(t, "export", collection.Avatars[0].ID, "--url", server.URL)
	if code != 0 {
		t.Fatal(stderr)
	}
	code, local, stderr := invoke(t, "render", "--style", "pebble", "--color", "sage", "--seed", "cli-pebble")
	if code != 0 || local != saved {
		t.Fatal("local render and remote saved export differ", code, stderr)
	}
	code, remote, stderr := invoke(t, "render", "--url", server.URL, "--style", "pebble", "--color", "sage", "--seed", "cli-pebble")
	if code != 0 || remote != saved {
		t.Fatal("remote render and saved export differ", code, stderr)
	}
	for _, args := range [][]string{
		{"generate", "--data-dir", t.TempDir(), "--style", "gorey", "--color", "sage", "--json"},
		{"render", "--style", "pebble", "--color", "missing", "--seed", "x", "--json"},
		{"export", collection.Avatars[0].ID, "--url", server.URL, "--color", "walnut", "--json"},
	} {
		code, _, stderr := invoke(t, args...)
		if code != 2 || !strings.Contains(stderr, `"code":"invalid_request"`) {
			t.Fatalf("accepted invalid CLI request %v: %d %s", args, code, stderr)
		}
	}
}
