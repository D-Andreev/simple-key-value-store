package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

// newTestFileStore returns a FileStore rooted at a fresh temp dir, failing
// the test immediately if construction errors.
func newTestFileStore(t *testing.T) (*FileStore, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore(%q) returned unexpected error: %v", dir, err)
	}
	return s, dir
}

func TestFileStore_SetGet(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "simple key value", key: "foo", value: "bar"},
		{name: "empty value", key: "empty", value: ""},
		{name: "overwrite existing key", key: "foo", value: "baz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := newTestFileStore(t)
			if err := s.Set(tt.key, tt.value); err != nil {
				t.Fatalf("Set(%q, %q) returned unexpected error: %v", tt.key, tt.value, err)
			}

			got, err := s.Get(tt.key)
			if err != nil {
				t.Fatalf("Get(%q) returned unexpected error: %v", tt.key, err)
			}
			if got != tt.value {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.value)
			}
		})
	}
}

func TestFileStore_Get_MissingKey(t *testing.T) {
	s, _ := newTestFileStore(t)

	_, err := s.Get("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Get(missing) error = %v, want ErrKeyNotFound", err)
	}
}

func TestFileStore_Delete(t *testing.T) {
	s, _ := newTestFileStore(t)
	if err := s.Set("foo", "bar"); err != nil {
		t.Fatalf("Set(foo, bar) returned unexpected error: %v", err)
	}

	if err := s.Delete("foo"); err != nil {
		t.Fatalf("Delete(foo) returned unexpected error: %v", err)
	}

	if _, err := s.Get("foo"); !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Get(foo) after delete error = %v, want ErrKeyNotFound", err)
	}
}

func TestFileStore_Delete_MissingKey(t *testing.T) {
	s, _ := newTestFileStore(t)

	err := s.Delete("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Delete(missing) error = %v, want ErrKeyNotFound", err)
	}
}

func TestFileStore_List(t *testing.T) {
	tests := []struct {
		name string
		set  map[string]string
		want map[string]string
	}{
		{name: "empty store", set: map[string]string{}, want: map[string]string{}},
		{
			name: "multiple keys",
			set:  map[string]string{"b": "2", "a": "1", "c": "3"},
			want: map[string]string{"a": "1", "b": "2", "c": "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := newTestFileStore(t)
			for k, v := range tt.set {
				if err := s.Set(k, v); err != nil {
					t.Fatalf("Set(%q, %q) returned unexpected error: %v", k, v, err)
				}
			}

			got := s.List()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileStore_List_ReturnsCopy(t *testing.T) {
	s, _ := newTestFileStore(t)
	if err := s.Set("foo", "bar"); err != nil {
		t.Fatalf("Set(foo, bar) returned unexpected error: %v", err)
	}

	list := s.List()
	list["foo"] = "mutated"

	got, err := s.Get("foo")
	if err != nil {
		t.Fatalf("Get(foo) returned unexpected error: %v", err)
	}
	if got != "bar" {
		t.Errorf("internal state mutated via List() result: Get(foo) = %q, want %q", got, "bar")
	}
}

// TestFileStore_PersistsAcrossInstances is the load-on-start behavior:
// data written by one FileStore instance must be visible to a fresh
// FileStore constructed later against the same data dir, simulating a kvs
// process restart.
func TestFileStore_PersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()

	first, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore(%q) returned unexpected error: %v", dir, err)
	}
	if err := first.Set("foo", "bar"); err != nil {
		t.Fatalf("Set(foo, bar) returned unexpected error: %v", err)
	}
	if err := first.Set("baz", "qux"); err != nil {
		t.Fatalf("Set(baz, qux) returned unexpected error: %v", err)
	}
	if err := first.Delete("baz"); err != nil {
		t.Fatalf("Delete(baz) returned unexpected error: %v", err)
	}

	second, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore(%q) (second instance) returned unexpected error: %v", dir, err)
	}

	got, err := second.Get("foo")
	if err != nil {
		t.Fatalf("Get(foo) on reloaded store returned unexpected error: %v", err)
	}
	if got != "bar" {
		t.Errorf("Get(foo) on reloaded store = %q, want %q", got, "bar")
	}

	if _, err := second.Get("baz"); !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Get(baz) on reloaded store error = %v, want ErrKeyNotFound (deleted before restart)", err)
	}
}

func TestNewFileStore_CreatesMissingDataDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "data-dir")

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("precondition failed: %q already exists", dir)
	}

	s, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore(%q) returned unexpected error: %v", dir, err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("data dir %q was not created: %v", dir, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q was created but is not a directory", dir)
	}

	if err := s.Set("foo", "bar"); err != nil {
		t.Fatalf("Set(foo, bar) returned unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, storeFileName)); err != nil {
		t.Errorf("store file was not created in %q: %v", dir, err)
	}
}

func TestNewFileStore_InvalidJSON_FailsLoudly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, storeFileName)
	if err := os.WriteFile(path, []byte("{not valid json"), 0o600); err != nil {
		t.Fatalf("failed to seed corrupt store file: %v", err)
	}

	_, err := NewFileStore(dir)
	if err == nil {
		t.Fatal("NewFileStore with corrupt store.json: expected error, got nil")
	}
}

func TestNewFileStore_DataDirIsRegularFile_FailsLoudly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(path, []byte("i am a file"), 0o600); err != nil {
		t.Fatalf("failed to seed regular file: %v", err)
	}

	_, err := NewFileStore(path)
	if err == nil {
		t.Fatal("NewFileStore with data dir pointing at a regular file: expected error, got nil")
	}
}

// TestFileStore_ConcurrentAccess mirrors TestMemoryStore_ConcurrentAccess:
// it exercises Set/Get/Delete/List from many goroutines at once. It
// doesn't assert on the interleaved values (there's no deterministic
// outcome to check) — its job is to give `go test -race` something to
// catch if FileStore's locking is ever broken.
func TestFileStore_ConcurrentAccess(t *testing.T) {
	s, _ := newTestFileStore(t)

	const goroutines = 10
	const opsPerGoroutine = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("key-%d", (g+i)%10)
				_ = s.Set(key, fmt.Sprintf("value-%d-%d", g, i))
				_, _ = s.Get(key)
				_ = s.Delete(key)
				_ = s.List()
			}
		}(g)
	}
	wg.Wait()
}

var _ Store = (*FileStore)(nil)
