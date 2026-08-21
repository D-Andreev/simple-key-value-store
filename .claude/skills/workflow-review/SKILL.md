---
name: workflow-review
description: >-
  AI review phase for GitHub-issue workflows. Fresh-eyes scenario verification
  and principles review on the work branch; handoff on workflow/state; posts
  one PR comment with verdict. Use when a routine fires on workflow:review or
  for issue $0.
metadata:
  internal: true
---

# Workflow: Review

Independent **fresh-eyes** review of the PR branch. Run once, autonomously — **no interactive fix loop**, **no waiting for human approval** in the session.

Triggered by label **`workflow:review`**. Reads handoff from **`workflow/state`** / `issues/{n}/`; reviews code on **`work_branch`**. See [handoff-format.md](../workflow-routines/handoff-format.md), [state-schema.md](../workflow-routines/state-schema.md), [label-rules.md](../workflow-routines/label-rules.md), [metrics.md](../workflow-routines/metrics.md).

## Preconditions

1. Issue has label **`workflow:review`**.
2. `workflow/state` has `issues/{n}/` with `requirements_approved: true`, `implement-handoff.md`, and `work_branch` in `state.json`. `wfr review start` enforces `work_branch` being set itself — it aborts if implement hasn't run.
3. If `review-report.md` exists and `workflow_label` is `workflow:human-review` — stop; human review phase.

If preconditions fail, post a short issue comment via `wfr issue comment` and stop.

## Fresh-eyes rule

**Ignore prior implementation chat** (if any leaked into context). Base judgments only on:

- `issues/{n}/requirements.md`, `implement-handoff.md`, `language.md` on **`workflow/state`**
- `workflow/PROJECT.md`, `workflow/learnings/gotchas.md`
- `git diff {base_branch}...HEAD` on **`work_branch`**
- Code and tests on that branch

State in the review-report header: **"Fresh-eyes: artifacts and diff only."**

## Sequence (autonomous — complete in one run)

0. **Ensure the `wfr` CLI is installed**:
   ```bash
   command -v wfr >/dev/null 2>&1 || curl -fsSL https://raw.githubusercontent.com/D-Andreev/ai-workflow-routines/main/scripts/install.sh | sh
   ```
   `wfr` installs `gh` itself on first use if it's missing from the sandbox — no separate `gh` install step needed.
1. **Read issue**; checkout **`workflow/state`**; pull latest; read `issues/{n}/` handoff files.
2. Checkout **`work_branch`** (`workflow/issue-{n}`); pull latest for diff/code.
3. **Post session comment** — vary phrasing (handoff-format review start bank). Link session + PR:
   ```bash
   wfr issue comment --issue {n} --body "{text}"
   ```
4. **Start** — switch back to `workflow/state` if needed, then:
   ```bash
   wfr review start --issue {n}
   ```
   Verifies implement ran (`work_branch` set), updates `state.json` (`phase: review`, `status: ai_running`, history `started`), commits + pushes.
5. **Review pass** — scenario verification + principles review (below) against work-branch diff.
6. **Verdict** — `APPROVE` | `APPROVE WITH NOTES` | `REQUEST CHANGES`.
7. Write `issues/{n}/review-report.md` and `issues/{n}/review-findings.json` on `workflow/state` (templates below). For findings, author only `severity`/`summary`/`paths` per finding — ids, `required`, and the mechanical envelope fields are derived, not authored.
8. **Complete** — one call does the rest:
   ```bash
   wfr review complete --issue {n} --verdict "{verdict}" --summary "{one short paragraph}" --notes "{optional bullet list}"
   ```
   Reads `pr_number`/`pr_url` already recorded in `state.json` (set when implement opened the PR — nothing is looked up live on GitHub), validates `review-report.md` (structural + that its `## Verdict` section matches `--verdict` exactly), captures `review_head_sha` from the work branch's remote tip, finalizes `review-findings.json` (assigns `F1`…`Fn` ids, derives `required` from severity, injects `review_head_sha`/`pr_number`/`schema_version`/`created_at`), computes `critical_count`/`minor_count`/`notes_count` from the findings and appends `review_completed` to `metrics.jsonl`, updates `state.json` (`review_verdict`, `review_head_sha`, `pr_number`/`pr_url`, `status: done`, history), commits + pushes, posts the **one full-detail** comment on the **PR** (`gh pr comment` — never `gh pr review`), and **swaps labels last** (`workflow:human-review`). Fails closed — if validation or the commit fails, no comment and no label swap; if the comment fails, no label swap. Prints the PR URL.
