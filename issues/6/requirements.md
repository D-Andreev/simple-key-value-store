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

## Acceptance criteria
- [ ] CLI is implemented in Go and builds to a single static binary
- [ ] Store logic sits behind a `Store` interface (`Get`/`Set`/`Delete`/`List`) with an in-memory implementation; no persistence in this issue

## Approved by human
- [ ] Pending — say `approve requirements` in the session when ready
