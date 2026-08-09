---
name: workflow-clarify
description: >-
  Clarification phase for GitHub-issue workflows. Writes ephemeral handoff
  files on the long-lived workflow/state branch, grills requirements one
  question at a time in the session (humans answer only in session), posts
  short status issue comments, and sets workflow:implement. Use when a
  routine fires on workflow:start or for issue $0.
disable-model-invocation: true
metadata:
  internal: true
---

# Workflow: Clarify

Grill the plan before implementation. **No application code changes. No test runs. No work branch.**

**Working tree is `workflow/state`.** Stay checked out there for all handoff writes. The state branch has **no application source** — when you need product code, `workflow/PROJECT.md`, or learnings, read them from **`origin/{base_branch}`** (usually `main`) via `git show` / `git ls-tree`. Do **not** read those paths from the working tree, and do **not** checkout `base_branch` mid-clarify.

**Q&A channel is the session only.** Ask clarifying questions in the Claude Code session chat. Humans answer **in the session** — never on the GitHub issue thread. Do **not** wait for, poll, or expect issue comments as answers. `approve requirements` is also said **in the session**.

**First repo action after label swap:** ensure `workflow/state` and init `issues/{n}/`. Commit handoff files on each Q&A turn. Issue comments are **agent→human status only** (session link at start; approved requirements at approve) — not a reply channel for Q&A.

On **`approve requirements`** (session message), final handoff commit on `workflow/state`, post **approved requirements on the issue**, swap to `workflow:implement`, stop.

See [handoff-format.md](../workflow-routines/handoff-format.md), [state-schema.md](../workflow-routines/state-schema.md), [label-rules.md](../workflow-routines/label-rules.md).

## Reading product files (from `base_branch`)

Handoff lives on `workflow/state`; product files live on `base_branch`. Always stay on `workflow/state` and read via git:

```bash
git fetch origin {base_branch}
git show origin/{base_branch}:workflow/PROJECT.md
git show origin/{base_branch}:path/to/source.ext
git ls-tree -r --name-only origin/{base_branch} | head
```

| File | How |
|------|-----|
| `workflow/PROJECT.md` | `git show origin/{base_branch}:workflow/PROJECT.md` |
| `workflow/learnings/gotchas.md` | Same pattern from `base_branch` |
| Application code | Read-only from `origin/{base_branch}:…` — never from the state-branch working tree |

After handoff exists, **only write** under `issues/{n}/` on **`workflow/state`** during clarify.

## Handoff files (on `workflow/state`)

| File | Path |
|------|------|
| `state.json` | `issues/{n}/state.json` |
| `metrics.jsonl` | `issues/{n}/metrics.jsonl` (append-only analytics) |
| `task.md` | `issues/{n}/task.md` |
| `language.md` | `issues/{n}/language.md` |
| `requirements.md` | `issues/{n}/requirements.md` |
| `adrs.md` | `issues/{n}/adrs.md` (optional) |

`workflow/state` is **source of truth** for machine handoff — commit after every update. Do **not** create `workflow/issue-{n}` during clarify.

Analytics events: [metrics.md](../workflow-routines/metrics.md).

## Trigger modes

### A — Routine start (`workflow:start`)

Start sequence → grilling loop.

### B — Session continuation (`workflow:clarify`, same session)

Checkout `workflow/state`, read `issues/{n}/`, continue.

### C — Manual (`/workflow-clarify {issue_number}`)

Same as A/B.

## Start sequence (mode A)

1. **Swap labels first** — `workflow:clarify`. Nothing else on GitHub before this.
2. **Read issue** — number, title, body, labels, URL.
3. **Verify init** — `workflow/PROJECT.md` on `base_branch` via `git show origin/{base_branch}:workflow/PROJECT.md`; else stop → `/workflow-init`.
4. **Ensure `workflow/state` + initial handoff commit** (see handoff-format):
   - Ensure-or-create long-lived `workflow/state`
   - Write `state.json` per [fixture](../workflow-routines/fixtures/state-example-clarify-start.json)
   - Write initial `task.md`, `language.md`, `requirements.md` under `issues/{n}/`
   - Create empty `issues/{n}/metrics.jsonl`
   - Commit + push `workflow/state`
