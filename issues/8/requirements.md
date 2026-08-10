# Requirements: issue-8

## Original ask
## Summary
Scaffold project skeleton.  Use golang, standart lib, cobra for CLI

## Acceptance criteria
- [ ] Project builds, tests and lints in github actions
- [ ] Initial structure for CLI with subcommands, not REPL is created
- [ ] No business logic, just structure for now
- [ ] In memory storage for now

## Clarifications
| # | Question | Answer | Recommended |
|---|----------|--------|-------------|
| 1 | Does "no business logic, just structure" mean the `set/get/delete/list` subcommands are pure stubs (print placeholder, no wiring), or should they be wired to a real in-memory `Store` for basic get/set/delete/list, just without validation/error-message polish? | (A) Wired to a real in-memory `Store` — basic get/set/delete/list, no validation or polished error handling yet | (A) Wired to a real in-memory `Store` |
| 2 | Project layout and Go module path? | `cmd/kvs/main.go` (entrypoint) + `internal/cli/` (commands) + `internal/store/` (interface + in-memory impl); module `github.com/D-Andreev/simple-key-value-store`; binary name `kvs` | Same |

## Acceptance criteria
- [ ] CLI exposes subcommands `set <key> <value>`, `get <key>`, `delete <key>`, `list`, each performing the corresponding basic operation against an in-memory `Store` (no input validation or polished error messages yet — that's future work)
- [ ] Store logic sits behind a `Store` interface with an in-memory implementation
- [ ] Project laid out as `cmd/kvs/main.go` (entrypoint), `internal/cli/` (commands), `internal/store/` (Store interface + in-memory impl); module path `github.com/D-Andreev/simple-key-value-store`
- [ ] ...

## Approved by human
- [ ] Pending — say `approve requirements` in the session when ready