9. **Post short completion comment on the issue** — a one-liner (verdict + markdown link to the PR from step 8), not the full report:
   ```bash
   wfr issue comment --issue {n} --body "{short text with PR link}"
   ```
10. **Stop.**

Closeout (did humans address findings?) is **not** part of this phase — the **close routine** runs when the issue is **closed** with `workflow:human-review`.

## Review pass

**No tests or build in review.** Implement already ran tests and recorded results in `implement-handoff.md`. Review is **fresh-eyes only**: artifacts, diff, and code reading — never `npm test`, `npm run build`, or equivalent.

### 1. Scenario verification

1. Read artifacts, diff, and implement-handoff test results — built vs required.
2. Derive **scenarios** from requirements and implement-handoff "Suggested review scenarios":
   - Happy path
   - Edge cases from requirements
   - Error / failure paths
   - Regression risks (PROJECT.md domain)
3. **Verify by reading** — trace through code and diff (manual/logical walkthrough). Do **not** execute tests or build.
4. Cross-check implement-handoff test claims; note gaps if untested areas matter.
5. Record results for the report (method: `manual` / `code trace`, not `test`).

### 2. Principles review (same pass)

After scenarios, review:

1. Open 🔴/🟡 from scenario analysis (severity + fix approach)
2. Areas scenarios cannot judge (design, security boundaries, maintainability)
3. Checklist: **Security**, **Design / maintainability**, **Conventions**

Apply stack-idiomatic practices from PROJECT.md and manifests.

**Do not** re-run passed scenarios. Reference them under "Scenario overlap avoided" in the report.

### 3. Verdict

| Verdict | When |
|---------|------|
| **APPROVE** | Requirements met per diff/artifacts; implement-handoff test results accepted; no meaningful issues |
| **APPROVE WITH NOTES** | Shippable; minor suggestions or non-blocking gaps |
| **REQUEST CHANGES** | Must-fix bugs, logic gaps, missing AC, security/design blockers visible in diff/code |

## PR comment (one comment only)

`wfr review complete` builds and posts this — you supply the two pieces that need judgment, it owns the structure and the `gh pr comment` call:

```markdown
**AI review: {--verdict}**

{--summary: one short paragraph, what was checked and overall outcome}

{--notes: bullet list, max 3 — must-fix items for REQUEST CHANGES, non-blocking notes for APPROVE WITH NOTES; omit for clean APPROVE}

Full report: `issues/{n}/review-report.md` on `workflow/state`.
```

**Vary the wording** in `--summary`/`--notes` — use the review example bank or write fresh copy. `wfr review complete` records `pr_comment_posted` in handoff history automatically. **Do not** post a second comment explaining review API failures.

## Issue completion comment (short, separate from the PR comment)

After `wfr review complete` returns (it prints the PR URL), post a **short** completion comment on the **issue** — verdict plus a markdown link to the PR, not the full report:

```markdown
AI review done — **{--verdict}**. See [the PR]({pr_url}) for the full report.
```

Vary the wording; don't reuse boilerplate across issues. The detailed report lives on the PR (previous section) and on `workflow/state`'s `review-report.md` — never duplicate it into the issue comment.

## review-report.md template

`wfr review complete` validates the four top-level headings (`## Verdict`, `## Scenario verification`, `## Principles review`, `## Recommendation`) are present, and that `## Verdict`'s content matches `--verdict` exactly — write freely within each section.

