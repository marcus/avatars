package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalEndpointRejectsNonLoopback(t *testing.T) {
	for _, address := range []string{"0.0.0.0:7447", "example.com:7447", "missing-port"} {
		if _, err := localEndpoint(address); err == nil {
			t.Fatalf("accepted %q", address)
		}
	}
	got, err := localEndpoint("localhost:8123")
	if err != nil || got != "http://127.0.0.1:8123" {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestLifecycleLockSerializesConcurrentEnsure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lifecycle.lock")
	first, err := acquireLifecycleLock(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if _, err = acquireLifecycleLock(ctx, path); err == nil {
		t.Fatal("second caller acquired held lifecycle lock")
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	first, err = acquireLifecycleLock(context.Background(), path)
	if err != nil {
		t.Fatal("lock was not reusable:", err)
	}
}

func TestProbeRejectsForeignAndMalformedHealth(t *testing.T) {
	for _, handler := range []http.Handler{
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "no", http.StatusForbidden) }),
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("not-json")) }),
	} {
		server := httptest.NewServer(handler)
		if _, err := probeService(context.Background(), server.URL); err == nil {
			server.Close()
			t.Fatal("accepted foreign health response")
		}
		server.Close()
	}
}
