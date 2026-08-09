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

**First repo action after label swap:** `wfr clarify init` (ensures `workflow/state`, inits `issues/{n}/`). Each Q&A turn commits via `wfr clarify answer`. Issue comments are **agent→human status only**, posted via `wfr issue comment` (session link at start; approved requirements at approve) — not a reply channel for Q&A.

On **`approve requirements`** (session message), `wfr clarify approve` does the whole sequence in one call: final handoff commit on `workflow/state`, post **approved requirements on the issue**, swap to `workflow:implement`, stop.

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

0. **Ensure the `wfr` CLI is installed** — it owns every handoff write below (state.json, metrics.jsonl, label swaps, requirements finalize) so nothing is hand-typed:
   ```bash
   command -v wfr >/dev/null 2>&1 || curl -fsSL https://raw.githubusercontent.com/D-Andreev/ai-workflow-routines/main/scripts/install.sh | sh
   ```
1. **Swap labels first**:
   ```bash
   wfr label swap --issue {n} --from workflow:start --to workflow:clarify
   ```
   Nothing else on GitHub before this.
2. **Read issue** — number, title, body, labels, URL.
3. **Verify init** — `workflow/PROJECT.md` on `base_branch` via `git show origin/{base_branch}:workflow/PROJECT.md`; else stop → `/workflow-init`.
4. **Ensure `workflow/state` + initial handoff commit**:
   ```bash
   wfr clarify init --issue {n} --issue-url {issue_url} --title "{title}" --base-branch {base_branch} --body "{issue body}"
   ```
   One call: ensures/creates `workflow/state`, writes `state.json` per [fixture](../workflow-routines/fixtures/state-example-clarify-start.json), writes the initial `task.md`/`language.md`/`requirements.md` shells under `issues/{n}/`, creates an empty `metrics.jsonl`, and commits + pushes.
5. **Post session comment** — pick a **fresh phrasing** from handoff-format example bank (or invent one). **Must link** session + **state tree** (`…/tree/workflow/state/issues/{n}`). Reference the issue topic when natural:
   ```bash
   wfr issue comment --issue {n} --body "{session comment text}"
   ```
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
2. Edit `requirements.md` and `language.md` on `workflow/state` — this prose stays model-authored, the CLI never writes it.
3. Log the turn — appends the validated `clarify_turn` line to `metrics.jsonl`, bumps `state.json`, and commits + pushes `workflow/state`, all in one call:
   ```bash
   wfr clarify answer --issue {n} --q-index {i} --category {category} --recommendation-outcome {outcome} --question "{question text}"
   ```
   - `--category` — exactly one from [metrics.md](../workflow-routines/metrics.md#question-categories); `wfr` rejects anything else.
   - `--recommendation-outcome` — exactly one of `skipped` | `accepted_recommendation` | `accepted_with_adjustment` | `rejected_recommendation`; `wfr` rejects anything else.
   - Never rewrite prior JSONL lines — `wfr clarify answer` only appends.
4. Ask next question in the session, or ask for `approve requirements` in the session.

## On `approve requirements` (session message)

One call does the whole sequence — validates `requirements.md` is complete, finalizes `state.json` (`requirements_approved: true`, `status: done`, history), commits + pushes `workflow/state`, posts the approval comment (header + `---` + full `requirements.md`, checkbox block stripped), and swaps `workflow:clarify` → `workflow:implement` **last**. It fails closed: if any step errors, later steps (including the label swap) don't run.

```bash
wfr clarify approve --issue {n} --header "{varied, engaging header — see handoff-format approve example bank, not always 'Requirements approved'}"
```

**Stop** once this succeeds. If it fails, the error says which step failed and whether the label was swapped — fix the cause and re-run rather than swapping labels by hand.

## requirements.md template

`wfr clarify init` generates this from [internal/handoff/templates/requirements.md.tmpl](../../internal/handoff/templates/requirements.md.tmpl) — see that file for the exact shape (`## Original ask`, `## Clarifications` table, `## Acceptance criteria`, `## Approved by human`). Don't hand-author it; edit the generated file's `## Clarifications` rows and `## Acceptance criteria` items as answers come in.

## GitHub writes (clarify)

| When | Git | Issue | Via |
|------|-----|-------|-----|
| Start | Ensure `workflow/state` + init commit under `issues/{n}/` (incl. empty `metrics.jsonl`) | Session comment with **session + state tree links** | `wfr label swap` → `wfr clarify init` → `wfr issue comment` |
| Q&A turn | COMMIT + push on `workflow/state` (handoff + **append** `clarify_turn` to `metrics.jsonl`) | **None** — questions and answers stay in the session | `wfr clarify answer` |
| Approve | COMMIT + push on `workflow/state` | Header + **full requirements.md** → label swap last | `wfr clarify approve` |

## Hard rules

- `state.json`, `metrics.jsonl`, label swaps, and issue comments are written only via the `wfr` CLI (`wfr clarify {init,answer,approve}`, `wfr label swap`, `wfr issue comment`) — never hand-edit `state.json`, hand-append to `metrics.jsonl`, or call `gh issue comment`/`gh issue edit` directly.
- Never write application source code.
- Never create `workflow/issue-{n}` during clarify.
- **Stay on `workflow/state`** — only write under `issues/{n}/`. Read product source / `PROJECT.md` / learnings from **`origin/{base_branch}`**, not the working tree.
- **Ensure state branch at start** — before session comment and Q1.
- **Q&A in session only** — never ask clarifying questions via `gh issue comment`; never wait for issue-thread answers.
- **Issue comments:** varied, engaging — see handoff-format example bank, posted via `wfr issue comment` (or internally by `wfr clarify approve`). Never repeat the same comment verbatim across issues. Only at **start** (session link) and **approve** (requirements), plus failures.
- **Commit handoff after every Q&A turn** and at approve — include a `clarify_turn` metrics line each answered/skipped question.
- **Never put `state.json` or other machine handoff in issue comments** — publish **approved `requirements.md` only** at clarify approve.
- If a `wfr` command fails (e.g. push fails), post a short failure comment via `wfr issue comment` and **stop**; do not swap to `workflow:implement`.
- **Label swap first** at start; **last** at approve.
