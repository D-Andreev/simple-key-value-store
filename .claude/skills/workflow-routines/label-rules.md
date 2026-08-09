# GitHub Label Rules

Exactly **one** active workflow label per issue.

## Label swap order

When a phase **advances**, **label swap is usually last** — after handoff commit/push on `workflow/state`, PR creation, and issue comments.

**Exception — clarify start:** swap to `workflow:clarify` **first**.

**Clarify start order:** label swap → **ensure `workflow/state` + init handoff commit** → session comment → Q1 **in session**.

```bash
# implement complete
# 1. git commit + push on workflow/state — issues/{n}/ …
# 2. gh pr create --draft … (head = workflow/issue-{n})
# 3. gh issue comment … (varied human comment)
# 4. gh issue edit … ← LAST
```

Record `labels_updated` in `state.json` when committing; swap labels after all other writes.

**Handoff:** branch `workflow/state`, files under `issues/{n}/` (incl. append-only `metrics.jsonl`). **Work branch:** `workflow/issue-{n}` (created at implement). Issue comments: engaging, varied status for humans — [handoff-format.md](handoff-format.md#issue-comments-status-only). Clarify Q&A answers come from the **session**, not the issue thread. Analytics: [metrics.md](metrics.md).

## Label swap

```bash
gh issue edit 42 --remove-label "workflow:start" --add-label "workflow:clarify"
gh issue edit 42 --remove-label "workflow:clarify" --add-label "workflow:implement"
gh issue edit 42 --remove-label "workflow:implement" --add-label "workflow:review"
gh issue edit 42 --remove-label "workflow:review" --add-label "workflow:human-review"
gh issue edit 42 --remove-label "workflow:human-review" --add-label "workflow:done"
```

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