5. **Post session comment** — pick a **fresh phrasing** from handoff-format example bank (or invent one). **Must link** session + **state tree** (`…/tree/workflow/state/issues/{n}`). Reference the issue topic when natural.

   Do **not** reuse the same clarify-start comment across issues.
6. Ask **first question in the session** (chat). Do not post it as an issue comment.

## Grilling loop

One question at a time **in the session**, with a recommended answer. Explore the codebase **from `origin/{base_branch}`** (via `git show` / `git ls-tree`) before asking — stay on `workflow/state`.

After each question: **stop and wait for the human's next session message.** Do not post the question as an issue comment. Do not treat issue-thread replies as answers.

Set `status` to `awaiting_human` while waiting — that means awaiting a **session** reply, not an issue comment.

## Domain modeling (`language.md`)

Update on `workflow/state` when terms resolve — same format as `PROJECT.md` `## Language`. Implement merges to PROJECT.md later.

## Resume

Checkout `workflow/state`, read `issues/{n}/`, or use session comment link. Continue grilling from the last unanswered question using the **session** transcript.

## On human answers (session messages)

1. Read the answer from the **session** (not the issue thread).
2. Update `requirements.md`, `language.md`, `state.json` on `workflow/state`.
3. **Append one `clarify_turn` to `metrics.jsonl`** (same commit):
   - `category` — exactly one from [metrics.md](../workflow-routines/metrics.md#question-categories)
   - `recommendation_outcome` — exactly one of: `skipped` | `accepted_recommendation` | `accepted_with_adjustment` | `rejected_recommendation`
   - `q_index`, `question` text
   - Never rewrite prior JSONL lines
4. Commit + push `workflow/state`.
5. Ask next question in the session, or ask for `approve requirements` in the session.

## On `approve requirements` (session message)

1. Verify `requirements.md` complete.
2. Finalize `state.json` — `requirements_approved: true`, `status: done`, history.
3. Check approval in `requirements.md`.
4. Commit + push `workflow/state`.
5. **Post approval comment** on the issue:
   - **Varied header** (see handoff-format approve example bank — not always "Requirements approved")
   - `---`
   - **Full approved `requirements.md`**
   ```bash
   gh issue comment {n} --body-file /tmp/requirements-comment.md
   ```
6. **Swap labels last** — `workflow:implement`. **Stop.**

## requirements.md template

```markdown
# Requirements: issue-{number}

## Original ask
{from task.md}

## Clarifications
| # | Question | Answer | Recommended |
|---|----------|--------|-------------|

## Acceptance criteria
- [ ] ...

## Approved by human
- [ ] Pending — say `approve requirements` in the session when ready
```

## GitHub writes (clarify)

| When | Git | Issue |
|------|-----|-------|
| Start | Ensure `workflow/state` + init commit under `issues/{n}/` (incl. empty `metrics.jsonl`) | Session comment with **session + state tree links** |
| Q&A turn | COMMIT + push on `workflow/state` (handoff + **append** `clarify_turn` to `metrics.jsonl`) | **None** — questions and answers stay in the session |
| Approve | COMMIT + push on `workflow/state` | Header + **full requirements.md** → label swap last |

## Hard rules

- Never write application source code.
- Never create `workflow/issue-{n}` during clarify.
- **Stay on `workflow/state`** — only write under `issues/{n}/`. Read product source / `PROJECT.md` / learnings from **`origin/{base_branch}`**, not the working tree.
- **Ensure state branch at start** — before session comment and Q1.
- **Q&A in session only** — never ask clarifying questions via `gh issue comment`; never wait for issue-thread answers.
- **Issue comments:** varied, engaging — see handoff-format example bank. Never repeat the same comment verbatim across issues. Only at **start** (session link) and **approve** (requirements), plus failures.
- **Commit handoff after every Q&A turn** and at approve — include a `clarify_turn` metrics line each answered/skipped question.
- **Never put `state.json` or other machine handoff in issue comments** — publish **approved `requirements.md` only** at clarify approve.
- If push fails, short issue comment and **stop**; do not swap to `workflow:implement`.
- **Label swap first** at start; **last** at approve.
