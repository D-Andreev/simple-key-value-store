---
name: workflow-implement
description: >-
  Implementation phase for GitHub-issue workflows. Reads handoff from
  workflow/state, creates the work branch, merges language into PROJECT.md,
  implements with TDD red-green cycles, commits handoff updates on
  workflow/state, and sets workflow:review. Use when a routine fires on
  workflow:implement or for issue $0.
metadata:
  internal: true
---

# Workflow: Implement

Build per approved `requirements.md`. **Bugfix mode** (`state.mode: bugfix`) follows bugfix process below.

Triggered by **`workflow:implement`**. Reads handoff from **`workflow/state`** / `issues/{n}/`. Creates **`workflow/issue-{n}`** for code. See [handoff-format.md](../workflow-routines/handoff-format.md), [state-schema.md](../workflow-routines/state-schema.md), [label-rules.md](../workflow-routines/label-rules.md).

## Preconditions

1. Label **`workflow:implement`** (only one workflow label).
2. **`workflow/state`** has `issues/{n}/state.json` with `requirements_approved: true`; `requirements.md` approval checked. `wfr implement start` enforces this itself — it aborts before making any changes if `requirements_approved` isn't `true`.
3. If `implement-handoff.md` exists and phase past implement — stop; review should run.

If preconditions fail, post short issue comment via `wfr issue comment`. **Do not invent a second work branch** if one already exists for this issue.

## Start sequence

0. **Ensure the `wfr` CLI is installed**:
   ```bash
   command -v wfr >/dev/null 2>&1 || curl -fsSL https://raw.githubusercontent.com/D-Andreev/ai-workflow-routines/main/scripts/install.sh | sh
   ```
   `wfr` installs `gh` itself on first use if it's missing from the sandbox — no separate `gh` install step needed.
1. Read issue.
2. Verify `workflow/PROJECT.md` on `base_branch` (read via fetch/checkout of base or work branch after create).
3. Post session comment — **vary phrasing** (handoff-format implement example bank). Must include an actual markdown link to the session (`[text]({session_url})`, not just prose mentioning it); work-branch link after create:
   ```bash
   wfr issue comment --issue {n} --checkpoint implement-session-start --body "{session comment text}"
   ```
   `--checkpoint` records a marker in `state.json` history — a duplicate/retried call for the same issue becomes a no-op instead of posting a second comment.
4. **Start** — one call: verifies `requirements_approved: true` (aborts before touching anything otherwise), commits handoff start on `workflow/state` (`phase: implement`, `status: ai_running`, `work_branch`, history `started`), ensures `workflow/issue-{n}` exists (idempotent — creates from `base_branch` if missing, reuses it and records `work_branch_created` in history if newly created, otherwise leaves history untouched), and leaves you **checked out on the work branch**:
   ```bash
   wfr implement start --issue {n} --base-branch {base_branch}
   ```
5. Merge language → PROJECT.md on **work branch**; commit ADRs if present.
6. Feature or bugfix process on **work branch**.
7. Run the project's automated checks (tests, lint, build — whatever the repo normally runs) on **work branch**. Fix failures and re-run until clean. **Do not move on to Complete with known-failing checks.**
8. Complete sequence.

## Prepare repo

`wfr implement start` (above) creates/reuses `workflow/issue-{n}` from `origin/{base_branch}` and checks it out — you don't run this by hand.

1. On the work branch: merge `language.md` → `workflow/PROJECT.md` `## Language`. This stays model-authored (prose merge, not a mechanical write).
2. Commit `adrs.md` → `docs/adr/` if present.
3. Commit workflow merge if changed; push **work branch**.

## Feature / bugfix process

TDD red-green per PROJECT.md; push **work branch**. Never commit `issues/` handoff files onto the work branch.

## Running checks before completing

Before switching to `workflow/state` to write the handoff, actually run the repo's automated checks on the work branch — whatever it normally uses (test runner, linter, build/typecheck) — and fix real failures until they're clean. This is on you, the model: `wfr` does not run or enforce checks itself. `## Test results` in `implement-handoff.md` must report the real output of that run (counts, failures fixed, or why something was skipped) — never an assumed or invented pass.

## implement-handoff.md

Written by the model to `issues/{n}/implement-handoff.md` on **`workflow/state`** right before calling `wfr implement complete` (switch there per [handoff-format.md](../workflow-routines/handoff-format.md#switching-between-state-and-work-branches) if you're still on the work branch). Seven required sections, each with real content (not left as `...`): `## Summary`, `## Branch`, `## Changes`, `## TDD cycles`, `## Test results`, `## Acceptance criteria`, `## Suggested review scenarios`. `wfr implement complete` validates all seven are present and filled in before it will proceed.

## Complete sequence

One call does the whole sequence — pushes final work-branch commits, opens the draft PR (`gh pr create --draft --head workflow/issue-{n} --base {base_branch}`, capturing `pr_number`/`pr_url`), validates `implement-handoff.md`, updates `state.json` (`status: done`, `workflow_label: workflow:review`, `pr_number`, `pr_url`, history), commits + pushes on `workflow/state`, posts the **completion comment** (`--comment-body`, vary phrasing per the handoff-format implement complete bank), and **swaps labels last** (`workflow:review`). It fails closed: if any step errors, later steps (including the label swap) don't run — a failed comment leaves the label unswapped, and a bare retry after full success is rejected (`implement already completed for this issue`) instead of opening a second PR or reposting the comment.

`--comment-body` is a template, not the final text: write it with the literal token `{pr_url}` wherever the draft PR link goes (you don't know the real URL yet — the PR doesn't exist until this call creates it). `wfr` substitutes the real link before posting.

```bash
wfr implement complete --issue {n} --base-branch {base_branch} --pr-title "{title}" --pr-body "{body}" --comment-body "{varied text with the literal token {pr_url}}"
```

**Stop** once this succeeds. If it fails, the error says which step failed and whether the comment/label were applied — fix the cause and re-run rather than swapping labels or posting comments by hand.

## Writable locations

| Location | Files |
|----------|-------|
| `workflow/state` → `issues/{n}/` | `state.json`, `implement-handoff.md` |
| `workflow/issue-{n}` | app code, tests, `workflow/PROJECT.md`, `docs/adr/` |
| Issue | short comments only |

## Hard rules

- `state.json` and the work-branch checkout/creation are written only via `wfr implement {start,complete}` — never hand-edit `state.json`.
- **Create the work branch at start** if missing; never a second work branch for the same issue — `wfr implement start` enforces this.
- **Handoff only on `workflow/state`** — keep PR diffs free of `state.json` / handoff markdown.
- **Commit handoff at start and complete** on `workflow/state`.
- **Never put artifacts in issue comments** — session comment via `wfr issue comment`, completion comment via `wfr implement complete --comment-body`, both short status only.
- If a `wfr` command fails (e.g. push fails), post a short failure comment via `wfr issue comment` and **stop**; do not advance labels.
- **Label swap always last** — `wfr implement complete` does this internally, failing closed if any earlier step errored. Never `gh issue edit` or `wfr label swap` by hand for this transition. PR always **draft**.
- **Never call `wfr implement complete` with known-failing tests, lint, or build.** Run the checks yourself first and fix them — `wfr` does not run or gate on checks.
