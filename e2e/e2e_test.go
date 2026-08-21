// Package e2e builds the kvs binary and exercises it as a subprocess,
// asserting on stdout/stderr/exit code per command.
package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "kvs-e2e")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	binPath = filepath.Join(dir, "kvs")
	build := exec.Command("go", "build", "-o", binPath, "../cmd/kvs")
	if out, err := build.CombinedOutput(); err != nil {
		panic("failed to build kvs binary: " + err.Error() + "\n" + string(out))
	}

	os.Exit(m.Run())
}

func runKVS(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err == nil {
		return outBuf.String(), errBuf.String(), 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return outBuf.String(), errBuf.String(), exitErr.ExitCode()
	}
	t.Fatalf("failed to run kvs %v: %v", args, err)
	return "", "", -1
}

func TestSet_Success(t *testing.T) {
	stdout, stderr, code := runKVS(t, "set", "foo", "bar")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 (stdout=%q stderr=%q)", code, stdout, stderr)
	}
}

func TestGet_MissingKey(t *testing.T) {
	stdout, stderr, code := runKVS(t, "get", "missing")
	if code == 0 {
		t.Errorf("exit code = 0, want non-zero (stdout=%q)", stdout)
	}
	if stderr != "key not found\n" {
		t.Errorf("stderr = %q, want %q", stderr, "key not found\n")
	}
}

func TestDelete_MissingKey(t *testing.T) {
	stdout, stderr, code := runKVS(t, "delete", "missing")
	if code == 0 {
		t.Errorf("exit code = 0, want non-zero (stdout=%q)", stdout)
	}
	if stderr != "key not found\n" {
		t.Errorf("stderr = %q, want %q", stderr, "key not found\n")
	}
}

func TestList_EmptyStore(t *testing.T) {
	stdout, stderr, code := runKVS(t, "list")
	if code != 0 {
		t.Errorf("exit code = %d, want 0 (stderr=%q)", code, stderr)
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want empty", stdout)
	}
}

// TestPersistence_SetThenRestartThenGet is the headline --data-dir
// scenario: a value set by one kvs process must be readable by a
// completely separate kvs process (a fresh restart) pointed at the same
// data dir.
func TestPersistence_SetThenRestartThenGet(t *testing.T) {
	dataDir := t.TempDir()

	stdout, stderr, code := runKVS(t, "--data-dir", dataDir, "set", "foo", "bar")
	if code != 0 {
		t.Fatalf("set exit code = %d, want 0 (stdout=%q stderr=%q)", code, stdout, stderr)
	}

	// New process, same --data-dir: simulates a restart.
	stdout, stderr, code = runKVS(t, "--data-dir", dataDir, "get", "foo")
	if code != 0 {
		t.Errorf("get after restart: exit code = %d, want 0 (stderr=%q)", code, stderr)
	}
	if stdout != "bar\n" {
		t.Errorf("get after restart: stdout = %q, want %q", stdout, "bar\n")
	}
}

// TestPersistence_DeleteThenRestartThenGet checks that a delete is also
// durable across a restart, not just a set.
func TestPersistence_DeleteThenRestartThenGet(t *testing.T) {
	dataDir := t.TempDir()

	if _, stderr, code := runKVS(t, "--data-dir", dataDir, "set", "foo", "bar"); code != 0 {
		t.Fatalf("set exit code = %d, want 0 (stderr=%q)", code, stderr)
	}
	if _, stderr, code := runKVS(t, "--data-dir", dataDir, "delete", "foo"); code != 0 {
		t.Fatalf("delete exit code = %d, want 0 (stderr=%q)", code, stderr)
	}

	stdout, stderr, code := runKVS(t, "--data-dir", dataDir, "get", "foo")
	if code == 0 {
		t.Errorf("get after restart: exit code = 0, want non-zero (stdout=%q)", stdout)
	}
	if stderr != "key not found\n" {
		t.Errorf("get after restart: stderr = %q, want %q", stderr, "key not found\n")
	}
}

// TestWithoutDataDir_NotPersistedAcrossProcesses pins down the "unchanged
// default behavior" acceptance criterion: without --data-dir, each kvs
// invocation is still its own in-memory store, so a value set by one
// process is not visible to the next.
func TestWithoutDataDir_NotPersistedAcrossProcesses(t *testing.T) {
	if _, stderr, code := runKVS(t, "set", "foo", "bar"); code != 0 {
		t.Fatalf("set exit code = %d, want 0 (stderr=%q)", code, stderr)
	}

	stdout, stderr, code := runKVS(t, "get", "foo")
	if code == 0 {
		t.Errorf("get in a fresh process: exit code = 0, want non-zero (stdout=%q)", stdout)
	}
	if stderr != "key not found\n" {
		t.Errorf("get in a fresh process: stderr = %q, want %q", stderr, "key not found\n")
	}
}

// TestDataDir_AutoCreated checks that a --data-dir that doesn't exist yet
// is created rather than requiring the user to pre-create it.
func TestDataDir_AutoCreated(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "does-not-exist-yet")

	stdout, stderr, code := runKVS(t, "--data-dir", dataDir, "set", "foo", "bar")
	if code != 0 {
		t.Fatalf("set exit code = %d, want 0 (stdout=%q stderr=%q)", code, stdout, stderr)
	}
	if _, err := os.Stat(dataDir); err != nil {
		t.Errorf("--data-dir was not auto-created: %v", err)
	}
}

// TestDataDir_PointsAtRegularFile_FailsLoudly checks that kvs refuses to
// run, with a clear non-zero exit, when --data-dir names an existing
// regular file instead of a directory.
func TestDataDir_PointsAtRegularFile_FailsLoudly(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "im-a-file")
	if err := os.WriteFile(dataDir, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("failed to seed regular file: %v", err)
	}

	stdout, stderr, code := runKVS(t, "--data-dir", dataDir, "set", "foo", "bar")
	if code == 0 {
		t.Errorf("exit code = 0, want non-zero (stdout=%q)", stdout)
	}
	if stderr == "" {
		t.Error("stderr is empty, want a clear error message")
	}
}

// TestDataDir_CorruptStoreFile_FailsLoudly checks that kvs refuses to run,
// with a clear non-zero exit, rather than silently discarding data, when
// store.json exists but isn't valid JSON.
func TestDataDir_CorruptStoreFile_FailsLoudly(t *testing.T) {
	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, "store.json"), []byte("{not valid json"), 0o600); err != nil {
		t.Fatalf("failed to seed corrupt store file: %v", err)
	}

	stdout, stderr, code := runKVS(t, "--data-dir", dataDir, "get", "foo")
	if code == 0 {
		t.Errorf("exit code = 0, want non-zero (stdout=%q)", stdout)
	}
	if stderr == "" {
		t.Error("stderr is empty, want a clear error message")
	}
}
