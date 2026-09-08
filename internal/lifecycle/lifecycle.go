// Package lifecycle defines the local Avatars service identity and control plane.
// The process and locking shape is adapted from github.com/marcus/comms.
package lifecycle

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/marcus/avatars/internal/buildinfo"
)

type LaunchMode string

const (
	Auto       LaunchMode = "auto"
	Foreground LaunchMode = "foreground"
	Supervised LaunchMode = "supervised"
)

type Status struct {
	Service    string     `json:"service,omitempty"`
	APIVersion string     `json:"api_version,omitempty"`
	Version    string     `json:"version,omitempty"`
	Commit     string     `json:"commit,omitempty"`
	InstanceID string     `json:"instance_id,omitempty"`
	PID        int        `json:"pid,omitempty"`
	StartedAt  string     `json:"started_at,omitempty"`
	LaunchMode LaunchMode `json:"launch_mode,omitempty"`
	Endpoint   string     `json:"endpoint,omitempty"`
	DataDir    string     `json:"data_dir,omitempty"`
}

func NewStatus(mode LaunchMode, endpoint, dataDir string) Status {
	return Status{Service: "avatars", APIVersion: "v1", Version: buildinfo.Version, Commit: buildinfo.Commit,
		InstanceID: fmt.Sprintf("avatars-%d-%d", os.Getpid(), time.Now().UnixNano()), PID: os.Getpid(),
		StartedAt: time.Now().UTC().Format(time.RFC3339Nano), LaunchMode: mode, Endpoint: endpoint, DataDir: dataDir}
}

func Paths(instance string) (socket, lock, log string, err error) {
	base := os.Getenv("XDG_RUNTIME_DIR")
	if base == "" {
		base, err = os.UserCacheDir()
		if err != nil {
			return "", "", "", err
		}
	}
	sum := sha256.Sum256([]byte(instance))
	dir := filepath.Join(base, "avatars", fmt.Sprintf("%x", sum[:6]))
	return filepath.Join(dir, "service.sock"), filepath.Join(dir, "service.lifecycle.lock"), filepath.Join(dir, "service.log"), nil
}

type Controller struct {
	status Status
	once   sync.Once
	done   chan struct{}
}

func NewController(s Status) *Controller    { return &Controller{status: s, done: make(chan struct{})} }
func (c *Controller) Done() <-chan struct{} { return c.done }

func (c *Controller) Serve(ctx context.Context, path string) error {
	ln, err := listen(path)
	if err != nil {
		return err
	}
	return c.serve(ctx, path, ln)
}

// Start binds the control socket before returning, so HTTP readiness cannot
// race ahead of the restricted lifecycle control plane.
func (c *Controller) Start(ctx context.Context, path string) error {
	ln, err := listen(path)
	if err != nil {
		return err
	}
	go func() { _ = c.serve(ctx, path, ln) }()
	return nil
}

func listen(path string) (net.Listener, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if _, err := os.Lstat(path); err == nil {
		conn, dialErr := net.DialTimeout("unix", path, 200*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			return nil, fmt.Errorf("lifecycle socket already has a live owner: %s", path)
		}
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("remove stale lifecycle socket: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		ln.Close()
		return nil, err
	}
	return ln, nil
}

func (c *Controller) serve(ctx context.Context, path string, ln net.Listener) error {
	defer ln.Close()
	defer os.Remove(path)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/status", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, c.status) })
	mux.HandleFunc("POST /v1/shutdown", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			InstanceID string `json:"instance_id"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&body); err != nil {
			writeJSON(w, 400, map[string]any{"error": map[string]string{"code": "invalid_request", "message": "invalid shutdown request"}})
			return
		}
		if body.InstanceID != c.status.InstanceID {
			writeJSON(w, 409, map[string]any{"error": map[string]string{"code": "instance_changed", "message": "service instance changed"}})
			return
		}
		if c.status.LaunchMode != Auto {
			writeJSON(w, 409, map[string]any{"error": map[string]string{"code": "service_not_owned", "message": "only an auto-started service can be stopped here"}})
			return
		}
		writeJSON(w, 202, map[string]any{"status": "stopping", "instance_id": c.status.InstanceID})
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		time.AfterFunc(10*time.Millisecond, func() { c.once.Do(func() { close(c.done) }) })
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 2 * time.Second}
	go func() {
		select {
		case <-ctx.Done():
		case <-c.done:
		}
		stop, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(stop)
	}()
	err := srv.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
