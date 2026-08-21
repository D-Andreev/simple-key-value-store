package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/D-Andreev/simple-key-value-store/internal/store"
)

// runWithStore executes the root command against the given store with the
// provided args and returns combined stdout, stderr, and the error (if any).
func runWithStore(t *testing.T, s store.Store, args []string) (stdout, stderr string, err error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	cmd := NewRootCmd(s)
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

func TestSetCmd(t *testing.T) {
	s := store.NewMemoryStore()
	_, stderr, err := runWithStore(t, s, []string{"set", "foo", "bar"})
	if err != nil {
		t.Fatalf("set returned unexpected error: %v (stderr=%q)", err, stderr)
	}

	got, getErr := s.Get("foo")
	if getErr != nil {
		t.Fatalf("Get(foo) after set returned error: %v", getErr)
	}
	if got != "bar" {
		t.Errorf("Get(foo) = %q, want %q", got, "bar")
	}
}

func TestGetCmd(t *testing.T) {
	tests := []struct {
		name       string
		seed       map[string]string
		key        string
		wantStdout string
		wantErr    bool
	}{
		{
			name:       "existing key",
			seed:       map[string]string{"foo": "bar"},
			key:        "foo",
			wantStdout: "bar\n",
		},
		{
			name:    "missing key",
			seed:    map[string]string{},
			key:     "missing",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store.NewMemoryStore()
			for k, v := range tt.seed {
				if err := s.Set(k, v); err != nil {
					t.Fatalf("Set(%q, %q) returned unexpected error: %v", k, v, err)
				}
			}

			stdout, stderr, err := runWithStore(t, s, []string{"get", tt.key})
			if tt.wantErr {
				if err == nil {
					t.Fatalf("get %q: expected error, got nil", tt.key)
				}
				if !strings.Contains(stderr, "key not found") {
					t.Errorf("get %q: stderr = %q, want it to contain %q", tt.key, stderr, "key not found")
				}
				return
			}
			if err != nil {
				t.Fatalf("get %q returned unexpected error: %v", tt.key, err)
			}
			if stdout != tt.wantStdout {
				t.Errorf("get %q: stdout = %q, want %q", tt.key, stdout, tt.wantStdout)
			}
		})
	}
}

func TestDeleteCmd(t *testing.T) {
	t.Run("existing key", func(t *testing.T) {
		s := store.NewMemoryStore()
		if err := s.Set("foo", "bar"); err != nil {
			t.Fatalf("Set(foo, bar) returned unexpected error: %v", err)
		}

		_, stderr, err := runWithStore(t, s, []string{"delete", "foo"})
		if err != nil {
			t.Fatalf("delete returned unexpected error: %v (stderr=%q)", err, stderr)
		}
		if _, getErr := s.Get("foo"); getErr == nil {
			t.Error("Get(foo) after delete: expected error, got nil")
		}
	})

	t.Run("missing key", func(t *testing.T) {
		s := store.NewMemoryStore()

		_, stderr, err := runWithStore(t, s, []string{"delete", "missing"})
		if err == nil {
			t.Fatal("delete missing: expected error, got nil")
		}
		if !strings.Contains(stderr, "key not found") {
			t.Errorf("delete missing: stderr = %q, want it to contain %q", stderr, "key not found")
		}
	})
}

func TestListCmd(t *testing.T) {
	tests := []struct {
		name       string
		seed       map[string]string
		wantStdout string
	}{
		{name: "empty store", seed: map[string]string{}, wantStdout: ""},
		{
			name:       "multiple keys sorted",
			seed:       map[string]string{"b": "2", "a": "1", "c": "3"},
			wantStdout: "a=1\nb=2\nc=3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store.NewMemoryStore()
			for k, v := range tt.seed {
				if err := s.Set(k, v); err != nil {
					t.Fatalf("Set(%q, %q) returned unexpected error: %v", k, v, err)
				}
			}

			stdout, stderr, err := runWithStore(t, s, []string{"list"})
			if err != nil {
				t.Fatalf("list returned unexpected error: %v (stderr=%q)", err, stderr)
			}
			if stdout != tt.wantStdout {
				t.Errorf("list: stdout = %q, want %q", stdout, tt.wantStdout)
			}
		})
	}
}
