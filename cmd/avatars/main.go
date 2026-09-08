// Command avatars generates, saves, and serves avatars through a shared core.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/marcus/avatars/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
