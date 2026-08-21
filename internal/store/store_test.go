package store

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestMemoryStore_SetGet(t *testing.T) {
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
			s := NewMemoryStore()
			s.Set(tt.key, tt.value)

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

func TestMemoryStore_Get_MissingKey(t *testing.T) {
	s := NewMemoryStore()

	_, err := s.Get("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Get(missing) error = %v, want ErrKeyNotFound", err)
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	s := NewMemoryStore()
	s.Set("foo", "bar")

	if err := s.Delete("foo"); err != nil {
		t.Fatalf("Delete(foo) returned unexpected error: %v", err)
	}

	if _, err := s.Get("foo"); !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Get(foo) after delete error = %v, want ErrKeyNotFound", err)
	}
}

func TestMemoryStore_Delete_MissingKey(t *testing.T) {
	s := NewMemoryStore()

	err := s.Delete("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Delete(missing) error = %v, want ErrKeyNotFound", err)
	}
}

func TestMemoryStore_List(t *testing.T) {
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
			s := NewMemoryStore()
			for k, v := range tt.set {
				s.Set(k, v)
			}

			got := s.List()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryStore_List_ReturnsCopy(t *testing.T) {
	s := NewMemoryStore()
	s.Set("foo", "bar")

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

// TestMemoryStore_ConcurrentAccess exercises Set/Get/Delete from many
// goroutines at once. It doesn't assert on the interleaved values (there's
// no deterministic outcome to check) — its job is to give `go test -race`
// something to catch if MemoryStore's locking is ever broken.
func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	s := NewMemoryStore()

	const goroutines = 50
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := fmt.Sprintf("key-%d", (g+i)%10)
				s.Set(key, fmt.Sprintf("value-%d-%d", g, i))
				_, _ = s.Get(key)
				_ = s.Delete(key)
				_ = s.List()
			}
		}(g)
	}
	wg.Wait()
}

var _ Store = (*MemoryStore)(nil)
