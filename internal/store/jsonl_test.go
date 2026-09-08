package store_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/marcus/avatars/internal/library"
	"github.com/marcus/avatars/internal/store"
)

func record(id string) library.Collection {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	return library.Collection{
		ID: "col_" + id, Name: "Collection " + id, Style: "gorey", CreatedAt: now,
		Avatars: []library.Avatar{{ID: "av_" + id, CollectionID: "col_" + id, Style: "gorey", Seed: id, CreatedAt: now}},
	}
}

func TestPersistenceAndPrivatePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "library.jsonl")
	s, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(context.Background(), record("one")); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(context.Background(), record("two")); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	collections, err := reopened.List(context.Background())
	if err != nil || len(collections) != 2 || collections[0].ID != "col_one" || collections[1].ID != "col_two" {
		t.Fatalf("saved collections missing after restart: %+v, %v", collections, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("library must be private: %v, %v", info, err)
	}
	data, err := os.ReadFile(path)
	if err != nil || bytes.Count(data, []byte{'\n'}) != 2 {
		t.Fatalf("want one JSONL record per batch: %q, %v", data, err)
	}
}

func TestCorruptionRefusesReadsAndWritesWithoutChangingFile(t *testing.T) {
	for _, corruption := range []struct {
		name, content, want string
	}{
		{"partial JSON", `{"id":"interrupted`, "incomplete final record"},
		{"missing newline", `{}`, "incomplete final record"},
		{"malformed record", "bad JSON\n", "corrupt record"},
		{"missing fields", "{}\n", "corrupt record"},
		{"null record", "null\n", "corrupt record"},
	} {
		t.Run(corruption.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "library.jsonl")
			s, err := store.New(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Save(context.Background(), record("good")); err != nil {
				t.Fatal(err)
			}
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.WriteString(corruption.content); err != nil {
				t.Fatal(err)
			}
			f.Close()
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			_, listErr := s.List(context.Background())
			saveErr := s.Save(context.Background(), record("next"))
			for _, err := range []error{listErr, saveErr} {
				if err == nil || !strings.Contains(err.Error(), corruption.want) || !strings.Contains(err.Error(), "line 2") {
					t.Fatalf("want explicit line 2 corruption error, got %v", err)
				}
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("corrupt file was changed: %v", err)
			}
		})
	}
}

func TestDuplicateIDsDoNotCorruptLog(t *testing.T) {
	s, err := store.New(filepath.Join(t.TempDir(), "library.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(context.Background(), record("one")); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(context.Background(), record("one")); err == nil {
		t.Fatal("duplicate collection ID accepted")
	}
	duplicate := record("two")
	duplicate.Avatars[0].ID = "av_one"
	if err := s.Save(context.Background(), duplicate); err == nil {
		t.Fatal("duplicate avatar ID accepted")
	}
	list, err := s.List(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatalf("duplicate write damaged existing record: %+v, %v", list, err)
	}
}

func TestInitializationDoesNotBlockAndLockWaitHonorsCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.jsonl")
	s, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}
	type initialization struct {
		store *store.Store
		err   error
	}
	initialized := make(chan initialization, 1)
	go func() {
		reopened, err := store.New(path)
		initialized <- initialization{reopened, err}
	}()
	select {
	case result := <-initialized:
		if result.err != nil {
			t.Fatal(result.err)
		}
		s = result.store
	case <-time.After(time.Second):
		t.Fatal("initialization waited for the data lock without a cancellable context")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if _, err := s.List(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock wait ignored cancellation: %v", err)
	}
}

func TestConcurrentProcessesPreserveAllCollections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "library.jsonl")
	const writers = 4
	var wg sync.WaitGroup
	for writer := 0; writer < writers; writer++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run=^TestProcessWriter$")
			cmd.Env = append(os.Environ(), "AVATARS_TEST_STORE_PATH="+path, fmt.Sprintf("AVATARS_TEST_WRITER=%d", writer))
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("writer %d: %v\n%s", writer, err, output)
			}
		}()
	}
	wg.Wait()
	s, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	collections, err := s.List(context.Background())
	if err != nil || len(collections) != writers*8 {
		t.Fatalf("concurrent writes lost collections: %d, %v", len(collections), err)
	}
}

func TestProcessWriter(t *testing.T) {
	path := os.Getenv("AVATARS_TEST_STORE_PATH")
	if path == "" {
		t.Skip("subprocess helper")
	}
	s, err := store.New(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if err := s.Save(context.Background(), record(fmt.Sprintf("%s-%d", os.Getenv("AVATARS_TEST_WRITER"), i))); err != nil {
			t.Fatal(err)
		}
	}
}
