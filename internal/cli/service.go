package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/marcus/avatars/internal/discovery"
	"github.com/marcus/avatars/internal/lifecycle"
)

type serviceResult struct {
	lifecycle.Status
	State           string `json:"status"`
	LifecycleSocket string `json:"lifecycle_socket,omitempty"`
	LogPath         string `json:"log_path"`
}

func (a *app) serviceCommand(ctx context.Context, args []string) error {
	if a.remote {
		return bad("service commands cannot use --url / AVATARS_URL")
	}
	if len(args) == 0 {
		return bad("usage: avatars service ensure|status|stop")
	}
	verb := args[0]
	f := a.flags("service " + verb)
	timeout := f.Duration("timeout", 5*time.Second, "startup timeout")
	noAuto := f.Bool("no-auto-start", false, "do not start a missing service")
	listenDefault := os.Getenv("AVATARS_LISTEN")
	if listenDefault == "" {
		listenDefault = discovery.DefaultAddress
	}
	listenAddress := f.String("listen", listenDefault, "loopback HTTP address")
	publicURL := f.String("public-url", os.Getenv("AVATARS_PUBLIC_URL"), "trusted HTTPS proxy origin")
	if err := parse(f, args[1:]); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return bad("service " + verb + " takes no positional arguments")
	}
	if verb != "ensure" && verb != "status" && verb != "stop" {
		return bad("usage: avatars service ensure|status|stop")
	}
	endpoint, err := localEndpoint(*listenAddress)
	if err != nil {
		return bad(err.Error())
	}
	socket, lockPath, logPath, err := lifecycle.Paths(endpoint)
	if err != nil {
		return err
	}
	status, probeErr := probeService(ctx, endpoint)
	if status.Endpoint == "" {
		status.Endpoint = endpoint
	}
	wantedDataDir, err := canonicalDataDir(a.dataDir)
	if err != nil {
		return err
	}
	if verb == "status" {
		if probeErr != nil {
			if !isDialFailure(probeErr) {
				return fmt.Errorf("service status is unknown; incompatible response at %s: %w", endpoint, probeErr)
			}
			return a.emitService(serviceResult{State: "stopped", Status: lifecycle.Status{Endpoint: endpoint}, LifecycleSocket: socket, LogPath: logPath})
		}
		if err := validateService(ctx, status, "", socket); err != nil {
			return err
		}
		return a.emitService(serviceResult{State: "running", Status: status, LifecycleSocket: socket, LogPath: logPath})
	}
	if verb == "stop" {
		if probeErr != nil {
			return fmt.Errorf("service is not running: %w", probeErr)
		}
		if status.InstanceID == "" || status.LaunchMode == "" {
			return fmt.Errorf("refusing to stop a service without lifecycle identity; stop its foreground or supervisor owner")
		}
		if status.LaunchMode != lifecycle.Auto {
			return fmt.Errorf("refusing to stop %s service; use its owning foreground process or supervisor", status.LaunchMode)
		}
		if err := shutdownService(ctx, socket, status.InstanceID); err != nil {
			return err
		}
		return a.emitService(serviceResult{State: "stopping", Status: status, LifecycleSocket: socket, LogPath: logPath})
	}
	if probeErr == nil {
		if err := validateService(ctx, status, wantedDataDir, socket); err != nil {
			return err
		}
		return a.emitService(serviceResult{State: "running", Status: status, LifecycleSocket: socket, LogPath: logPath})
	}
	if !isDialFailure(probeErr) {
		return fmt.Errorf("refusing to replace service at %s: %w", endpoint, probeErr)
	}
	if *noAuto || autoStartDisabled() {
		return fmt.Errorf("avatars service is unavailable and automatic startup is disabled (log: %s)", logPath)
	}
	readyCtx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	lock, err := acquireLifecycleLock(readyCtx, lockPath)
	if err != nil {
		return fmt.Errorf("acquire lifecycle lock: %w (log: %s)", err, logPath)
	}
	defer lock.Close()
	if status, err = probeService(readyCtx, endpoint); err == nil {
		if status.Endpoint == "" {
			status.Endpoint = endpoint
		}
		if err := validateService(readyCtx, status, wantedDataDir, socket); err != nil {
			return err
		}
		return a.emitService(serviceResult{State: "running", Status: status, LifecycleSocket: socket, LogPath: logPath})
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	childArgs := []string{"serve", "--listen", *listenAddress, "--launch-mode", "auto", "--lifecycle-socket", socket}
	if *publicURL != "" {
		childArgs = append(childArgs, "--public-url", *publicURL)
	}
	if a.dataDir != "" {
		childArgs = append(childArgs, "--data-dir", a.dataDir)
	}
	if _, err = startDetached(exe, childArgs, logPath); err != nil {
		return fmt.Errorf("start avatars service: %w (log: %s)", err, logPath)
	}
	ticker := time.NewTicker(30 * time.Millisecond)
	defer ticker.Stop()
	for {
		status, err = probeService(readyCtx, endpoint)
		if err == nil {
			if status.Endpoint == "" {
				status.Endpoint = endpoint
			}
			if status.LaunchMode != lifecycle.Auto {
				return fmt.Errorf("unexpected service answered readiness at %s", endpoint)
			}
			if validationErr := validateService(readyCtx, status, wantedDataDir, socket); validationErr != nil {
				select {
				case <-readyCtx.Done():
					return fmt.Errorf("service lifecycle readiness timed out: %w (log: %s)", readyCtx.Err(), logPath)
				case <-ticker.C:
					continue
				}
			}
			return a.emitService(serviceResult{State: "running", Status: status, LifecycleSocket: socket, LogPath: logPath})
		}
		select {
		case <-readyCtx.Done():
			return fmt.Errorf("service readiness timed out: %w (log: %s)", readyCtx.Err(), logPath)
		case <-ticker.C:
		}
	}
}

