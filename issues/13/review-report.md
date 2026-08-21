# Review Report

**Fresh-eyes:** judgments based on artifacts and diff only (`main...HEAD` on `workflow/issue-13`).

## Verdict
APPROVE WITH NOTES

## Scenario verification

### Scenarios verified

| # | Scenario | Method | Result | Notes |
|---|----------|--------|--------|-------|
| 1 | `kvs --data-dir <dir> set foo bar` then a fresh process `kvs --data-dir <dir> get foo` returns `bar` | code trace | pass | `FileStore.Set` flushes synchronously; `NewFileStore` loads `store.json` on construction. Matches `TestPersistence_SetThenRestartThenGet`. |
| 2 | Without `--data-dir`, behavior is unchanged (in-memory only) | code trace | pass | `main.go`'s `newStore` returns `store.NewMemoryStore()` when `dataDirFromArgs` yields `""`; `MemoryStore` is untouched except `Set`'s new signature. |
| 3 | `--data-dir` points at an existing regular file | code trace | pass | `ensureDir` stats the path, returns a clear error (`"data dir %q is not a directory"`) without ever touching `store.json`. |
| 4 | `store.json` exists but contains invalid JSON | code trace | pass | `loadStoreFile` unmarshal error is wrapped and returned from `NewFileStore` — no silent empty-store fallback. |
| 5 | `Set`/`Delete` flush failure rollback | code trace only (no test) | pass (by inspection) | `Set` captures `prev`/`existed` before mutating, restores on `flushLocked` error; `Delete` captures `prev` before removing, restores on error. Logic is correct by trace, but nothing in `file_store_test.go` exercises an actual flush failure (e.g. read-only data dir) — see Suggestions. |
| 6 | `--data-dir` after the subcommand / `--data-dir=<dir>` form | code trace only (no test) | pass (by inspection) | `dataDirFromArgs` scans all of `os.Args[1:]` positionally-independent, and pflag's default interspersed parsing lets a persistent flag appear after a subcommand — so both forms should resolve identically. Not backed by an automated test despite being called out as a suggested review scenario. See Suggestions. |
| 7 | Atomic write-then-rename leaves no debris on the happy path | code trace | pass | `flushLocked` writes to `store.json.tmp` then renames over `store.json`; on success no `.tmp` remains. |
| 8 | `Store.Set` interface-change ripple — no other caller silently discards the new error | code trace (`grep -rn '\.Set(' internal/ cmd/ e2e/`) | pass | Only non-test call site is `internal/cli/set.go`, which checks and propagates the error. Test call sites all check it too (required by `errcheck`). |

### Requirements coverage
- [x] Data set via `kvs set` is retrievable via `kvs get` after restart, with `--data-dir` — `TestPersistence_SetThenRestartThenGet`.
- [x] Without `--data-dir`, behavior unchanged — `TestWithoutDataDir_NotPersistedAcrossProcesses`; pre-existing e2e tests untouched.
- [x] Synchronous flush on `Set`/`Delete`, no batching — `flushLocked` runs inline under the write lock.
- [x] `main.go` pre-scan + docs-only persistent flag on root — present in both files, wired as described.
- [x] Auto-create `--data-dir` if missing — `ensureDir`/`TestNewFileStore_CreatesMissingDataDir`. (`store.json` itself is created lazily on first flush rather than at construction — a reasonable reading of "on startup" given clarify Q3 only committed to auto-creating the *directory*; not a gap.)
- [x] Fail loudly on invalid JSON — `TestNewFileStore_InvalidJSON_FailsLoudly` / `TestDataDir_CorruptStoreFile_FailsLoudly`.
- [x] Fail loudly when `--data-dir` is a regular file — `TestNewFileStore_DataDirIsRegularFile_FailsLoudly` / `TestDataDir_PointsAtRegularFile_FailsLoudly`.
- [x] Unit tests mirroring `store_test.go` coverage — `file_store_test.go` covers set/get, missing-key get/delete, list (+ copy semantics), overwrite, persistence-across-instances, auto-create, both fail-loudly paths, and concurrent access under `-race`.
- [x] e2e set → restart → get — `TestPersistence_SetThenRestartThenGet`.
- [x] `go test ./...` passes — cited from `implement-handoff.md`, not re-run per review policy.

