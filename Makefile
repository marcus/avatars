PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS = -s -w -X github.com/marcus/avatars/internal/buildinfo.Version=$(VERSION) -X github.com/marcus/avatars/internal/buildinfo.Commit=$(COMMIT)

release_goals := $(filter check-release-state release-tap,$(MAKECMDGOALS))
ifneq ($(release_goals),)
ifneq ($(origin RELEASE_VERSION),environment)
$(error set RELEASE_VERSION in the environment, for example: RELEASE_VERSION=v1.0.0 make $(firstword $(release_goals)))
endif
endif

.PHONY: build install install-local install-worktree use-homebrew verify-homebrew install-status test test-race vet fmt fmt-check clean release-snapshot check-release-state release release-dry-run release-tap

build:
	mkdir -p bin
	go build -ldflags '$(LDFLAGS)' -o bin/avatars ./cmd/avatars

install:
	install -d '$(BINDIR)'
	go build -ldflags '$(LDFLAGS)' -o '$(BINDIR)/avatars' ./cmd/avatars

install-local:
	./scripts/dev-install.sh install-local

install-worktree:
	./scripts/dev-install.sh install-worktree

use-homebrew:
	./scripts/dev-install.sh use-homebrew

verify-homebrew:
	./scripts/dev-install.sh verify-homebrew

install-status:
	./scripts/dev-install.sh status

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal pkg

fmt-check:
	@test -z "$$(gofmt -l cmd internal pkg)" || { gofmt -l cmd internal pkg; exit 1; }

clean:
	rm -rf bin dist

release-snapshot:
	goreleaser release --snapshot --clean

check-release-state:
	@test -n "$${RELEASE_VERSION:-}" || { echo 'RELEASE_VERSION=vX.Y.Z is required' >&2; exit 2; }
	./scripts/check-release-state.sh pre-tag

release:
	./scripts/release.sh

release-dry-run:
	./scripts/release.sh --dry-run

release-tap:
	@test -n "$${RELEASE_VERSION:-}" || { echo 'RELEASE_VERSION=vX.Y.Z is required' >&2; exit 2; }
	./scripts/publish-homebrew-tap.sh
