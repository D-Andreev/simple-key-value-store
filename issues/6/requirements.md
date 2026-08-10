# Requirements: issue-6

## Original ask
## Summary
Scaffold project initial structure. Use best patterns.

## Scope
- [ ] e2e and unit tests
- [ ] Build, lint and test in github actions

## Clarifications
| # | Question | Answer | Recommended |
|---|----------|--------|-------------|
| 1 | What language/runtime should the CLI be built in? | Go | Go |
| 2 | Should this skeleton include on-disk persistence, or in-memory only (persistence deferred)? | In-memory only, behind a `Store` interface | In-memory only, behind a `Store` interface |
| 3 | What commands should the CLI expose, and how invoked (subcommands vs. REPL)? | Subcommands: `set`, `get`, `delete`, `list` | Subcommands: `set`, `get`, `delete`, `list` |
| 4 | CLI parsing approach and project layout? | `cobra` + standard `cmd/`+`internal/` layout | `cobra` + standard `cmd/`+`internal/` layout |
| 5 | Testing/CI stack: stdlib `testing` + subprocess e2e + `golangci-lint` + GH Actions (build/vet/lint/test on push+PR)? | Yes, all of it | Yes, all of it |

## Acceptance criteria
- [ ] CLI is implemented in Go and builds to a single static binary
- [ ] Store logic sits behind a `Store` interface (`Get`/`Set`/`Delete`/`List`) with an in-memory implementation; no persistence in this issue
- [ ] CLI exposes subcommands `set <key> <value>`, `get <key>`, `delete <key>`, `list`; exit code 0 on success, non-zero with a clear message on failure (e.g. missing key)
- [ ] Project uses `cobra` for command routing, laid out as `cmd/kvs/main.go` (entrypoint), `internal/cli/` (commands), `internal/store/` (Store interface + in-memory impl)
- [ ] Unit tests (stdlib `testing`, table-driven) cover the `Store` implementation and CLI commands
- [ ] E2E tests build the binary and exercise it as a subprocess, asserting stdout/exit code per command
- [ ] `golangci-lint` configured and passing
- [ ] GitHub Actions workflow runs `go build`, `go vet`, `golangci-lint run`, `go test ./...` on push/PR to `main`

## Approved by human
- [ ] Pending — say `approve requirements` in the session when ready
