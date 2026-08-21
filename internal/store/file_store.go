package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// storeFileName is the name of the JSON file FileStore persists to inside
// the configured data directory.
const storeFileName = "store.json"

// FileStore is a Store implementation that persists data as JSON on disk.
// It loads any existing data from <dataDir>/store.json when constructed,
// so data survives process restarts, and flushes the full snapshot back to
// that file synchronously on every Set/Delete — there is no batching or
// periodic flush, since kvs is a one-shot CLI with no long-running process
// to defer a flush to. It is safe for concurrent use.
type FileStore struct {
	mu   sync.RWMutex
	path string
	data map[string]string
}

// NewFileStore returns a FileStore backed by <dataDir>/store.json.
//
// dataDir is created (including any missing parents) if it does not exist.
// If dataDir exists but is a regular file rather than a directory,
// NewFileStore fails loudly with an error rather than trying (and failing
// confusingly) to write store.json inside it. If store.json already
// exists, its contents are loaded into memory; if it exists but contains
// invalid JSON, NewFileStore returns an error instead of silently falling
// back to an empty store.
func NewFileStore(dataDir string) (*FileStore, error) {
	if err := ensureDir(dataDir); err != nil {
		return nil, err
	}

	path := filepath.Join(dataDir, storeFileName)
	data, err := loadStoreFile(path)
	if err != nil {
		return nil, err
	}

	return &FileStore{path: path, data: data}, nil
}

// ensureDir makes sure dataDir exists as a directory, creating it (and any
// missing parents) if needed. It fails if dataDir already exists but is a
// regular file.
func ensureDir(dataDir string) error {
	info, err := os.Stat(dataDir)
	switch {
	case err == nil:
		if !info.IsDir() {
			return fmt.Errorf("data dir %q is not a directory", dataDir)
		}
		return nil
	case os.IsNotExist(err):
		if mkErr := os.MkdirAll(dataDir, 0o755); mkErr != nil {
			return fmt.Errorf("create data dir %q: %w", dataDir, mkErr)
		}
		return nil
	default:
		return fmt.Errorf("stat data dir %q: %w", dataDir, err)
	}
}

// loadStoreFile reads and decodes path, returning an empty map if the file
// does not exist yet (first run in a fresh data dir) and an error if it
// exists but its contents aren't valid JSON.
func loadStoreFile(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("read store file %q: %w", path, err)
	}

	if len(bytes.TrimSpace(raw)) == 0 {
		return make(map[string]string), nil
	}

	data := make(map[string]string)
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("store file %q contains invalid JSON: %w", path, err)
	}
	return data, nil
}

// Set implements Store. The new value is flushed to disk before Set
// returns; if the flush fails, the in-memory state is rolled back so it
// stays consistent with what's on disk, and the error is returned.
func (s *FileStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev, existed := s.data[key]
	s.data[key] = value
	if err := s.flushLocked(); err != nil {
		if existed {
			s.data[key] = prev
		} else {
			delete(s.data, key)
		}
		return err
	}
	return nil
}

// Get implements Store.
func (s *FileStore) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return value, nil
}

// Delete implements Store. The removal is flushed to disk before Delete
// returns; if the flush fails, the in-memory state is rolled back so it
// stays consistent with what's on disk, and the error is returned.
func (s *FileStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}
	delete(s.data, key)
	if err := s.flushLocked(); err != nil {
		s.data[key] = prev
		return err
	}
	return nil
}

// List implements Store.
func (s *FileStore) List() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out
}

// flushLocked writes the full in-memory snapshot to disk as JSON, via a
// write-then-rename so a crash or failure mid-write can never leave
// store.json truncated or corrupt. Callers must hold s.mu for writing.
func (s *FileStore) flushLocked() error {
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store data: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return fmt.Errorf("write store file %q: %w", s.path, err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace store file %q: %w", s.path, err)
	}
	return nil
}
