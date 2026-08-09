---
name: workflow-implement
description: >-
  Implementation phase for GitHub-issue workflows. Reads handoff from
  workflow/state, creates the work branch, merges language into PROJECT.md,
  implements with TDD red-green cycles, commits handoff updates on
  workflow/state, and sets workflow:review. Use when a routine fires on
  workflow:implement or for issue $0.
disable-model-invocation: true
metadata:
  internal: true
---

# Workflow: Implement

Build per approved `requirements.md`. **Bugfix mode** (`state.mode: bugfix`) follows bugfix process below.

Triggered by **`workflow:implement`**. Reads handoff from **`workflow/state`** / `issues/{n}/`. Creates **`workflow/issue-{n}`** for code. See [handoff-format.md](../workflow-routines/handoff-format.md), [state-schema.md](../workflow-routines/state-schema.md), [label-rules.md](../workflow-routines/label-rules.md).

## Preconditions

1. Label **`workflow:implement`** (only one workflow label).
2. **`workflow/state`** has `issues/{n}/state.json` with `requirements_approved: true`; `requirements.md` approval checked.
3. If `implement-handoff.md` exists and phase past implement — stop; review should run.

If preconditions fail, post short issue comment. **Do not invent a second work branch** if one already exists for this issue.

## Start sequence

1. Read issue.
2. **Checkout `workflow/state`**; pull latest; read `issues/{n}/` per handoff-format.
3. Verify `workflow/PROJECT.md` on `base_branch` (read via fetch/checkout of base or work branch after create).
4. Post session comment — **vary phrasing** (handoff-format implement example bank). Must link session; work-branch link after create.
5. **Commit handoff (start)** on `workflow/state` — update `state.json`: `phase: implement`, `status: ai_running`, `work_branch: workflow/issue-{n}`, history `started`; push.
6. **Create work branch** if missing: `workflow/issue-{n}` from `base_branch`; push; append history `work_branch_created` on next handoff commit (or same start commit if created before push).
7. Merge language → PROJECT.md on **work branch**; commit ADRs if present.
8. Feature or bugfix process on **work branch**.
9. Complete sequence.

## Prepare repo

1. Create `workflow/issue-{n}` from `origin/{base_branch}` **once**. If it already exists, checkout and pull — **do not create another**.
2. On the work branch: merge `language.md` → `workflow/PROJECT.md` `## Language`.
3. Commit `adrs.md` → `docs/adr/` if present.
4. Commit workflow merge if changed; push **work branch**.

## Feature / bugfix process

TDD red-green per PROJECT.md; push **work branch**. Never commit `issues/` handoff files onto the work branch.

## implement-handoff.md

Write to `issues/{n}/implement-handoff.md` on **`workflow/state`** at complete. (Use feature template with Summary, Branch, Changes, TDD cycles, Test results, Acceptance criteria, Suggested review scenarios.)

## Complete sequence

1. Push final work-branch commits.
2. `gh pr create --draft --head workflow/issue-{n} --base {base_branch} …` — capture `pr_number` / `pr_url`.
3. Checkout `workflow/state`; write `implement-handoff.md`; update `state.json` (`status: done`, `workflow_label: workflow:review`, `pr_number`, `pr_url`, history).
4. Commit + push handoff on `workflow/state`.
5. Short issue comment — **vary phrasing** (handoff-format implement complete bank). Must link draft PR.
6. **Swap labels last** — `workflow:review`. **Stop.**

## Writable locations

| Location | Files |
|----------|-------|
| `workflow/state` → `issues/{n}/` | `state.json`, `implement-handoff.md` |
| `workflow/issue-{n}` | app code, tests, `workflow/PROJECT.md`, `docs/adr/` |
| Issue | short comments only |

## Hard rules

- **Create the work branch at start** if missing; never a second work branch for the same issue.
- **Handoff only on `workflow/state`** — keep PR diffs free of `state.json` / handoff markdown.
- **Commit handoff at start and complete** on `workflow/state`.
- **Never put artifacts in issue comments.**
- Push failure → short comment and **stop**; do not advance labels.
- **Label swap always last.** PR always **draft**.
