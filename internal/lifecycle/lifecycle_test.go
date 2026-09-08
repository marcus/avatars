package lifecycle

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func unixClient(path string) *http.Client {
	return &http.Client{Transport: &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", path)
	}}}
}

func waitSocket(t *testing.T, path string) {
	t.Helper()
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); {
		if _, err := os.Stat(path); err == nil {
			if conn, dialErr := net.DialTimeout("unix", path, 20*time.Millisecond); dialErr == nil {
				conn.Close()
				return
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("lifecycle socket was not created")
}

func socketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "av-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return filepath.Join(dir, "s.sock")
}

func TestControllerRestrictsSocketAndPinsShutdownIdentity(t *testing.T) {
	path := socketPath(t)
	status := NewStatus(Auto, "http://127.0.0.1:1234", "/tmp/data")
	c := NewController(status)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- c.Serve(ctx, path) }()
	waitSocket(t, path)
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("socket mode=%v err=%v", info.Mode().Perm(), err)
	}
	body, _ := json.Marshal(map[string]string{"instance_id": "wrong"})
	res, err := unixClient(path).Post("http://unix/v1/shutdown", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("status=%d", res.StatusCode)
	}
	select {
	case <-c.Done():
		t.Fatal("stale identity stopped service")
	default:
	}
	body, _ = json.Marshal(map[string]string{"instance_id": status.InstanceID})
	res, err = unixClient(path).Post("http://unix/v1/shutdown", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal("accepted shutdown must return its response before closing:", err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		t.Fatalf("status=%d", res.StatusCode)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("controller did not stop")
	}
}

func TestControllerDoesNotUnlinkLiveOwner(t *testing.T) {
	path := socketPath(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	first := NewController(NewStatus(Auto, "http://127.0.0.1:1234", ""))
	done := make(chan error, 1)
	go func() { done <- first.Serve(ctx, path) }()
	waitSocket(t, path)
	second := NewController(NewStatus(Foreground, "http://127.0.0.1:1235", ""))
	if err := second.Serve(context.Background(), path); err == nil {
		t.Fatal("second owner replaced live lifecycle socket")
	}
	res, err := unixClient(path).Get("http://unix/v1/status")
	if err != nil {
		t.Fatal("first owner no longer reachable:", err)
	}
	res.Body.Close()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("first controller did not stop")
	}
}

func TestForegroundControllerRefusesLifecycleShutdown(t *testing.T) {
	path := socketPath(t)
	status := NewStatus(Foreground, "http://127.0.0.1:1234", "")
	c := NewController(status)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- c.Serve(ctx, path) }()
	waitSocket(t, path)
	body, _ := json.Marshal(map[string]string{"instance_id": status.InstanceID})
	res, err := unixClient(path).Post("http://unix/v1/shutdown", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusConflict {
		t.Fatalf("status=%d", res.StatusCode)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("controller did not stop")
	}
}
