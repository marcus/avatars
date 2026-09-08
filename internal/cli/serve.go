package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/marcus/avatars/internal/discovery"
	"github.com/marcus/avatars/internal/httpapi"
	"github.com/marcus/avatars/internal/lifecycle"
	"github.com/marcus/avatars/internal/studio"
)

func (a *app) serve(ctx context.Context, args []string) error {
	if a.remote {
		return bad("serve cannot use --url / AVATARS_URL")
	}
	f := a.flags("serve")
	address := f.String("listen", discovery.DefaultAddress, "listen address")
	open := f.Bool("open", false, "open studio")
	publicURL := f.String("public-url", os.Getenv("AVATARS_PUBLIC_URL"), "trusted HTTPS proxy origin")
	launchMode := f.String("launch-mode", string(lifecycle.Foreground), "service owner: foreground, auto, or supervised")
	lifecycleSocket := f.String("lifecycle-socket", "", "restricted local lifecycle socket")
	if e := parse(f, args); e != nil {
		return e
	}
	if f.NArg() != 0 {
		return bad("serve takes no positional arguments")
	}
	normalized, err := httpapi.NormalizePublicURL(*publicURL)
	if err != nil {
		return bad("--public-url: " + err.Error())
	}
	*publicURL = normalized
	mode := lifecycle.LaunchMode(*launchMode)
	if mode != lifecycle.Auto && mode != lifecycle.Foreground && mode != lifecycle.Supervised {
		return bad("--launch-mode must be foreground, auto, or supervised")
	}
	if mode == lifecycle.Auto && *lifecycleSocket == "" {
		return bad("--lifecycle-socket is required for auto launch mode")
	}
	host, _, e := net.SplitHostPort(*address)
	if e != nil {
		return bad("--listen must be a loopback host:port")
	}
	ip := net.ParseIP(host)
	if host == "localhost" {
		ip = net.ParseIP("127.0.0.1")
		*address = "127.0.0.1:" + (*address)[len("localhost:"):]
	}
	if ip == nil || !ip.IsLoopback() {
		return bad("--listen must use a loopback IP or localhost")
	}
	if e = a.local(); e != nil {
		return e
	}
	listener, e := net.Listen("tcp", *address)
	if e != nil {
		return e
	}
	status := lifecycle.NewStatus(mode, "http://"+listener.Addr().String(), a.dataDir)
	var controller *lifecycle.Controller
	if *lifecycleSocket != "" {
		controller = lifecycle.NewController(status)
		if err := controller.Start(ctx, *lifecycleSocket); err != nil {
			_ = listener.Close()
			return fmt.Errorf("start lifecycle control: %w", err)
		}
	}
	server := &http.Server{Handler: httpapi.NewWithConfig(a.service, a.engine, studio.Handler(), httpapi.Config{PublicURL: *publicURL, Status: &status}), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	studioURL := "http://" + listener.Addr().String()
	if *publicURL != "" {
		studioURL = *publicURL
	}
	if a.json {
		e = a.emit(map[string]string{"url": studioURL, "data_dir": a.dataDir, "status": "listening"})
	} else {
		fmt.Fprintf(a.out, "Avatar Studio\n%s\nLibrary: %s\nPress Ctrl-C to stop.\n", studioURL, a.dataDir)
	}
	if e != nil {
		_ = listener.Close()
		return e
	}
	if *open {
		if e := openBrowser(studioURL); e != nil {
			fmt.Fprintf(a.errOut, "Could not open browser: %v. Open %s\n", e, studioURL)
		}
	}
	done := make(chan struct{})
	if controller != nil {
		go func() {
			select {
			case <-controller.Done():
				stop, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = server.Shutdown(stop)
			case <-done:
			}
		}()
	}
	go func() {
		select {
		case <-ctx.Done():
			stop, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(stop)
		case <-done:
		}
	}()
	e = server.Serve(listener)
	close(done)
	if errors.Is(e, http.ErrServerClosed) {
		return nil
	}
	return e
}
func openBrowser(address string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", address)
	case "linux":
		c = exec.Command("xdg-open", address)
	default:
		return fmt.Errorf("browser opening is unsupported on %s", runtime.GOOS)
	}
	return c.Run()
}
