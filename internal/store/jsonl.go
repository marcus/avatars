// Package store provides the local JSONL adapter for the avatar library.
package store

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/marcus/avatars/internal/library"
)

// Store keeps one collection per line. Each operation opens and locks the file
// independently, allowing both multiple processes and multiple goroutines.
// Files must not be replaced or edited while an avatars process is using them.
type Store struct {
	path string
}

var _ library.Store = (*Store)(nil)

func New(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("library path is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve library path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0700); err != nil {
		return nil, fmt.Errorf("create library directory: %w", err)
	}
	s := &Store{path: abs}
	// Initialization does not read or write collection data and must not wait
	// for another process's lock. Actual operations acquire a cancellable lock.
	f, err := os.OpenFile(abs, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open library: %w", err)
	}
	if err := f.Chmod(0600); err != nil {
		f.Close()
		return nil, fmt.Errorf("set library permissions: %w", err)
	}
	return s, f.Close()
}

// Save validates the existing log before appending and fsyncs one whole batch.
// An interrupted write is reported by subsequent reads; it is never discarded
// or overwritten automatically, preserving evidence needed for manual repair.
func (s *Store) Save(ctx context.Context, collection library.Collection) error {
	if err := validate(collection); err != nil {
		return fmt.Errorf("invalid collection: %w", err)
	}
	data, err := json.Marshal(collection)
	if err != nil {
		return fmt.Errorf("encode collection: %w", err)
	}
	data = append(data, '\n')
	f, err := s.open(ctx, syscall.LOCK_EX)
	if err != nil {
		return err
	}
	defer f.Close()
	collections, err := read(ctx, f)
	if err != nil {
		return err
	}
	newIDs := map[string]bool{collection.ID: true}
	for _, item := range collection.Avatars {
		newIDs[item.ID] = true
	}
	for _, existing := range collections {
		if newIDs[existing.ID] {
			return fmt.Errorf("ID %q already exists", existing.ID)
		}
		for _, item := range existing.Avatars {
			if newIDs[item.ID] {
				return fmt.Errorf("ID %q already exists", item.ID)
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	n, err := f.Write(data)
	if err != nil {
		return fmt.Errorf("append collection: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("append collection: %w", io.ErrShortWrite)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync collection: %w", err)
	}
	return nil
}

func (s *Store) List(ctx context.Context) ([]library.Collection, error) {
	f, err := s.open(ctx, syscall.LOCK_SH)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return read(ctx, f)
}

func (s *Store) open(ctx context.Context, operation int) (*os.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("open library: %w", err)
	}
	for {
		err = syscall.Flock(int(f.Fd()), operation|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) && !errors.Is(err, syscall.EINTR) {
			f.Close()
			return nil, fmt.Errorf("lock library: %w", err)
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			f.Close()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	if err := f.Chmod(0600); err != nil {
		f.Close()
		return nil, fmt.Errorf("set library permissions: %w", err)
	}
	return f, nil
}

func read(ctx context.Context, f *os.File) ([]library.Collection, error) {
	reader := bufio.NewReader(f)
	collections := []library.Collection{}
	ids := make(map[string]bool)
	for line := 1; ; line++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		data, err := reader.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			if len(data) == 0 {
				return collections, nil
			}
			return nil, fmt.Errorf("library %s line %d: incomplete final record (missing newline); repair the file before continuing", f.Name(), line)
		}
		if err != nil {
			return nil, fmt.Errorf("read library line %d: %w", line, err)
		}
		var collection library.Collection
		if err := json.Unmarshal(data, &collection); err != nil {
			return nil, fmt.Errorf("library %s line %d: corrupt record: %w", f.Name(), line, err)
		}
		if err := validate(collection); err != nil {
			return nil, fmt.Errorf("library %s line %d: corrupt record: %w", f.Name(), line, err)
		}
		if ids[collection.ID] {
			return nil, fmt.Errorf("library %s line %d: duplicate collection ID %q", f.Name(), line, collection.ID)
		}
		ids[collection.ID] = true
		for _, item := range collection.Avatars {
			if ids[item.ID] {
				return nil, fmt.Errorf("library %s line %d: duplicate avatar ID %q", f.Name(), line, item.ID)
			}
			ids[item.ID] = true
		}
		collections = append(collections, collection)
	}
}

// validate checks record shape, not creation policy. Historical records remain
// readable if a future version changes its supported styles or request limits.
func validate(collection library.Collection) error {
	if collection.ID == "" || collection.Style == "" || collection.CreatedAt.IsZero() || len(collection.Avatars) == 0 {
		return errors.New("collection must include ID, style, created_at, and avatars")
	}
	ids := make(map[string]bool)
	for _, item := range collection.Avatars {
		if item.ID == "" || item.ID == collection.ID || item.CollectionID != collection.ID || item.Style != collection.Style || item.CreatedAt.IsZero() {
			return errors.New("avatar must include ID, matching collection and style, and created_at")
		}
		if ids[item.ID] {
			return fmt.Errorf("duplicate avatar ID %q", item.ID)
		}
		ids[item.ID] = true
	}
	return nil
}
