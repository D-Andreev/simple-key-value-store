---
name: workflow-close
description: >-
  Close phase for GitHub-issue workflows. When an issue closes with
  workflow:human-review, grades whether humans addressed AI review findings,
  writes findings-grade.json, appends close_completed to metrics.jsonl, and
  sets workflow:done. Use when the close routine fires on issue closed +
  workflow:human-review, or via /workflow-close.
disable-model-invocation: true
metadata:
  internal: true
---

# Workflow: Close

Autonomous closeout after human review. **No application code changes.**

Triggered by the **close routine** when a GitHub **issue is closed** and has label **`workflow:human-review`** (GitHub event — not a PR trigger). Typical path: linked issue auto-closes on PR merge. Also runnable manually via `/workflow-close`.

Primary work today: grade AI review findings against the post-review diff. Room for additional closeout steps later.

See [metrics.md](../workflow-routines/metrics.md#close-close_completed), [handoff-format.md](../workflow-routines/handoff-format.md), [label-rules.md](../workflow-routines/label-rules.md).

## Preconditions

1. Issue was **closed** and has (or had at close) label **`workflow:human-review`**.
2. `workflow/state` has `issues/{n}/review-findings.json` (written at AI review complete).
3. `findings-grade.json` is missing — or the user asked to **re-run close** (default: skip if already graded).

If the issue lacks `workflow:human-review` or findings are missing, **stop**.

## Inputs

| Input | Source |
|-------|--------|
| Issue `n` | Closed-issue event / user / issue URL |
| `review_head_sha` | `review-findings.json` (required) |
| `pr_head_sha` | Merged PR head for this issue — from `state.json` `pr_number` / `work_branch` via `gh pr view` |
| Method | **`llm`** (default for findings grading) |

## Sequence (autonomous — complete in one run)

1. **Resolve issue `n`** — from the GitHub issue-closed trigger (or manual invoke). Confirm label `workflow:human-review` with `gh issue view {n} --json labels` (closed issues keep labels).
2. **Checkout `workflow/state`**; pull latest; read `issues/{n}/review-findings.json` and `state.json`.
3. If `findings-grade.json` exists and this is not an explicit re-run — **stop**.
4. **Resolve PR head SHA** — prefer `state.json` `pr_number`; else `gh pr list --head {work_branch} --state merged`. Use the merged PR’s head OID as `pr_head_sha`. If no merged PR, stop with a short issue comment.
5. Fetch SHAs; get the post-review diff:
   ```bash
   git fetch origin {review_head_sha} {pr_head_sha}
   git diff --name-only {review_head_sha}...{pr_head_sha}
   git diff {review_head_sha}...{pr_head_sha}
   ```
6. **Grade each finding** (`method: llm`) → `addressed` | `partial` | `ignored` | `unknown`:
   - Read the finding `summary` + diff for its `paths` (or the full post-review diff if `paths` is empty).
   - Judge whether the human implemented the suggestion. Prefer diff evidence over “file was touched.”
7. Write `issues/{n}/findings-grade.json` (full grade object per [metrics.md](../workflow-routines/metrics.md#close-close_completed)).
8. **Append** one `close_completed` line to `issues/{n}/metrics.jsonl` (`phase: close`, `method: llm`). Do not rewrite prior lines.
9. Update `state.json` — `phase: close`, `workflow_label: workflow:done`, `status: done`, `updated_at`; history `findings_graded` + `phase_completed` / `labels_updated`; set `pr_number` if missing.
10. Commit + push `workflow/state`:
    ```bash
    git add issues/{n}/findings-grade.json issues/{n}/metrics.jsonl issues/{n}/state.json
    git commit -m "workflow(issue-{n}): close"
    git push origin workflow/state
    ```
11. Short **issue** comment — varied summary (addressed / partial / ignored counts; call out ignored **critical**). Link state tree if natural.
12. **Swap labels last** — remove `workflow:human-review`, add **`workflow:done`**. **Stop.**

## Disposition rules (LLM)

| `disposition` | When |
|---------------|------|
| `addressed` | Diff clearly implements the suggestion (or equivalent fix) |
| `partial` | Related change landed but the core issue remains / incomplete |
| `ignored` | No meaningful change related to the finding |
| `unknown` | Cannot tell (no paths, empty diff, or ambiguous) |

`required: true` findings (critical) matter most — highlight any `ignored` critical in the issue comment.

Path-touch is a **hint** only for LLM grading, not the verdict.

## Empty findings

If `findings` is `[]` (typical `APPROVE`): still write grade + metrics with `suggestions_applicable: false`, all counts zero. Still swap to `workflow:done`.

## Hard rules

- **Only write** under `issues/{n}/` on `workflow/state`.
- **Append-only** `metrics.jsonl`.
- Do not change application code or the work branch.
- **Only run for issues with `workflow:human-review`** — ignore unrelated closes.
- **Label swap last** — `workflow:human-review` → **`workflow:done`**.
- Use **`llm`** method in the metrics event (this routine’s default).
- Idempotent: skip if already graded unless re-run was requested.