### Issues found (from review)
- 🟡 Minor: `FileStore.Set`/`Delete` flush-failure rollback path (the one piece of new logic that isn't a straight-line write) has no automated test — see Suggestions.
- 🟡 Minor: `--data-dir` positioned after the subcommand, and the `--data-dir=<dir>` form, aren't covered by any e2e test even though `implement-handoff.md` calls out verifying exactly this as a suggested review scenario — see Suggestions.

### Implement test results (cited, not re-run)
- `go build ./...`, `go vet ./...`, `gofmt -l .` clean.
- `go test ./...` and `go test -race ./...`: all packages pass (`e2e`, `internal/cli`, `internal/store`).
- `golangci-lint run ./...`: 0 issues.
- 32 top-level tests (48 incl. subtests): 12 pre-existing `MemoryStore` tests + 12 new `FileStore` tests + 4 pre-existing `internal/cli` tests + 4 pre-existing e2e tests (unchanged assertions) + 6 new persistence e2e tests.

### Gaps in test coverage
- No test exercises the flush-failure rollback branch in `Set`/`Delete` (e.g. a read-only data dir forcing `flushLocked` to fail) — logic verified correct by manual trace only.
- No test exercises `--data-dir` in the "after the subcommand" or `--data-dir=<dir>` forms, despite the dual-parsing design (manual `os.Args` pre-scan in `main.go` + a separate Cobra-registered persistent flag in `root.go`) being exactly the kind of thing a future refactor could silently desync.
- A dangling `store.json.tmp` isn't cleaned up if `os.WriteFile`/`os.Rename` fails mid-flush in `flushLocked` — explicitly out of scope per requirements ("crash-safety guarantees beyond clean shutdown"), so not counted as a finding, just noting it's untested and undocumented as a known limitation.

### Tests/build in review
- **Not run** — review is diff + code reading only; implement phase owns execution.

## Principles review

### Summary
Clean, well-scoped implementation. `FileStore` mirrors `MemoryStore`'s locking discipline, uses an atomic write-then-rename flush, and rolls back in-memory state on a flush failure so memory never disagrees with disk. The `Store.Set` signature change is threaded through every call site correctly. The unrelated stdout/stderr bug fix in `get.go`/`list.go` is well-justified (root-caused, explained in `implement-handoff.md`, necessary to actually verify the new e2e persistence tests) rather than opportunistic scope creep.

### Critical (must fix)
- None.

### Suggestions (should consider)
- Add a unit test that forces `flushLocked` to fail (e.g. `os.Chmod` the data dir to read-only before a `Set`/`Delete`, restoring permissions in cleanup) and asserts the in-memory map is unchanged afterward — the rollback branch is the one piece of state-machine logic in this PR that isn't exercised by any test today.
- Add one e2e case for `kvs get --data-dir <dir> <key>` (flag after the subcommand) or `kvs --data-dir=<dir> get <key>` — `implement-handoff.md` names this as a suggested review scenario and the dual-source-of-truth design (`main.go`'s manual pre-scan vs. `root.go`'s registered flag) is fragile to a future refactor of either side without a regression test pinning the contract.

### Nice to have
- A short doc comment on `flushLocked` (or nearby) noting that a `store.json.tmp` can be left behind if `os.WriteFile`/`os.Rename` fails mid-flush, and that cleanup of it is intentionally out of scope per the "crash-safety beyond clean shutdown" requirements carve-out — would save a future reader from wondering if it's an oversight.

### Scenario overlap avoided
- Did not re-verify `MemoryStore` behavior (pre-existing, only touched for the `Set` signature change) beyond confirming its `Set` always returns `nil`.
- Did not re-derive the stdout/stderr bug fix from scratch — `implement-handoff.md`'s root-cause explanation (`cmd.Println`/`cmd.Printf` default to stderr without `cmd.SetOut`) checks out against `get.go`/`list.go`'s diff and matches Cobra's documented behavior.

### Principles applied
- Security: `store.json` written with `0o600`; no secrets/PII involved; no new attack surface (single local file, no network).
- Design/maintainability: `Store` interface stays minimal; `FileStore` and `MemoryStore` are symmetric and independently testable; error wrapping (`fmt.Errorf("...: %w", err)`) is idiomatic and consistent throughout.
- Conventions: doc comments on every exported type/function follow the existing repo style; error handling mirrors the established `ErrKeyNotFound` pattern in `get.go`/`delete.go`.

## Recommendation
Shippable as-is — no correctness blockers found by diff/code reading. The two suggestions (rollback test, flag-position/`=`-form test) are worth picking up as fast follow-ups given they're both scenarios the implement phase itself flagged as needing verification but didn't turn into tests.
