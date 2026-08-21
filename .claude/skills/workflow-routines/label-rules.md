# GitHub Label Rules

Exactly **one** active workflow label per issue.

## Label swap order

When a phase **advances**, **label swap is usually last** — after handoff commit/push on `workflow/state`, PR creation, and issue comments.

**Exception — clarify start:** swap to `workflow:clarify` **first**.

**Clarify start order:** label swap → **ensure `workflow/state` + init handoff commit** → session comment → Q1 **in session**.

```bash
# implement complete — opens the PR itself, so the comment template uses the
# literal token {pr_url} and wfr substitutes it after creating the PR; one atomic call
# wfr implement complete --issue {n} … --comment-body "… {pr_url} …"
#   → pushes work branch, opens draft PR, commits + pushes state.json, posts the completion comment, swaps label ← LAST — internally

# review complete — PR number/URL already recorded in state.json (set by implement), no live lookup; one atomic call
# wfr review complete --issue {n} --verdict "…" --summary "…"
#   → reads pr_number/pr_url from state.json, finalizes handoff, posts the one PR comment, swaps label ← LAST — internally

# close complete — merged PR already exists, so this is one atomic call too
# wfr close complete --issue {n} --summary "…"
#   → resolves merged PR SHA, finalizes findings-grade.json, posts the issue comment, swaps label ← LAST — internally
```

Record `labels_updated` in `state.json` when committing; swap labels after all other writes.

**Handoff:** branch `workflow/state`, files under `issues/{n}/` (incl. append-only `metrics.jsonl`). **Work branch:** `workflow/issue-{n}` (created at implement). Issue comments: engaging, varied status for humans — [handoff-format.md](handoff-format.md#issue-comments-status-only). Clarify Q&A answers come from the **session**, not the issue thread. Analytics: [metrics.md](metrics.md).

## Label swap

`wfr label swap` is generic across phases — it reads the issue's current labels, rejects a `--from`/`--to` pair that doesn't match reality or isn't one of the five valid transitions, then runs the `gh issue edit` pair:

```bash
wfr label swap --issue 42 --from workflow:start --to workflow:clarify
wfr label swap --issue 42 --from workflow:clarify --to workflow:implement
wfr label swap --issue 42 --from workflow:implement --to workflow:review
wfr label swap --issue 42 --from workflow:review --to workflow:human-review
wfr label swap --issue 42 --from workflow:human-review --to workflow:done
```

All five transitions are backed by `wfr` today. `wfr clarify approve`, `wfr review complete`, and `wfr close complete` each call this internally for their own swap — you only run it standalone at clarify start (`workflow:start` → `workflow:clarify`) and implement complete (`workflow:implement` → `workflow:review`, separate from `wfr implement complete` since that step needs the PR URL first).

## Triggers

| Trigger | Routine |
|---------|---------|
| Label `workflow:start` | Clarify |
| Label `workflow:implement` | Implement |
| Label `workflow:review` | AI Review |
| GitHub: issue **closed** + label `workflow:human-review` | Close |

## Session comments (status only)

Warm, lightly playful, professional — **vary wording every time**; never copy the same template. Reference something specific from the issue when you can. Example bank: [handoff-format.md#issue-comments-status-only](handoff-format.md#issue-comments-status-only).

Clarify Q&A is **session-only** — these comments point humans at the session; they are not where answers are collected.

| Phase | Must include |
|-------|----------------|
| Clarify start | Session link + **state tree** link (`…/tree/workflow/state/issues/{n}`) |
| Clarify approve | Varied header + `---` + full `requirements.md` |
| Implement start | Session link (+ work branch if helpful) |
| Implement complete | Draft PR link |
| Review start | Session + PR links |
| Review complete | Verdict + link to the **one** PR comment |
| Close complete | Addressed / ignored counts (+ critical callouts) |

## Advance commands

| Phase | User says | Effect |
|-------|-----------|--------|
| Clarify | `approve requirements` **in the session** | Commit handoff on state → post requirements on issue → **`workflow:implement` last** |
