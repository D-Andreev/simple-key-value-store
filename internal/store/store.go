// Package store defines the key/value storage abstraction used by the CLI,
// decoupling command logic from the underlying storage implementation.
package store

import (
	"errors"
	"sync"
)

// ErrKeyNotFound is returned by Get and Delete when the requested key does
// not exist in the store.
var ErrKeyNotFound = errors.New("key not found")

// Store abstracts key/value storage. Keys and values are strings.
// Implementations must be safe for concurrent use.
type Store interface {
	// Set stores value under key, overwriting any existing value.
	Set(key, value string)
	// Get returns the value stored under key, or ErrKeyNotFound if it
	// does not exist.
	Get(key string) (string, error)
	// Delete removes key from the store, or returns ErrKeyNotFound if it
	// does not exist.
	Delete(key string) error
	// List returns a snapshot of all key/value pairs currently stored.
	// Mutating the returned map does not affect the store.
	List() map[string]string
}

// MemoryStore is an in-memory Store implementation backed by a Go map.
// It is safe for concurrent use. Data is not persisted across restarts.
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewMemoryStore returns an empty, ready-to-use MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]string)}
}

// Set implements Store.
func (s *MemoryStore) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get implements Store.
func (s *MemoryStore) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return value, nil
}

// Delete implements Store.
func (s *MemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; !ok {
		return ErrKeyNotFound
	}
	delete(s.data, key)
	return nil
}

// List implements Store.
func (s *MemoryStore) List() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out
}
