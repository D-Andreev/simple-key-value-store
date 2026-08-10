# Review Report

**Fresh-eyes:** judgments based on artifacts and diff only (`main...HEAD` on `workflow/issue-8`).

## Verdict
REQUEST CHANGES

## Scenario verification

### Scenarios verified

| # | Scenario | Method | Result | Notes |
|---|----------|--------|--------|-------|
| 1 | `set foo bar` — exit 0, no output | code trace | pass | `internal/cli/set.go` calls `s.Set` and returns nil; e2e `TestSet_Success` covers it |
| 2 | `get missing` — exit 1, stderr `key not found` | code trace | pass | `internal/cli/get.go` maps `ErrKeyNotFound` to `cmd.PrintErrln("key not found")` + non-nil return; `SilenceErrors`/`SilenceUsage` on root avoids extra cobra output |
| 3 | `delete missing` — exit 1, stderr `key not found` | code trace | pass | same pattern in `internal/cli/delete.go`, mirrored by e2e `TestDelete_MissingKey` |
| 4 | `list` on fresh store — exit 0, empty stdout | code trace | pass | `newListCmd` iterates zero keys, prints nothing |
| 5 | `list` sorted output (`b`,`a`,`c` seeded → `a=1\nb=2\nc=3\n`) | code trace | pass | `sort.Strings(keys)` before printing; matches `cli_test.go` table case |
| 6 | `MemoryStore.List()` returns a defensive copy | code trace | pass | `List()` builds a fresh map from `s.data`; `store_test.go`'s `TestMemoryStore_List_ReturnsCopy` asserts mutating the result doesn't affect internal state |
| 7 | CI green on the PR (build, vet, lint, test) | **live CI check on PR #9** | **fail** | See Critical below — `build-lint-test` check run concluded `failure` |

### Requirements coverage
- [x] `set`/`get`/`delete`/`list` wired to a real in-memory `Store`, string keys/values
- [x] Store behind a `Store` interface with in-memory impl
- [x] Layout `cmd/kvs/main.go`, `internal/cli/`, `internal/store/`; module path `github.com/D-Andreev/simple-key-value-store`
- [x] Unit tests (table-driven) for `Store` and CLI commands
- [x] E2E tests build the binary and exercise it as a subprocess
- [ ] **`golangci-lint` configured and passing** — configured, but **failing in CI** (config-schema mismatch, see below)
- [ ] **GitHub Actions workflow runs build/vet/lint/test on push/PR to `main`** — workflow exists, but the run on PR #9 stops at the lint step and never reaches `go test ./...`
- [x] `get`/`delete` on missing key: non-zero exit + `key not found`; success exits 0
- [x] `list` sorted, one `key=value` per line
- [x] `go.mod` targets a current stable Go (`go 1.24.7`), no pinned minor beyond that

### Issues found (from review)
- 🔴 Critical: `.golangci.yml` (`version: "2"`, `linters.default: standard`) is golangci-lint **v2** schema, but `.github/workflows/ci.yml`'s `golangci-lint-action@v6` step with `version: latest` installed **golangci-lint v1.64.8**, which rejects that schema outright (`jsonschema: "linters" does not validate ... additional properties 'default' not allowed`). The check run for PR #9 (`build-lint-test`, run 31364387265) **concluded `failure`** at the lint step, before the `Test` step ever ran. This directly contradicts AC #1 ("Project builds, tests and lints in github actions") and the "golangci-lint configured and passing" AC — CI is red on this PR right now, not passing as implement-handoff's local run suggested (the local machine likely has a v2-series `golangci-lint` binary installed, masking the mismatch that surfaces in CI's pinned-older installer).

### Implement test results (cited, not re-run)
- implement-handoff.md reports local: `go build ./...` clean, `go vet ./...` clean, `gofmt -l .` clean, `go test ./... -race` all green (`e2e`, `internal/cli`, `internal/store`), `golangci-lint run ./...` → `0 issues`.
- These are plausible for a local `golangci-lint` v2 binary, but do not reflect the CI environment's pinned installer, which failed the config before running.

### Gaps in test coverage
- No test/CI signal currently proves `go test ./...` succeeds under the actual GitHub Actions environment, since the lint step blocks the pipeline before the test step runs.

### Tests/build in review
- **Not run** — review is diff + code reading (plus reading the PR's existing CI check-run result) only; implement phase owns execution.

## Principles review

### Summary
The Go code itself (`store`, `cli`, `cmd/kvs`, tests) is clean, idiomatic scaffolding: a small mutex-guarded `Store` interface/impl, thin Cobra subcommands, table-driven unit tests, and a real subprocess e2e suite. The one blocking problem is entirely in the CI configuration, not the application code: the checked-in `.golangci.yml` schema doesn't match the golangci-lint version the pinned CI action actually installs, so the workflow is red on this PR.

### Critical (must fix)
- `.golangci.yml:1-4` vs `.github/workflows/ci.yml:24-27` — schema/version mismatch breaks CI (`build-lint-test` failing on PR #9). Fix by either: (a) pinning `golangci-lint-action`'s `version` input to a v2 release (e.g. `v2.x.x`) that understands the current config schema, or (b) rewriting `.golangci.yml` to the v1 schema the pinned `latest` (currently resolving to v1.64.8) binary expects. Re-verify the check run goes green afterward.

### Suggestions (should consider)
- `internal/cli/get.go` / `delete.go` silently swallow any non-`ErrKeyNotFound` error (no stderr output, since `SilenceErrors` is set on root) — not reachable today since `Store` only ever returns `ErrKeyNotFound`, but worth a generic "unexpected error" fallback print once real I/O-backed stores can fail in other ways.

### Nice to have
- None beyond the above; the scaffolding is minimal and matches the "no business logic yet" scope well.

### Scenario overlap avoided
- Did not re-run or duplicate the unit/e2e assertions already covered in `implement-handoff.md`'s TDD-cycle notes and test output — verified by reading the test files and tracing the corresponding command code instead.

### Principles applied
- Security: no new attack surface at this scaffolding stage (no untrusted I/O, no persistence yet).
- Design/maintainability: `Store` interface cleanly decouples CLI from storage, consistent with `workflow/PROJECT.md`'s `## Language` entry; defensive copy on `List()` avoids aliasing bugs.
- Conventions: idiomatic Cobra command wiring, table-driven Go tests, package doc comments present.

## Recommendation
Not ready to merge as-is — CI is failing on PR #9 due to the golangci-lint config/version mismatch; application code is otherwise solid and should merge cleanly once the lint step is fixed and the pipeline is confirmed green end-to-end.