```markdown
# Review Report

**Fresh-eyes:** judgments based on artifacts and diff only (`{base_branch}...HEAD` on `{work_branch}`).

## Verdict
APPROVE | APPROVE WITH NOTES | REQUEST CHANGES

## Scenario verification

### Scenarios verified

| # | Scenario | Method | Result | Notes |
|---|----------|--------|--------|-------|
| 1 | ... | manual/code trace | pass/fail | ... |

### Requirements coverage
- [ ] Each acceptance criterion verified or gap-noted

### Issues found (from review)
- 🔴 Critical: ...
- 🟡 Minor: ...

### Implement test results (cited, not re-run)
- {from implement-handoff.md — e.g. unit/e2e counts, build status}

### Gaps in test coverage
- ...

### Tests/build in review
- **Not run** — review is diff + code reading only; implement phase owns execution.

## Principles review

### Summary
{2-3 sentences}

### Critical (must fix)
- {file:line}

### Suggestions (should consider)
- ...

### Nice to have
- ...

### Scenario overlap avoided
- ...

### Principles applied
- ...

## Recommendation
{One line: merge readiness and remaining risk}
```

## review-findings.json (required at complete)

Write alongside `review-report.md`, **before** running `wfr review complete`. Every 🔴/🟡/note in the report becomes one finding. Use concrete `paths` whenever possible — the merge grader uses them.

You author **only** the judgment call per finding:

```json
{"findings": [
  {"severity": "minor", "summary": "Retry helper does not cap max delay", "paths": ["src/notifications/retry.ts"]}
]}
```

`wfr review complete` reads this and overwrites it with the full normalized document — assigning `id`s (`F1`…`Fn` in array order), deriving `required` from `severity` (`critical` → `true`, `minor`/`note` → `false`), and injecting `schema_version`, `issue_number`, `pr_number`, `review_head_sha` (captured from the work branch's remote tip — you don't compute this yourself), `verdict` (from `--verdict`), and `created_at`. Rejects unknown `severity` values or a missing `summary`.

Full shape: [fixtures/review-findings-example.json](../workflow-routines/fixtures/review-findings-example.json). Schema: [metrics.md](../workflow-routines/metrics.md#review-findings-checklist).

Empty `findings: []` when verdict is clean `APPROVE` — still write the file (`{"findings": []}`), `wfr review complete` requires it to exist.

## Writable locations

| Location | When |
|----------|------|
| `workflow/state` → `issues/{n}/` | `state.json` start + complete (via `wfr review {start,complete}`); `review-report.md` model-authored, `review-findings.json` model-drafted + CLI-finalized, both at complete; **append** `review_completed` to `metrics.jsonl` at complete (via `wfr review complete`) |
| GitHub PR | **One** comment via `wfr review complete` (`gh pr comment` internally) — never `gh pr review` |
| GitHub issue | Short session/complete comments only, via `wfr issue comment` |

**Do not** commit fixes to `work_branch` during review — report only. Author addresses REQUEST CHANGES in a follow-up.

## Hard rules

- `state.json`, `review-findings.json`'s mechanical fields, `metrics.jsonl`, the PR comment, and the label swap are written only via `wfr review {start,complete}` — never hand-edit `state.json`, hand-append to `metrics.jsonl`, or call `gh pr comment`/`gh issue edit` directly.
- Fresh-eyes only — do not cite implementation-chat reasoning in the report.
- **Autonomous** — complete review and **one** PR comment in a single run.
- **One PR comment only**, via `wfr review complete`. **Never** `gh pr review`. **Vary wording** in `--summary`/`--notes` — do not reuse the same review comment across PRs.
- **Do not fix code** during review — verdict and findings only.
- Do not expand scope beyond requirements + review findings.
- Reference specific files and lines in findings.
- **Always write `review-findings.json`** before calling `wfr review complete` (even `{"findings": []}` — it errors if the file doesn't exist).
- **PR comment must be short** — details on `workflow/state` in `review-report.md`.
- **Full review detail (verdict, summary, notes) goes on the PR only.** The issue gets a short completion comment (verdict + PR link) posted separately via `wfr issue comment` after `wfr review complete` returns — never the other way around.
- **Commit handoff at start and complete** on `workflow/state` — `wfr review complete` appends `review_completed` to `metrics.jsonl` itself (never rewrites prior lines).
- If a `wfr` command fails (e.g. push fails), post a short issue comment via `wfr issue comment` and **stop**; do not advance labels.
- **Never put artifacts in issue comments.**
- **Label swap is always last** — `wfr review complete` does this internally, failing closed if any earlier step errored. See label-rules.md.
- **Never run tests or build** during review — cite implement-handoff results; verify by reading diff and code.
