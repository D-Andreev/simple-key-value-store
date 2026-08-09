# Branch Handoff Format (Ephemeral Files) + Human Issue Comments

Ephemeral machine files live on the **long-lived state branch** `workflow/state` under `issues/{n}/`. The **work branch** `workflow/issue-{n}` holds only product code (and shipped workflow docs). Issue comments are **agent→human status only** — engaging, short where possible. **Clarify Q&A (questions and answers) happens in the Claude Code session**, not the issue thread. **Approved `requirements.md` is posted on the issue at clarify approve.** Never put `state.json` or other machine handoff in comments.

Full schema: [state-schema.md](state-schema.md) · Analytics: [metrics.md](metrics.md)

## Split

| Layer | Where | Audience |
|-------|-------|----------|
| **Handoff files** | `workflow/state` → `issues/{n}/` | Agents read/write; audit forever |
| **Session** | Claude Code session chat | Clarify Q&A — humans answer **here only** |
| **Issue comments** | Issue thread | Agent→human status (links, approved requirements) — **not** a Q&A reply channel |
| **Work branch** | `workflow/issue-{n}` — app code, `workflow/PROJECT.md`, `docs/adr/` | Shipped artifacts (implement+) |

**Clarify writes only to `workflow/state`.** Implement **creates** `workflow/issue-{n}` for code and keeps updating handoff on `workflow/state`. Never merge `workflow/state` into `main`.

## Branch layout

**State branch:** `workflow/state` (created at **`/workflow-init`**; ensure-or-create at clarify start)

**Handoff directory:** `issues/{n}/` on `workflow/state`

**Work branch:** `workflow/issue-{n}` (created at **implement start** from `base_branch`)

| File | First writer | Updated by |
|------|--------------|------------|
| `state.json` | clarify (start) | **every phase** (start + complete) |
| `metrics.jsonl` | clarify (start — empty or first turn) | clarify (each Q&A); review (complete); **close (issue close)** — **append only** |
| `task.md` | clarify (start) | clarify (Q&A commits) |
| `language.md` | clarify (start) | clarify (Q&A commits) |
| `requirements.md` | clarify (start) | clarify (Q&A + approve) |
| `adrs.md` | clarify (optional) | clarify |
| `implement-handoff.md` | implement (complete) | — |
| `review-report.md` | review (complete) | — |
| `review-findings.json` | review (complete) | — (checklist for close grader) |
| `findings-grade.json` | close routine / `/workflow-close` | — |

Analytics event shapes and enums: [metrics.md](metrics.md).

During clarify, **only write** under `issues/{n}/` on `workflow/state` — not application code, not `workflow/PROJECT.md`, not a work branch.

## Issue comments (status only)

Clarify questions and human answers live in the **session**, not here. Issue comments are status for humans (session/PR links; approved requirements at clarify approve).

### Voice

Write like a sharp teammate in Claude Code — **warm, a little playful, never corny**. Professional always. Informative first: what happened, what you need from the human, links that work.

**Vary your wording every time.** Do not reuse the same opener, emoji, or sentence shape across issues or phases. Pull a detail from the issue title, a clarification, or the diff so the comment feels written for *this* task.

| Do | Don't |
|----|--------|
| Rotate phrasing — use the **example bank** below, mix lines, or invent fresh copy | Copy-paste the same template every run |
| One light touch of personality (metaphor, gentle humor, specific praise) | Robotic status lines ("Phase complete.") |
| Required markdown links (session, state tree, PR) | Placeholder URLs or backticks without links |
| 1–2 emojis **max**, and skip them sometimes | Emoji every comment or stacked emoji |
| `state.json` / machine handoff on `workflow/state` only | JSON or handoff dumps in comments |

**Exception — clarify approve:** header (varied, engaging) + `---` + **full approved `requirements.md`**. See below.

### Clarify start (required links)

Post **after** state-branch push. **Both links required:** Claude Code session + GitHub tree URL for this issue’s handoff (`…/tree/workflow/state/issues/{n}` via `gh repo view --json nameWithOwner -q .nameWithOwner`).

**Example bank — pick one style, adapt to the issue:**

```markdown
**Clarify** — Time to interrogate the spec (gently). [Hop into the session]({session_url}) — I'll ask one question at a time. Notes accumulating on [`workflow/state` · issues/42]({state_tree_url}).

**Clarify** — Requirements workshop is open. [Your turn in the session]({session_url}); I'm keeping the living spec under [`issues/42` on `workflow/state`]({state_tree_url}).

**Clarify** — Before we write code, let's agree on what "done" means. [Session here]({session_url}) · handoff on [`workflow/state`]({state_tree_url}).
```

Optional: `<!-- ai-workflow:handoff v5 issue=42 state_branch=workflow/state path=issues/42 -->`