func validateService(ctx context.Context, status lifecycle.Status, wantedDataDir, socket string) error {
	if status.Service != "avatars" || status.APIVersion != "v1" {
		return fmt.Errorf("incompatible service at %s", status.Endpoint)
	}
	if wantedDataDir != "" {
		running, err := canonicalDataDir(status.DataDir)
		if err != nil || running == "" || running != wantedDataDir {
			return fmt.Errorf("running service uses data directory %q; refusing requested %q", status.DataDir, wantedDataDir)
		}
	}
	if status.LaunchMode == lifecycle.Auto {
		controlled, err := lifecycleStatus(ctx, socket)
		if err != nil || controlled.InstanceID == "" || controlled.InstanceID != status.InstanceID {
			return fmt.Errorf("auto service at %s is missing matching lifecycle control", status.Endpoint)
		}
	}
	return nil
}

func canonicalDataDir(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolute), nil
}

func localEndpoint(address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", fmt.Errorf("--listen must be a loopback host:port")
	}
	if host == "localhost" {
		host = "127.0.0.1"
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil || !ip.IsLoopback() {
		return "", fmt.Errorf("--listen must use a loopback IP or localhost")
	}
	return (&url.URL{Scheme: "http", Host: net.JoinHostPort(host, port)}).String(), nil
}

func (a *app) emitService(v serviceResult) error {
	if a.json {
		return a.emit(v)
	}
	fmt.Fprintf(a.out, "%s %s (pid %d, %s)\n", v.State, v.Status.Endpoint, v.Status.PID, v.Status.LaunchMode)
	return nil
}
func autoStartDisabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AVATARS_AUTO_START"))) {
	case "0", "false", "no", "off":
		return true
	}
	return false
}
func probeService(ctx context.Context, endpoint string) (lifecycle.Status, error) {
	var out lifecycle.Status
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/api/v1/health", nil)
	if err != nil {
		return out, err
	}
	client := &http.Client{Timeout: 750 * time.Millisecond}
	res, err := client.Do(req)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return out, fmt.Errorf("health returned %s", res.Status)
	}
	err = json.NewDecoder(res.Body).Decode(&out)
	return out, err
}
func isDialFailure(err error) bool { var op *net.OpError; return errors.As(err, &op) }
func shutdownService(ctx context.Context, socket, id string) error {
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}
	defer transport.CloseIdleConnections()
	body, _ := json.Marshal(map[string]string{"instance_id": id})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://unix/v1/shutdown", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := (&http.Client{Transport: transport, Timeout: 2 * time.Second}).Do(req)
	if err != nil {
		return fmt.Errorf("lifecycle shutdown failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 202 {
		return fmt.Errorf("lifecycle shutdown refused with %s", res.Status)
	}
	return nil
}

func lifecycleStatus(ctx context.Context, socket string) (lifecycle.Status, error) {
	var out lifecycle.Status
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}
	defer transport.CloseIdleConnections()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://unix/v1/status", nil)
	res, err := (&http.Client{Transport: transport, Timeout: 500 * time.Millisecond}).Do(req)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return out, fmt.Errorf("lifecycle status returned %s", res.Status)
	}
	err = json.NewDecoder(res.Body).Decode(&out)
	return out, err
}
