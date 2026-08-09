# Workflow metrics (`metrics.jsonl`)

Append-only analytics log on **`workflow/state`** at `issues/{n}/metrics.jsonl`.

A later job reads this file (plus `state.json`) — agents **only append**, never rewrite or delete lines.

Full handoff layout: [handoff-format.md](handoff-format.md)

## File rules

| Rule | Detail |
|------|--------|
| Path | `issues/{n}/metrics.jsonl` on `workflow/state` |
| Format | One JSON object per line (JSONL); UTF-8; no trailing commas |
| Create | Touch empty file at clarify start (or create on first append) |
| Write | **Append** one event, then commit with other handoff updates |
| Do not | Edit prior lines, pretty-print the whole file, or put metrics on the work branch |

## Common fields (every event)

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `ts` | string (ISO 8601) | yes | Event time, UTC |
| `issue_number` | number | yes | GitHub issue number |
| `phase` | string | yes | `clarify` \| `implement` \| `review` \| `close` |
| `event` | string | yes | Event type (below) |
| `session_id` | string \| null | no | Claude session id when known |
| `schema_version` | number | yes | Always `1` for this schema |

---

## Clarify: `clarify_turn`

Append **once per answered (or skipped) question**, when committing the Q&A handoff update — not when asking the question.

```json
{
  "schema_version": 1,
  "ts": "2026-07-28T10:15:00.000Z",
  "issue_number": 42,
  "phase": "clarify",
  "event": "clarify_turn",
  "session_id": null,
  "q_index": 1,
  "category": "edge_case",
  "recommendation_outcome": "accepted_recommendation",
  "question": "Should retries apply to all notification channels or email only?"
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `q_index` | number | yes | 1-based; matches Clarifications table `#` |
| `category` | string | yes | **Only** values from [Question categories](#question-categories) |
| `recommendation_outcome` | string | yes | **Only** values from [Recommendation outcomes](#recommendation-outcomes) |
| `question` | string | yes | Short question text (one line preferred) |

### Question categories

Pick **exactly one** per question. Do not invent new categories.

| `category` | Use when the question is mainly about… |
|------------|------------------------------------------|
| `scope` | In/out of this issue; deferrals; MVP vs later |
| `behavior` | Happy-path product/UX behavior |
| `edge_case` | Boundaries, empty/null, concurrency, races, odd inputs |
| `data_model` | Entities, fields, persistence, migrations |
| `api_contract` | APIs, payloads, compatibility, versioning |
| `error_handling` | Failures, retries, timeouts, user-visible errors |
| `security` | Authn/authz, secrets, PII, tenancy |
| `performance` | Latency, throughput, caching, scale |
| `testing` | What/how to test; coverage expectations |
| `ops` | Config, flags, deploy, observability, runbooks |
| `domain_language` | Naming, terms for `language.md` |
| `dependency` | Libraries, services, external systems |

### Recommendation outcomes

Judge the human’s **session** answer against the recommended answer you offered. Use **exactly one**:

| `recommendation_outcome` | Meaning |
|--------------------------|---------|
| `skipped` | Human skipped / deferred / said to skip this question |
| `accepted_recommendation` | Went with the recommendation as-is (or equivalent) |
| `accepted_with_adjustment` | Accepted the recommendation but changed details |
| `rejected_recommendation` | Chose a different approach than the recommendation |

---

## Review: `review_completed`

Append **once** at review complete, in the same commit as `review-report.md` / final `state.json`.

```json
{
  "schema_version": 1,
  "ts": "2026-07-28T14:02:00.000Z",
  "issue_number": 42,
  "phase": "review",
  "event": "review_completed",
  "session_id": null,
  "verdict": "APPROVE WITH NOTES",
  "critical_count": 0,
  "minor_count": 1,
  "notes_count": 2
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `verdict` | string | yes | `APPROVE` \| `APPROVE WITH NOTES` \| `REQUEST CHANGES` |
| `critical_count` | number | yes | 🔴 must-fix / critical findings (integer ≥ 0) |
| `minor_count` | number | yes | 🟡 minor findings (integer ≥ 0) |
| `notes_count` | number | yes | Non-blocking notes / suggestions not counted as critical/minor (integer ≥ 0) |

**Counting rules**

- Align with `review-report.md` (Issues found + Principles critical/suggestions).
- `critical_count` → must-fix / 🔴 / Critical sections.
- `minor_count` → 🟡 / minor issues that are still “findings.”
- `notes_count` → shippable nits and “nice to have” that are not minor findings.
- Do not double-count the same finding across buckets.
- `APPROVE` usually has all zeros; `APPROVE WITH NOTES` usually has `notes_count` and/or `minor_count` > 0 and `critical_count` = 0; `REQUEST CHANGES` usually has `critical_count` ≥ 1.

At the same commit, also write **`review-findings.json`** (structured checklist) and set **`review_head_sha`** on `state.json` — see [Review findings checklist](#review-findings-checklist).

---

## Review findings checklist

Machine-readable findings written **once** at AI review complete to `issues/{n}/review-findings.json` on `workflow/state`. The **close routine** (or `/workflow-close`) scores whether humans addressed each item.

**Fixture:** [fixtures/review-findings-example.json](fixtures/review-findings-example.json)

```json
{
  "schema_version": 1,
  "issue_number": 42,
  "pr_number": 17,
  "review_head_sha": "a1b2c3d4e5f6789012345678901234567890abcd",
  "verdict": "APPROVE WITH NOTES",
  "created_at": "2026-07-26T14:02:00.000Z",
  "findings": [
    {
      "id": "F1",
      "severity": "minor",
      "summary": "Retry helper does not cap max delay",
      "paths": ["src/notifications/retry.ts"],
      "required": false
    }
  ]
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `review_head_sha` | string | yes | Full SHA of `work_branch` tip **at review time** (`git rev-parse HEAD` on the work branch before leaving review). Anchor for post-review diffs. |
| `pr_number` | number \| null | no | Draft/open PR number when known |
| `verdict` | string | yes | Same as `review_completed` |
| `findings[]` | array | yes | May be empty when `APPROVE` with nothing to track |
| `findings[].id` | string | yes | Stable id within the issue: `F1`, `F2`, … |
| `findings[].severity` | string | yes | `critical` \| `minor` \| `note` |
| `findings[].summary` | string | yes | One-line finding (matches review-report) |
| `findings[].paths` | string[] | yes | Repo-relative paths (files or dirs). Prefer concrete files. Empty only when truly unknown. |
| `findings[].required` | boolean | yes | `true` for `critical`; `false` for `minor` / `note` |

**Severity mapping from review-report**

| Report section | `severity` | `required` |
|----------------|------------|------------|
| 🔴 Critical / Critical (must fix) | `critical` | `true` |
| 🟡 Minor / Suggestions (should consider) | `minor` | `false` |
| Nice to have / non-blocking notes | `note` | `false` |

Do not duplicate the same finding across severities. Counts in `review_completed` must match `findings[]` lengths by severity.

---

## Close: `close_completed`

Append **once** when the **close** routine finishes findings grading (issue close + `workflow:human-review`, or `/workflow-close`). Same commit should write `findings-grade.json`.

```json
{
  "schema_version": 1,
  "ts": "2026-07-28T18:00:00.000Z",
  "issue_number": 42,
  "phase": "close",
  "event": "close_completed",
  "session_id": null,
  "method": "llm",
  "suggestions_applicable": true,
  "findings_total": 2,
  "findings_addressed": 1,
  "findings_partial": 0,
  "findings_ignored": 1,
  "findings_unknown": 0,
  "critical_total": 0,
  "critical_addressed": 0,
  "minor_total": 1,
  "minor_addressed": 1,
  "notes_total": 1,
  "notes_addressed": 0,
  "commits_since_review": 3,
  "dispositions": [
    {
      "id": "F1",
      "severity": "minor",
      "summary": "Retry helper does not cap max delay",
      "required": false,
      "disposition": "addressed",
      "paths_touched": 1,
      "paths_total": 1,
      "method": "llm"
    }
  ]
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `method` | string | yes | `llm` (routine default) \| `path_heuristic` (optional manual) |
| `suggestions_applicable` | boolean | yes | `false` when findings list was empty |
| `findings_*` | number | yes | Totals by disposition |
| `critical_*` / `minor_*` / `notes_*` | number | yes | Per-severity totals; `*_addressed` counts `addressed` **or** `partial` |
| `commits_since_review` | number | yes | `git rev-list --count {review_head_sha}..{pr_head_sha}` |
| `dispositions[]` | array | yes | One entry per finding (empty if none) |

### Dispositions

| `disposition` | Meaning |
|---------------|---------|
| `addressed` | Finding implemented (LLM: clear fix in diff; optional heuristic: all listed paths touched) |
| `partial` | Some related change (LLM: incomplete fix; optional heuristic: some paths touched) |
| `ignored` | No related change |
| `unknown` | Cannot judge (no paths / ambiguous) |

### Close graders

| Grader | When | Method |
|--------|------|--------|
| **Close** routine | Issue **closed** with label `workflow:human-review` (GitHub event) | `llm` — follow workflow-close |
| `/workflow-close` | Manual / re-run | `llm` |

Idempotent: skip if `findings-grade.json` already exists (unless explicit re-run).

No GitHub Action — close is a cloud routine like clarify / implement / review.

---

## Append recipe

```bash
# On workflow/state, after preparing the event object:
printf '%s\n' '{"schema_version":1,...}' >> issues/{n}/metrics.jsonl
git add issues/{n}/metrics.jsonl
# commit with the rest of the handoff update
```

Validate mentally: single line, valid JSON, enums only from this doc.