### Clarify approve (post requirements)

1. **Header** — one line, varied (celebrate, confirm scope, nod to something specific from clarifications). Not always "Requirements approved".
2. `---`
3. **Full `requirements.md`** from state branch (omit `## Approved by human` checkbox block).

**Header example bank:**

```markdown
**Locked in** — We agreed on retry semantics, backoff, and test coverage. Implement is queued unless you spot a hole below.

**Spec approved** — This is the contract for issue #42. Push back in the session if anything reads wrong.

**Green light on requirements** — Nice clarifications on the edge cases. Next phase picks up from here.
```

Then `---` and full requirements body. If over ~65k chars, post AC + clarifications + state-tree link; say so in the header.

### Implement

**Start — example bank:**

```markdown
**Implement** — Spec's approved; time to make it real on [`workflow/issue-42`]({branch_url}). [Watch or steer in session]({session_url}).

**Implement** — Rolling up sleeves on {short task reference from title}. Branch: [`workflow/issue-42`]({branch_url}) · [session]({session_url}).

**Implement** — Building against the approved AC. I'll mirror existing patterns where I can. [`workflow/issue-42`]({branch_url}) · [session]({session_url}).
```

**Complete — example bank:**

```markdown
**Draft PR is up** — [#17]({pr_url}) has the changes; left as **draft** so you can kick the tires locally before review.

**Ready for your eyes** — [#17]({pr_url}) is open (draft). Run it locally when you can; AI review follows.

**Implement done** — Code + tests on [#17]({pr_url}). Draft on purpose — mark ready when you're happy.
```

### Review

**Start — example bank:**

```markdown
**Review** — Fresh eyes, no implementation baggage. [Session]({session_url}) · diff in [#17]({pr_url}).

**Review** — Reading this like I've never seen the repo. [#17]({pr_url}) · [join in]({session_url}) if curious.

**Review** — Sanity-checking against the spec and the diff. [#17]({pr_url}) · [session]({session_url}).
```

**Complete — example bank** (link to **one** PR comment, not a formal review):

```markdown
**Review: APPROVE WITH NOTES** — Shippable; one small nit in the [PR comment]({pr_comment_url}). Full autopsy on [`workflow/state`]({state_tree_url}).

**Review: APPROVE** — Matches the spec; implement test results look good on paper. [Summary on the PR]({pr_comment_url}).

**Review: REQUEST CHANGES** — Found blockers — details in the [PR comment]({pr_comment_url}). Happy to re-review after fixes.
```

### Failures

```markdown
**Clarify** — Handoff push failed ({error}). Nothing else ran — [session]({session_url}) when you're ready to retry.

**Implement** — Couldn't push handoff ({error}). State branch may be stale; [session]({session_url}).
```

## Write protocol

### Git (cloud routines — required)

