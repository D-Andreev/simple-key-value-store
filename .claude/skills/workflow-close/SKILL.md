---
name: workflow-close
description: >-
  Close phase for GitHub-issue workflows. When an issue closes with
  workflow:human-review, grades whether humans addressed AI review findings,
  writes findings-grade.json, appends close_completed to metrics.jsonl, and
  sets workflow:done. Use when the close routine fires on issue closed +
  workflow:human-review, or via /workflow-close.
metadata:
  internal: true
---

# Workflow: Close

Autonomous closeout after human review. **No application code changes.**

Triggered by the **close routine** when a GitHub **issue is closed** and has label **`workflow:human-review`** (GitHub event — not a PR trigger). Typical path: linked issue auto-closes on PR merge. Also runnable manually via `/workflow-close`.

Primary work today: grade AI review findings against the post-review diff. Room for additional closeout steps later.

See [metrics.md](../workflow-routines/metrics.md#close-close_completed), [handoff-format.md](../workflow-routines/handoff-format.md), [label-rules.md](../workflow-routines/label-rules.md).

## Preconditions

1. Issue was **closed** and has (or had at close) label **`workflow:human-review`**. `wfr close complete` enforces this itself via `gh issue view` (closed issues keep labels) — aborts before any writes if it's missing.
2. `workflow/state` has `issues/{n}/review-findings.json` (written at AI review complete) and `state.json` has `pr_number` + `review_head_sha` set.
3. `findings-grade.json` is missing — or the user asked to **re-run close** (`--rerun`; default: skip if already graded).

If preconditions fail, **stop**.

## Sequence (autonomous — complete in one run)

0. **Ensure the `wfr` and `gh` CLIs are installed** — `gh` isn't always preinstalled in the sandbox:
   ```bash
   command -v wfr >/dev/null 2>&1 || curl -fsSL https://raw.githubusercontent.com/D-Andreev/ai-workflow-routines/main/scripts/install.sh | sh
   command -v gh >/dev/null 2>&1 || (apt-get update -qq && apt-get install -y gh)
   ```
1. **Resolve issue `n`** from the GitHub issue-closed trigger (or manual invoke). Checkout **`workflow/state`**; pull latest; read `issues/{n}/review-findings.json` and `state.json`.
2. **Grade each finding** (LLM judgment) against the post-review diff. You don't fetch the diff yourself — read the finding `summary`, then reason about whether it was addressed; `wfr close complete` (below) computes the actual diff and path-touch data mechanically. Write `issues/{n}/close-dispositions.json` with **only** the judgment call per finding — no ids to invent, just match against what's already in `review-findings.json`:
   ```json
   {"dispositions": [
     {"id": "F1", "disposition": "addressed"},
     {"id": "F2", "disposition": "ignored"}
   ]}
   ```
   One entry per finding in `review-findings.json`, `disposition` one of `addressed` | `partial` | `ignored` | `unknown` (see rules below). Empty `review-findings.json`? Write `{"dispositions": []}`.
3. **Complete** — one call does the rest:
   ```bash
   wfr close complete --issue {n} --summary "{one short varied line}"
   ```
   Confirms `workflow:human-review`, resolves the merged PR's head SHA (`gh pr view` on `state.json`'s `pr_number`, erroring if it isn't merged), fetches both SHAs and computes the diff + `commits_since_review` mechanically, cross-references `close-dispositions.json` against `review-findings.json` (rejecting missing or invented finding ids), derives `paths_touched`/`paths_total` per finding from the diff, computes all aggregate counts, writes `findings-grade.json`, appends `close_completed` to `metrics.jsonl`, updates `state.json` (`phase: close`, `workflow_label: workflow:done`, history), commits + pushes, posts the issue comment (your `--summary` + the computed counts + a critical-ignored callout if any), and **swaps labels last** (`workflow:human-review` → `workflow:done`). Fails closed at every step.
4. **Stop.**

Pass `--rerun` to re-grade an issue that already has `findings-grade.json` — for recovering a run that committed `state.json` but failed before the label swap, not for casually re-grading a fully closed issue (the label precondition blocks that once the label really is `workflow:done`).

## Disposition rules (LLM)

| `disposition` | When |
|---------------|------|
| `addressed` | Diff clearly implements the suggestion (or equivalent fix) |
| `partial` | Related change landed but the core issue remains / incomplete |
| `ignored` | No meaningful change related to the finding |
| `unknown` | Cannot tell (no paths, empty diff, or ambiguous) |

`required: true` findings (critical) matter most — `wfr close complete` highlights any `ignored` critical in the issue comment automatically, from the finalized grade.

Path-touch (`paths_touched`/`paths_total`, computed by `wfr close complete`) is a **hint** only for your grading — not something you compute, and not the verdict itself.

## Empty findings

If `review-findings.json`'s `findings` is `[]` (typical `APPROVE`): still write `{"dispositions": []}` to `close-dispositions.json` — `wfr close complete` requires the file to exist. It writes grade + metrics with `suggestions_applicable: false`, all counts zero, and still swaps to `workflow:done`.

## Hard rules

- `findings-grade.json`'s mechanical fields (ids, `paths_touched`/`paths_total`, aggregate counts, `commits_since_review`), `state.json`, `metrics.jsonl`, the issue comment, and the label swap are written only via `wfr close complete` — never hand-edit `state.json`, hand-append to `metrics.jsonl`, or call `gh issue comment`/`gh issue edit` directly.
- **Only write** `close-dispositions.json` yourself under `issues/{n}/` on `workflow/state` — everything else in that directory for this phase is CLI-owned.
- Do not change application code or the work branch.
- **Only run for issues with `workflow:human-review`** — `wfr close complete` enforces this; ignore unrelated closes.
- **Label swap last** — `workflow:human-review` → **`workflow:done`**, handled internally by `wfr close complete`.
- **`llm`** method always (this routine's only mode) — `wfr close complete` sets it.
- **Never invent a disposition for a finding not in `review-findings.json`** — `wfr close complete` rejects extra or missing ids.
- Idempotent: `wfr close complete` skips (errors) if already graded unless `--rerun` is passed.
