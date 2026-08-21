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
