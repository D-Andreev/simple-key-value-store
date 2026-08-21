# Implement handoff: issue-11

## Summary
Investigation confirmed what clarify already suspected: `Set`/`Get`/`Delete`/`List` were already implemented on `main` in `internal/store/store.go`, with `string` keys/values, `ErrKeyNotFound` semantics on `Get`/`Delete`, and a `sync.RWMutex` guarding all access — satisfying every acceptance criterion from the issue except one. The one real gap was a concurrency test: nothing exercised `MemoryStore` from multiple goroutines, and CI runs `go test ./...` without `-race`. This implementation adds exactly that test; no store or CLI production code changed.

## Branch
`workflow/issue-11` (base: `main`)

## Changes
- `internal/store/store_test.go` — added `TestMemoryStore_ConcurrentAccess`: 50 goroutines × 100 iterations each of `Set`/`Get`/`Delete`/`List` against a shared `MemoryStore`, run under `go test -race` to demonstrate the store's locking holds up under concurrent access. No assertions on interleaved values (no deterministic outcome to check) — its job is to give the race detector something to catch.
- No other files changed. No new store/CLI methods added; `Set`/`Get`/`Delete`/`List` signatures unchanged (`string` in/out), no `Exists` method added, `Delete` on a missing key still returns `ErrKeyNotFound` — matching the approved requirements exactly.
- `workflow/PROJECT.md` — not touched; `language.md` for this issue had no new domain terms to merge (existing `## Language` section already documents `Store`, in-memory store, and `kvs`).

## TDD cycles
Since the store implementation already existed and passed, TDD here applied to the one new test, not new production code:
1. **Red-ish baseline**: ran the full existing suite (`go test ./... -v`) before any change — all green, confirming no regression risk and that no concurrency test existed yet.
2. **Green**: wrote `TestMemoryStore_ConcurrentAccess` against the existing `MemoryStore` (already mutex-protected) and ran `go test ./... -race -v` — passed immediately, confirming the existing locking is correct. No production code needed to change to make it pass, which is itself the expected outcome per requirements clarification #4 (issue was already-satisfied).

## Test results
Ran on the work branch, all passing:
- `go build ./...` — clean
- `go vet ./...` — clean
- `golangci-lint run ./...` — `0 issues.`
- `gofmt -l .` — no output (all files formatted)
- `go test ./... -race -v`:
  - `internal/store`: `TestMemoryStore_SetGet` (3 subtests), `TestMemoryStore_Get_MissingKey`, `TestMemoryStore_Delete`, `TestMemoryStore_Delete_MissingKey`, `TestMemoryStore_List` (2 subtests), `TestMemoryStore_List_ReturnsCopy`, `TestMemoryStore_ConcurrentAccess` (new) — all `PASS`
  - `internal/cli`: `TestSetCmd`, `TestGetCmd` (2 subtests), `TestDeleteCmd` (2 subtests), `TestListCmd` (2 subtests) — all `PASS`
  - `e2e`: `TestSet_Success`, `TestGet_MissingKey`, `TestDelete_MissingKey`, `TestList_EmptyStore` — all `PASS`
  - Overall: `ok` for all three packages, no race conditions detected, no failures.

## Acceptance criteria
- [x] Deleting a key removes it (subsequent `Get` returns `ErrKeyNotFound`); deleting a missing key also returns `ErrKeyNotFound` (unchanged from current behavior) — verified via existing `TestMemoryStore_Delete` / `TestMemoryStore_Delete_MissingKey`, no code changed.
- [x] `Set`/`Get` keep `string` keys and values — no signature change to `[]byte` — confirmed unchanged in `internal/store/store.go`.
- [x] No `Exists` method added — deferred out of this issue — confirmed, not present.
- [x] Implement phase re-runs the existing store/CLI test suite and confirms it passes; no new store or CLI code is written unless a genuine gap is found — re-ran full suite (see Test results); the one gap found (missing concurrency test) was closed with a test only, no production code.
- [x] Add a concurrent-access test for `MemoryStore` (parallel goroutines calling `Set`/`Get`/`Delete`), runnable under `go test -race`, to demonstrate the concurrency-safety AC — added `TestMemoryStore_ConcurrentAccess`, passes under `go test ./... -race`.

## Suggested review scenarios
1. Confirm `internal/store/store.go` is untouched in the diff (`git diff main...workflow/issue-11 -- internal/store/store.go` should be empty) — this PR should be test-only.
2. Read `TestMemoryStore_ConcurrentAccess` and check it actually exercises all three mutating/reading operations (`Set`, `Get`, `Delete`, plus `List`) concurrently, and that removing the `sync.RWMutex` from `MemoryStore` (a quick local experiment) makes `go test -race ./internal/store/...` fail — confirming the test has teeth.
3. Run `go test ./... -race` locally and confirm it's green, matching the reported results.
4. Sanity-check the requirements reconciliation: re-read the issue's literal proposal (`[]byte` values, idempotent `Delete`, optional `Exists`) against `requirements.md`'s clarified acceptance criteria, and confirm the shipped code matches the clarified version, not the issue's literal wording.
