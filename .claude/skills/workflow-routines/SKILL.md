---
name: workflow-routines
description: >-
  Shared rules for GitHub-issue workflow routines — handoff format, state
  schema, label triggers. Use when reading issue handoff or wiring routine
  webhooks.
disable-model-invocation: true
---

# Workflow Routines (shared reference)

## Label flow

```
workflow:start → clarify → workflow:implement → implement → workflow:review
  → AI review (autonomous, one PR comment) → workflow:human-review
  → [optional local comprehension] → merge PR → close issue
  → close routine (GitHub: issue closed + workflow:human-review) → workflow:done
```

No routine for `workflow:human-review` or comprehension. When the issue is **closed** while labeled `workflow:human-review`, the **close** routine grades AI review findings and sets **`workflow:done`**.

Docs: [handoff-format.md](handoff-format.md) · [state-schema.md](state-schema.md) · [label-rules.md](label-rules.md) · [metrics.md](metrics.md)

**Label swap is always the last GitHub write** when a phase advances. Handoff lives on long-lived **`workflow/state`**. Clarify writes only there; implement **creates the work branch**. Implement opens PRs as **draft**. Append analytics to `issues/{n}/metrics.jsonl` (clarify turns + review complete + close).

## Phase skills

| Phase | Skill | Routine trigger |
|-------|-------|-----------------|
| init | workflow-init | — |
| clarify | workflow-clarify | Label `workflow:start` |
| implement | workflow-implement | Label `workflow:implement` |
| AI review | workflow-review | Label `workflow:review` |
| human review | — | — (GitHub PR) |
| close | workflow-close | GitHub: issue closed + label `workflow:human-review` |
| comprehension | workflow-comprehension | — (local session, optional) |

Prompts: repo `routines/`.
