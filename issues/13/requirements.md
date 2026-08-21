# Requirements: issue-13

## Original ask
Summary

Add an optional on-disk persistence backend so key/value data survives process restarts, without changing the CLI's command surface.

Motivation

MemoryStore currently loses all data when the process exits, which limits kvs to ephemeral/scripting use. Users running it as a lightweight local store need data to persist across restarts. Since storage is already abstracted behind the store.Store interface, this can be added as a new implementation without touching CLI command logic.

Proposed solution

Add a FileStore implementation of store.Store that:
- Loads existing data from a file on startup (if present).
- Persists to disk on every mutating call (Set/Delete), or via periodic/batched flush — TBD based on perf needs.
- Uses a simple encoding (JSON to start; can revisit for a binary/WAL format later).

CLI changes:
kvs --data-dir ~/.kvs set foo bar   # persists to ~/.kvs/store.json
kvs --data-dir ~/.kvs get foo       # reads back after restart
- Default --data-dir could fall back to in-memory-only (current behavior) if unset, to avoid breaking existing scripts/tests.
- root.go wires up the chosen Store implementation based on the flag.

Alternatives considered

- Embedded DB (BoltDB/BadgerDB) — more robust (real WAL, crash-safety) but adds a dependency; can be a follow-up if JSON-file proves too slow/unsafe.
- Always-on persistence with no flag — rejected for now to keep the default zero-config, in-memory experience for quick/scripted use.

Scope

- [ ] store.FileStore implementing store.Store, backed by a JSON file
- [ ] Load-on-start / flush-on-write behavior
- [ ] --data-dir (or similar) CLI flag wired into root.go
- [ ] Unit tests for FileStore (mirroring store_test.go coverage for MemoryStore)
- [ ] e2e test covering set → restart process → get

Out of scope

- Crash-safety guarantees beyond "don't lose data on clean shutdown" (e.g. full WAL/fsync semantics)
- Concurrent multi-process access to the same data file
- Migration tooling between storage backends

Acceptance criteria

- [ ] Data set via kvs set is retrievable via kvs get after the process restarts, when --data-dir is provided
- [ ] Without --data-dir, behavior is unchanged (in-memory only, matches current README)
- [ ] go test ./... passes, including new FileStore tests

Additional context

Related: internal/store/store.go defines the Store interface this would implement; internal/cli/root.go currently constructs store.NewMemoryStore() directly.

## Clarifications
| # | Question | Answer | Recommended |
|---|----------|--------|-------------|
| 1 | `kvs` is a one-shot CLI (each invocation is a fresh process that exits after one command), not a daemon. Should `FileStore` flush synchronously to disk on every `Set`/`Delete`, before the command returns, rather than periodic/batched flush? | Yes — synchronous flush on every mutating call. No batching/async flush; the "periodic/batched flush — TBD" option in the original ask is dropped since there's no live process to defer a flush to. | Yes |
| 2 | `main.go` currently constructs the store *before* `cli.NewRootCmd(s)`, but Cobra only parses flags during `root.Execute()`, so `NewRootCmd` can't know `--data-dir` at store-construction time. How should `--data-dir` be threaded through to select `MemoryStore` vs. `FileStore`? | Manual pre-scan of `os.Args` for `--data-dir` in `main.go` *before* building the store and calling `NewRootCmd`, keeping `--data-dir` also registered as a persistent flag on root for `--help`/usage text only (not as the source of truth for store selection). No restructuring of `newSetCmd`/`newGetCmd`/etc. to take a lazily-resolved store. | Manual pre-scan + persistent flag for docs only |

## Acceptance criteria
- [ ] Data set via `kvs set` is retrievable via `kvs get` after the process restarts, when `--data-dir` is provided
- [ ] Without `--data-dir`, behavior is unchanged (in-memory only, matches current README)
- [ ] `FileStore.Set`/`FileStore.Delete` write the full JSON snapshot to disk synchronously, before returning — no batched/periodic flush
- [ ] `main.go` pre-scans `os.Args` for `--data-dir` to select `MemoryStore` vs. `FileStore` before calling `cli.NewRootCmd`; `--data-dir` is also registered as a persistent flag on root so `kvs --help` documents it
- [ ] `go test ./...` passes, including new `FileStore` tests

## Approved by human
- [ ] Pending — say `approve requirements` in the session when ready
