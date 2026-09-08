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
