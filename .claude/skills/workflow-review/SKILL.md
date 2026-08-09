---
name: workflow-review
description: >-
  AI review phase for GitHub-issue workflows. Fresh-eyes scenario verification
  and principles review on the work branch; handoff on workflow/state; posts
  one PR comment with verdict. Use when a routine fires on workflow:review or
  for issue $0.
disable-model-invocation: true
metadata:
  internal: true
---

# Workflow: Review

Independent **fresh-eyes** review of the PR branch. Run once, autonomously — **no interactive fix loop**, **no waiting for human approval** in the session.

Triggered by label **`workflow:review`**. Reads handoff from **`workflow/state`** / `issues/{n}/`; reviews code on **`work_branch`**. See [handoff-format.md](../workflow-routines/handoff-format.md), [state-schema.md](../workflow-routines/state-schema.md), [label-rules.md](../workflow-routines/label-rules.md), [metrics.md](../workflow-routines/metrics.md).

## Preconditions

1. Issue has label **`workflow:review`**.
2. `workflow/state` has `issues/{n}/` with `requirements_approved: true`, `implement-handoff.md`, and `work_branch` in `state.json`.
3. If `review-report.md` exists and `workflow_label` is `workflow:human-review` — stop; human review phase.

If preconditions fail, post a short issue comment and stop.

## Fresh-eyes rule

**Ignore prior implementation chat** (if any leaked into context). Base judgments only on:

- `issues/{n}/requirements.md`, `implement-handoff.md`, `language.md` on **`workflow/state`**
- `workflow/PROJECT.md`, `workflow/learnings/gotchas.md`
- `git diff {base_branch}...HEAD` on **`work_branch`**
- Code and tests on that branch

State in the review-report header: **"Fresh-eyes: artifacts and diff only."**

## Sequence (autonomous — complete in one run)

1. **Read issue**; checkout **`workflow/state`**; pull latest; read `issues/{n}/` handoff files.
2. Checkout **`work_branch`** (`workflow/issue-{n}`); pull latest for diff/code.
3. **Post session comment** — vary phrasing (handoff-format review start bank). Link session + PR.
4. **Commit handoff (start)** on `workflow/state` — update `state.json`: `phase: review`, `status: ai_running`, history `started`; push.
5. **Find PR** — `gh pr list --head workflow/issue-{n} --json number,url`.
6. **Review pass** — scenario verification + principles review (below) against work-branch diff.
7. **Verdict** — `APPROVE` | `APPROVE WITH NOTES` | `REQUEST CHANGES`.
8. **Post one PR comment** with verdict — `gh pr comment` only. See below. **Never** `gh pr review`.
9. **On work branch** — capture `review_head_sha=$(git rev-parse HEAD)` before switching away.
10. **Update handoff on `workflow/state`:**
    - Write `review-report.md`
    - Write **`review-findings.json`** (structured checklist — see [metrics.md](../workflow-routines/metrics.md#review-findings-checklist); empty `findings` when `APPROVE` with nothing to track). Counts must match the report.
    - Update `state.json` — `review_verdict`, `review_head_sha`, `pr_number` / `pr_url` when known, `status: done`, history (`phase_completed`, `pr_comment_posted`)
    - **Append one `review_completed` line** to `metrics.jsonl` (`verdict`, `critical_count`, `minor_count`, `notes_count`)
    - Commit + push `workflow/state`
11. Post **short issue comment** — varied verdict line + PR comment link (handoff-format review complete bank).
12. **Swap labels last** — **`workflow:human-review`**. **Stop.**

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

Post **exactly one** comment on the PR. **Vary the wording** — use the review example bank or write fresh copy; include verdict in the opening line. **`gh pr comment` only.**

Full detail stays in `issues/{n}/review-report.md` on **`workflow/state`**.

**Template (adapt, stay concise — one comment, no follow-ups):**

```markdown
**AI review: {APPROVE | APPROVE WITH NOTES | REQUEST CHANGES}**

{One short paragraph: what was checked and overall outcome.}

{If REQUEST CHANGES: bullet list of must-fix items, max 3.}
{If APPROVE WITH NOTES: bullet list of non-blocking notes, max 3.}

Full report: `issues/{n}/review-report.md` on `workflow/state`.
```

**Post:**

```bash
gh pr comment {pr_number} --body "$(cat <<'EOF'
{comment body — include verdict in the opening line}
EOF
)"
```

Record `pr_comment_posted` in handoff history. **Do not** post a second comment explaining review API failures.

## review-report.md template

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

Write alongside `review-report.md`. Every 🔴/🟡/note in the report becomes one finding. Use concrete `paths` whenever possible — the merge grader uses them.

```bash
# On work_branch, before checkout workflow/state:
review_head_sha=$(git rev-parse HEAD)
```

Template shape: [fixtures/review-findings-example.json](../workflow-routines/fixtures/review-findings-example.json). Schema: [metrics.md](../workflow-routines/metrics.md#review-findings-checklist).

| `severity` | From report | `required` |
|-------------|-------------|------------|
| `critical` | Critical / 🔴 / must fix | `true` |
| `minor` | Minor / 🟡 / suggestions | `false` |
| `note` | Nice to have / non-blocking | `false` |

Ids: `F1`…`Fn` in report order. Empty `findings: []` when verdict is clean `APPROVE`.

## Writable locations

| Location | When |
|----------|------|
| `workflow/state` → `issues/{n}/` | `state.json` start + complete; `review-report.md` + **`review-findings.json`** at complete; **append** `review_completed` to `metrics.jsonl` at complete |
| GitHub PR | **One** comment via `gh pr comment` — never `gh pr review` |
| GitHub issue | Short session/complete comments only |

**Do not** commit fixes to `work_branch` during review — report only. Author addresses REQUEST CHANGES in a follow-up.

## Hard rules

- Fresh-eyes only — do not cite implementation-chat reasoning in the report.
- **Autonomous** — complete review and **one** PR comment in a single run.
- **One PR comment only** — `gh pr comment`. **Never** `gh pr review`. **Vary wording** — do not reuse the same review comment across PRs.
- **Do not fix code** during review — verdict and findings only.
- Do not expand scope beyond requirements + review findings.
- Reference specific files and lines in findings.
- **Always write `review-findings.json`** with `review_head_sha` at complete (even if `findings` is empty).
- **PR comment must be short** — details on `workflow/state` in `review-report.md`.
- **Commit handoff at start and complete** on `workflow/state` — at complete, append `review_completed` to `metrics.jsonl` (do not rewrite prior lines).
- If push fails, post short issue comment and **stop**; do not advance labels.
- **Never put artifacts in issue comments.**
- **Label swap is always last** — see label-rules.md.
- **Never run tests or build** during review — cite implement-handoff results; verify by reading diff and code.
