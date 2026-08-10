# Implement Handoff: issue-8

## Summary
Scaffolded the `kvs` CLI per the approved requirements: a Go/Cobra command-line
tool with `set`/`get`/`delete`/`list` subcommands wired against a real
in-memory `Store`. No persistence, no input validation beyond what Cobra
enforces (arg counts) — this is intentionally the scaffolding step, with
business-logic polish deferred to later issues. CI (build/vet/lint/test) and
`golangci-lint` are configured and green.

## Branch
`workflow/issue-8` (base: `main`)

## Changes
- `go.mod` / `go.sum` — module `github.com/D-Andreev/simple-key-value-store`, Go 1.24.7 (latest stable at implement time, no pinned minor version beyond that), dependency on `github.com/spf13/cobra`.
- `internal/store/store.go` — `Store` interface (`Set`, `Get`, `Delete`, `List`) and `MemoryStore`, a mutex-guarded in-memory implementation. `List()` returns a defensive copy. `ErrKeyNotFound` sentinel error.
- `internal/cli/root.go` — `NewRootCmd(store.Store) *cobra.Command`, wires all four subcommands; `SilenceUsage`/`SilenceErrors` set so error output stays clean (no cobra usage dump on business errors).
- `internal/cli/set.go`, `get.go`, `delete.go`, `list.go` — one Cobra subcommand each, injected with the `Store`.
  - `set <key> <value>` — writes to the store, no output on success.
  - `get <key>` — prints the value + newline on success; on missing key, prints `key not found` to stderr and returns an error (non-zero exit).
  - `delete <key>` — same missing-key behavior as `get`; no output on success.
  - `list` — prints `key=value` one per line, sorted by key; empty store prints nothing.
- `cmd/kvs/main.go` — entrypoint; builds a `MemoryStore`, runs the root command, `os.Exit(1)` on error.
- `internal/store/store_test.go`, `internal/cli/cli_test.go` — table-driven unit tests.
- `e2e/e2e_test.go` — builds the `kvs` binary in `TestMain` and drives it as a subprocess, asserting stdout/stderr/exit code per command.
- `.golangci.yml` — `golangci-lint` v2 config, default linter set.
- `.github/workflows/ci.yml` — GitHub Actions: `go build`, `go vet`, `golangci-lint-action`, `go test ./...` on push/PR to `main`.
- `.gitignore` — ignore build artifacts.
- `README.md` — build/usage/test instructions.
- `workflow/PROJECT.md` — merged approved `## Language` section (Store, in-memory store, kvs binary name).

## TDD cycles
1. **Store** — wrote `internal/store/store_test.go` (Set/Get, missing-key errors on Get/Delete, List sorted-map equality, List returns a defensive copy) against an undefined `MemoryStore`/`NewMemoryStore`/`ErrKeyNotFound` → confirmed red (`undefined: NewMemoryStore` etc.) → implemented `internal/store/store.go` → green (`go test ./internal/store/...` all pass).
2. **CLI commands** — wrote `internal/cli/cli_test.go` exercising the root command end-to-end via `cmd.Execute()` with buffered stdout/stderr, against an undefined `NewRootCmd` → confirmed red (`undefined: NewRootCmd`) → implemented `root.go` + `set.go`/`get.go`/`delete.go`/`list.go` → green (`go test ./internal/cli/...` all pass, including missing-key stderr assertions and sorted `list` output).
3. **E2E** — wrote `e2e/e2e_test.go` (builds the real binary, runs it as a subprocess) after `cmd/kvs/main.go` existed, to verify the compiled binary's exit codes and stderr text end-to-end → green on first run.
4. **Lint fixes** — `golangci-lint run` flagged an unchecked `os.RemoveAll` in the e2e test's cleanup and an unused test helper in `cli_test.go`; fixed both, re-ran to a clean `0 issues`.

## Test results
```
$ go build ./...
(clean)

$ go vet ./...
(clean)

$ gofmt -l .
(no output — all formatted)

$ go test ./... -race
?   	github.com/D-Andreev/simple-key-value-store/cmd/kvs	[no test files]
ok  	github.com/D-Andreev/simple-key-value-store/e2e	1.388s
ok  	github.com/D-Andreev/simple-key-value-store/internal/cli	1.022s
ok  	github.com/D-Andreev/simple-key-value-store/internal/store	1.022s

$ golangci-lint run ./...
0 issues.
```

## Acceptance criteria
- [x] CLI exposes subcommands `set <key> <value>`, `get <key>`, `delete <key>`, `list`, each performing the corresponding basic operation against an in-memory `Store` (no input validation or polished error messages yet); string keys, string values
- [x] Store logic sits behind a `Store` interface with an in-memory implementation
- [x] Project laid out as `cmd/kvs/main.go`, `internal/cli/`, `internal/store/`; module path `github.com/D-Andreev/simple-key-value-store`
- [x] Unit tests (stdlib `testing`, table-driven) cover the `Store` implementation and CLI commands
- [x] E2E tests build the binary and exercise it as a subprocess, asserting stdout/exit code per command
- [x] `golangci-lint` configured and passing
- [x] GitHub Actions workflow runs `go build`, `go vet`, `golangci-lint run`, `go test ./...` on push/PR to `main`
- [x] `get`/`delete` on a missing key exit non-zero with a `key not found` message; success paths exit 0
- [x] `list` prints one `key=value` per line, sorted by key, for deterministic output
- [x] `go.mod` targets the latest stable Go available at implement time (no pinned minor version)

## Suggested review scenarios
1. `go build -o kvs ./cmd/kvs && ./kvs set foo bar` — should exit 0 with no output (note: state is per-process; a following `./kvs get foo` in a *separate* invocation won't find it — no persistence yet, expected per spec).
2. `./kvs get missing` — expect exit 1, stderr `key not found`.
3. `./kvs delete missing` — expect exit 1, stderr `key not found`.
4. `./kvs list` on a fresh store — expect exit 0, empty stdout.
5. In `internal/cli/cli_test.go`, check the `list` sorted-output case (`b`, `a`, `c` seeded → `a=1\nb=2\nc=3\n`) for the sort-by-key guarantee.
6. In `internal/store/store_test.go`, check `TestMemoryStore_List_ReturnsCopy` — confirms `List()` can't be used to mutate internal state.
7. CI: confirm the `CI` GitHub Actions workflow is green on the PR (build, vet, lint, test).
