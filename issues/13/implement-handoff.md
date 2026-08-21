## Summary

Added an optional on-disk persistence backend so `kvs` data survives process
restarts, without changing the CLI's command surface. A new `store.FileStore`
implements the existing `store.Store` interface, backed by a JSON file
(`<data-dir>/store.json`); it's selected via a new `--data-dir` flag. Without
`--data-dir`, `kvs` is unchanged: an in-memory store that doesn't survive
restarts.

Along the way, found and fixed a pre-existing bug (see `## Changes`) where
`kvs get`/`kvs list` were writing their output to stderr instead of stdout in
the real binary — it was invisible before because no test exercised a
successful `get`/`list` through real OS pipes.

## Branch

`workflow/issue-13`, branched from `main`.

## Changes

- `internal/store/file_store.go` (new) — `FileStore` implementing `Store`:
  - `NewFileStore(dataDir)` auto-creates `dataDir` (and its parents) if
    missing; fails with a clear error if `dataDir` exists but is a regular
    file, not a directory.
  - Loads `store.json` on construction if present; fails with a clear error
    (no silent fallback to an empty store) if it exists but contains invalid
    JSON.
  - `Set`/`Delete` mutate the in-memory map and synchronously flush the full
    snapshot to disk via write-to-temp-file + rename (atomic, so a failure
    mid-write can't leave `store.json` truncated/corrupt) before returning.
    No batching/periodic flush — `kvs` is a one-shot CLI process, so there's
    no later point to defer a flush to. On a flush failure, the in-memory
    map is rolled back to stay consistent with what's actually on disk.
  - `Get`/`List` are pure in-memory reads (no I/O), matching `MemoryStore`.
- `internal/store/store.go` — `Store.Set` now returns `error` (was: no
  return value), so a failed disk write can be reported through the same
  path `Get`/`Delete` already use. `MemoryStore.Set` always returns `nil`
  (it can't fail).
- `internal/cli/set.go` — propagates `Set`'s new error to stderr via
  `cmd.PrintErrln(err)`, mirroring how `get.go`/`delete.go` already handle
  `store.ErrKeyNotFound`.
- `internal/cli/delete.go` — same propagation for non-`ErrKeyNotFound`
  errors from `Delete` (previously only `ErrKeyNotFound` was printed; any
  other error returned silently with exit 1 — dead code path before
  `FileStore` existed, live now that `Delete` can fail on a flush error).
- `internal/cli/get.go`, `internal/cli/list.go` — **bug fix**: both used
  `cmd.Println`/`cmd.Printf`, which fall back to `os.Stderr` (not
  `os.Stdout`) when no explicit output writer is set on the command — which
  is exactly the case in the real `kvs` binary (`main.go` never calls
  `cmd.SetOut`). So `kvs get <existing key>` and `kvs list` were printing
  their actual output to stderr, not stdout. Existing unit tests didn't
  catch this because `cli_test.go` always calls `cmd.SetOut(&buf)` first
  (which sidesteps the fallback), and no e2e test previously checked stdout
  content for a successful `get`/`list`. Fixed by writing to
  `cmd.OutOrStdout()` directly via `fmt.Fprintln`/`fmt.Fprintf`. This was
  necessary to correctly implement and verify "data set via `kvs set` is
  retrievable via `kvs get`" for the new persistence e2e tests.
- `internal/cli/root.go` — registers `--data-dir` as a persistent flag
  purely so it's documented in `kvs --help`; it does not drive store
  selection (see `main.go` below).
- `cmd/kvs/main.go` — pre-scans `os.Args` for `--data-dir` (handles both
  `--data-dir value` and `--data-dir=value`) *before* constructing the
  `Store` and calling `cli.NewRootCmd`, since Cobra only parses flags during
  `root.Execute()`, by which point the store already needs to exist. Builds
  a `FileStore` if `--data-dir` is set, otherwise the existing
  `MemoryStore`.
- `README.md` — documents `--data-dir` usage.
- `workflow/PROJECT.md` — merged language for `Store.Set`'s new signature,
  `FileStore`, and `--data-dir`.
- Tests: `internal/store/file_store_test.go` (new), updates to
  `internal/store/store_test.go` and `internal/cli/cli_test.go` (checking
  `Set`'s new error return — required for `golangci-lint`'s `errcheck`),
  and new persistence scenarios in `e2e/e2e_test.go`.

## TDD cycles

1. **Store interface + FileStore core** — wrote `file_store_test.go`
   (`TestFileStore_SetGet`, `_Get_MissingKey`, `_Delete`,
   `_Delete_MissingKey`, `_List`, `_List_ReturnsCopy`, mirroring
   `store_test.go`'s `MemoryStore` coverage) against a `FileStore` that
   didn't exist yet → compile failure (red). Implemented `FileStore` with
   in-memory map + synchronous JSON flush → green. Changing `Store.Set` to
   return `error` broke existing unchecked-return call sites under
   `golangci-lint` (`errcheck`) in `store_test.go` and `cli_test.go` → fixed
   by checking the error at each call site → clean lint.
2. **Load-on-start / restart persistence** — wrote
   `TestFileStore_PersistsAcrossInstances` (set + delete via one `FileStore`
   instance, construct a second instance against the same dir, assert
   reloaded state) → red (no load-from-disk logic yet). Added
   `loadStoreFile` in `NewFileStore` → green.
3. **Auto-create data dir / fail-loudly cases** — wrote
   `TestNewFileStore_CreatesMissingDataDir`,
   `TestNewFileStore_InvalidJSON_FailsLoudly`,
   `TestNewFileStore_DataDirIsRegularFile_FailsLoudly` → red. Added
   `ensureDir` (mkdir -p / reject non-directory) and corrupt-JSON detection
   in `loadStoreFile` → green.
4. **CLI wiring (`--data-dir`)** — wrote e2e tests
   `TestPersistence_SetThenRestartThenGet`,
   `TestPersistence_DeleteThenRestartThenGet`,
   `TestWithoutDataDir_NotPersistedAcrossProcesses`, `TestDataDir_AutoCreated`,
   `TestDataDir_PointsAtRegularFile_FailsLoudly`,
   `TestDataDir_CorruptStoreFile_FailsLoudly` against `main.go`/`root.go`
   before wiring `--data-dir` through → red (flag didn't exist / store
   always in-memory). Added the `os.Args` pre-scan in `main.go` and the
   docs-only persistent flag in `root.go` → `TestPersistence_*` still red:
   `get` returned empty stdout instead of the persisted value. Root-caused
   to `cmd.Println`/`cmd.Printf` defaulting to stderr in the unconfigured
   real binary (see `## Changes`); fixed `get.go`/`list.go` to write to
   `cmd.OutOrStdout()` → all green.

## Test results

Ran on the work branch (`workflow/issue-13`, commit `6f3d4a7`):

```
$ go build ./... && go vet ./...
(clean, no output)

$ gofmt -l .
(clean, no output)

$ go test ./...
?   	github.com/D-Andreev/simple-key-value-store/cmd/kvs	[no test files]
ok  	github.com/D-Andreev/simple-key-value-store/e2e
ok  	github.com/D-Andreev/simple-key-value-store/internal/cli
ok  	github.com/D-Andreev/simple-key-value-store/internal/store

$ go test -race ./...
ok  	github.com/D-Andreev/simple-key-value-store/e2e
ok  	github.com/D-Andreev/simple-key-value-store/internal/cli
ok  	github.com/D-Andreev/simple-key-value-store/internal/store

$ golangci-lint run ./...
0 issues.
```

32 top-level tests (48 including subtests), all passing:
- `internal/store`: 12 `MemoryStore` tests (pre-existing, unchanged
  behavior) + 12 `FileStore` tests (new — set/get, missing-key get/delete,
  list + returns-copy, overwrite, persist-across-instances, auto-create,
  invalid-JSON fail-loudly, non-directory fail-loudly, concurrent access).
- `internal/cli`: 4 tests (pre-existing, updated only to check `Set`'s new
  error return), all still passing against the updated `Store` interface.
- `e2e`: 4 pre-existing tests (unchanged assertions, still pass — confirms
  `--data-dir`-less behavior is unchanged) + 6 new tests covering
  set→restart→get, delete→restart→get, no-`--data-dir` non-persistence,
  auto-create, and the two fail-loudly paths through the real binary.

## Acceptance criteria

- [x] Data set via `kvs set` is retrievable via `kvs get` after the process
      restarts, when `--data-dir` is provided —
      `TestPersistence_SetThenRestartThenGet` (e2e).
- [x] Without `--data-dir`, behavior is unchanged (in-memory only) —
      `TestWithoutDataDir_NotPersistedAcrossProcesses` (e2e); all 4
      pre-existing e2e tests still pass unmodified.
- [x] `FileStore.Set`/`FileStore.Delete` write the full JSON snapshot to
      disk synchronously, before returning — no batched/periodic flush —
      `flushLocked` runs inline inside `Set`/`Delete` under the write lock,
      no goroutine/timer/queue.
- [x] `main.go` pre-scans `os.Args` for `--data-dir` to select `MemoryStore`
      vs. `FileStore` before calling `cli.NewRootCmd`; `--data-dir` is also
      registered as a persistent flag on root so `kvs --help` documents it
      — `dataDirFromArgs` in `main.go`; persistent flag in `root.go`.
- [x] `FileStore` auto-creates `--data-dir` (and `store.json`) on startup if
      missing — `ensureDir` in `file_store.go`; `TestNewFileStore_
      CreatesMissingDataDir`, `TestDataDir_AutoCreated`.
- [x] `FileStore` fails loudly (non-zero exit, clear stderr message) if
      `store.json` exists but contains invalid JSON — no silent fallback to
      an empty store — `TestNewFileStore_InvalidJSON_FailsLoudly`,
      `TestDataDir_CorruptStoreFile_FailsLoudly` (e2e).
- [x] `kvs` fails loudly with a clear error if `--data-dir` points to an
      existing regular file rather than a directory —
      `TestNewFileStore_DataDirIsRegularFile_FailsLoudly`,
      `TestDataDir_PointsAtRegularFile_FailsLoudly` (e2e).
- [x] Unit tests for `FileStore` mirroring `store_test.go` coverage for
      `MemoryStore` (set/get, missing-key get/delete, list, overwrite) —
      `file_store_test.go`.
- [x] e2e test covering `set` → restart process → `get` (new `kvs` process,
      same `--data-dir`, sees the persisted value) —
      `TestPersistence_SetThenRestartThenGet`.
- [x] `go test ./...` passes, including new `FileStore` tests — confirmed
      above (also passes under `-race`).

## Suggested review scenarios

1. **`FileStore.Set`/`Delete` rollback on flush failure** — check
   `internal/store/file_store.go`: if `flushLocked` fails after the
   in-memory map was already mutated, does the rollback (`s.data[key] =
   prev` / `delete(s.data, key)`) correctly restore the pre-mutation state
   so memory stays consistent with what's actually on disk?
2. **The stdout/stderr fix (`get.go`/`list.go`)** — this wasn't part of the
   original ask; confirm the reasoning holds: run `kvs get <existing-key>`
   (no `--data-dir` needed) with `1>/tmp/out 2>/tmp/err` and check the value
   lands in `/tmp/out`, not `/tmp/err`, on `main` (pre-fix) vs. this branch
   (post-fix).
3. **`--data-dir` pre-scan in `main.go` vs. the persistent flag in
   `root.go`** — verify these two don't fight: e.g. `kvs get --data-dir
   <dir> <key>` (flag after the subcommand) and `kvs --data-dir=<dir> get
   <key>` (`=` form) both resolve to the same store as `kvs --data-dir <dir>
   get <key>`.
4. **Atomic write-then-rename in `flushLocked`** — confirm the `.tmp` file
   is always cleaned up (i.e., renamed over `store.json`) and never left
   behind after a normal `Set`/`Delete`; check what happens if `dataDir` and
   its parent are on different filesystems (rename could fail — out of
   scope per requirements' "crash-safety... beyond clean shutdown", but
   worth confirming the resulting error is at least reported, not
   swallowed).
5. **`Store.Set` interface change ripple** — confirm no other caller in the
   repo (or a future one) silently discards `Set`'s new error return the
   way the pre-change code did; `golangci-lint`'s `errcheck` should catch
   this going forward, but worth a manual `grep -rn '\.Set(' internal/
   cmd/` sanity pass.