Handoff writes **must** use **git commit + push** on **`workflow/state`** via git/`gh` authenticated by the **[Claude GitHub App](https://github.com/apps/claude)**. Product code commits go on **`workflow/issue-{n}`** only.

**Do:**

- Ensure `workflow/state` exists (init or clarify start); commit handoff files on every clarify Q&A turn and at each phase start/complete **on `workflow/state`**
- Create `workflow/issue-{n}` at **implement start**; commit app code there
- `gh issue comment` for human comments (including **approved requirements** at clarify approve)
- Push after each handoff commit on `workflow/state`

**Do not:**

- Store machine handoff in issue comments (`state.json`, raw handoff paths) — **except** publishing approved `requirements.md` at clarify approve
- Put handoff files on the work branch (keeps PR diffs clean)
- Use gists for handoff
- Create a second **work** branch for the same issue
- Merge `workflow/state` into `main` / `base_branch`
- Write application code during clarify

### Ensure state branch

```bash
git fetch origin
if git rev-parse --verify origin/workflow/state >/dev/null 2>&1; then
  git checkout -B workflow/state origin/workflow/state
else
  git checkout --orphan workflow/state
  git rm -rf . 2>/dev/null || true
  mkdir -p issues
  printf '%s\n' '# Workflow state' '' 'Long-lived handoff and audit branch. Do not merge into main.' > README.md
  git add README.md
  git commit -m "workflow: init state branch"
  git push -u origin workflow/state
fi
```

### Clarify start (state branch first)

After label swap to `workflow:clarify`:

1. Read issue; verify `workflow/PROJECT.md` exists on `base_branch` (`git show origin/{base_branch}:workflow/PROJECT.md`).
2. Ensure `workflow/state` (above); **stay checked out on it**; write initial handoff:

```bash
mkdir -p issues/{n}
# write state.json, task.md, language.md, requirements.md (initial shell)
# create empty metrics.jsonl for analytics (append-only later)
: > issues/{n}/metrics.jsonl
git add issues/{n}/
git commit -m "workflow(issue-{n}): clarify — init handoff"
git push origin workflow/state
```

3. Post short session comment with **both markdown links** (session + state tree). See **Clarify start**.
4. Ask first question **in the session** (not as an issue comment). Wait for the human's **session** reply.

**Reading product code during clarify:** working tree is `workflow/state` (no app source). Use `git show origin/{base_branch}:path` / `git ls-tree` — do not checkout `base_branch`.

### Clarify Q&A (each human session answer)

1. Read the answer from the **session** — do not wait for or use issue-thread replies.
2. On `workflow/state`, update files under `issues/{n}/`.
3. **Append one `clarify_turn` line** to `issues/{n}/metrics.jsonl` — category + recommendation outcome per [metrics.md](metrics.md). Never rewrite prior lines.
4. Commit and push:

```bash
git add issues/{n}/
git commit -m "workflow(issue-{n}): clarify — update requirements"
git push origin workflow/state
```

5. Ask the next question in the session (or request `approve requirements` in the session).

### Clarify approve

1. Human says `approve requirements` **in the session**.
2. Finalize `state.json` (`requirements_approved: true`, `status: done`, history).
3. Commit and push handoff on `workflow/state`.
4. Post approval comment — engaging header + `---` + **full approved `requirements.md`**. See **Clarify approve (post requirements)** above.
5. **Swap labels last** — `workflow:implement`.

### Implement

1. Checkout / pull `workflow/state`; read `issues/{n}/state.json` and siblings. Require `requirements_approved: true`.
2. **At phase start** — update `state.json` on `workflow/state` (`phase: implement`, `work_branch`, history); commit + push.
3. **Create work branch** (first time only):

```bash
git fetch origin
git checkout -b workflow/issue-{n} origin/{base_branch}
git push -u origin workflow/issue-{n}
```

4. Do phase work on `workflow/issue-{n}` (code + PROJECT.md + ADRs). **Do not** add `issues/` handoff files to the work branch.
5. **At phase complete** — checkout `workflow/state`, write `implement-handoff.md` + update `state.json`; commit + push; open draft PR from work branch.

**Never create a second work branch** for the same issue. If `workflow/issue-{n}` already exists, reuse it.

### Switching between state and work branches

When the session is on the work branch and you need to update handoff:

```bash
git fetch origin
git checkout workflow/state
git pull origin workflow/state
# edit issues/{n}/…
git add issues/{n}/
git commit -m "workflow(issue-{n}): …"
git push origin workflow/state
git checkout workflow/issue-{n}   # resume code work
```

Prefer a clean working tree before switching. If needed, commit or stash work-branch changes first.

### Review

1. Pull `workflow/state`; read handoff. Checkout `work_branch` for diff/code.
2. **At phase start** — update `state.json` on `workflow/state`; commit + push.
3. Review (diff + code reading only — no tests/build). Do not commit code fixes.
4. Capture `review_head_sha` on `work_branch` (`git rev-parse HEAD`).
5. **At phase complete** — write `review-report.md` + **`review-findings.json`** + update `state.json` (`review_head_sha`, `review_verdict`, `pr_number`) on `workflow/state`; **append one `review_completed` line** to `metrics.jsonl`; commit + push.
6. Post short issue comment; **label swap last**.

### Close (issue close — close routine)

When the issue is **closed** while labeled `workflow:human-review`, the **close routine** (GitHub issues event — not PR) resolves the merged PR from `state.json`, diffs `{review_head_sha}...{pr_head_sha}`, scores each finding with **LLM** judgment, writes `findings-grade.json`, appends `close_completed` to `metrics.jsonl`, and swaps to **`workflow:done`**. Unrelated PR merges do not run this. See [metrics.md](metrics.md#close-close_completed) and [close-routine.md](../../routines/close-routine.md).

### On handoff write failure

1. Post a **short issue comment** — phase, push/commit failed, one-line error, session link.
2. **Stop** — do not swap labels, do not open PR.

## Read protocol

1. State branch: `workflow/state` (or `state_branch` in `state.json`).
2. Checkout / pull `workflow/state`; read `issues/{n}/state.json` (or `handoff_path`).
3. If state branch or `state.json` missing — clarify not started; implement must not run.
4. Load sibling files by path. For code/diff, checkout `work_branch` (set at implement start).

For implement: require `requirements_approved: true`.

**Legacy gist, v1 comment, or work-branch handoffs:** re-run clarify or migrate manually onto `workflow/state`.

## Session comments

Posted when **clarify**, **implement**, or **review** starts. See [label-rules.md](label-rules.md).

**Label swap is always the last GitHub write** when advancing phases (see [label-rules.md](label-rules.md)).
