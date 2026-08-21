---
name: workflow-init
description: >-
  Initialize the AI workflow for a project — scaffold workflow/ directories,
  seed gotchas.md, generate PROJECT.md, create GitHub workflow labels, and
  create the long-lived workflow/state branch. Use when setting up a new
  repo, when PROJECT.md is missing, or when the user runs /workflow-init.
metadata:
  internal: true
---

# Workflow: Init

One-time setup. Creates long-lived **`workflow/state`** for all issue handoffs (`issues/{n}/`). 

**No application code changes.**

## Process

Ensure the `wfr` CLI is installed, then run it — it owns the whole process below and is idempotent (safe to re-run):

```bash
command -v wfr >/dev/null 2>&1 || curl -fsSL https://raw.githubusercontent.com/D-Andreev/ai-workflow-routines/main/scripts/install.sh | sh
wfr init
```

`wfr init` installs `gh` itself on first use if it's missing from the sandbox. If GitHub still isn't usable afterward (install failed or not authenticated), it aborts immediately with an error before touching anything — no partial local writes. Otherwise it, in order:

1. Scaffolds `workflow/learnings/` and seeds `workflow/learnings/gotchas.md` if missing
2. Generates `workflow/PROJECT.md` if missing — never overwrites an existing one
3. **Creates GitHub labels** (`--force`, so re-runs update description/color without error)
4. **Ensures the long-lived `workflow/state` branch** (orphan; never merge to main) — no-op if `origin/workflow/state` already exists


## Verify

To check whether a repo is already initialized, without changing anything:

```bash
command -v wfr >/dev/null 2>&1 || curl -fsSL https://raw.githubusercontent.com/D-Andreev/ai-workflow-routines/main/scripts/install.sh | sh
wfr verify
```

Reports ✅/❌ for local files, the `workflow/state` branch, and the GitHub labels, and exits non-zero if anything's missing.
