# Task: issue-13

## Title
[Feature]: On disk persistance

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
