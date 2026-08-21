# Language

- **Store** — the interface abstracting key/value storage (`Get`, `Set`, `Delete`, `List`), decoupled from the CLI layer. String keys, string values. `Set` returns an error so implementations backed by disk can report a failed write.
- **In-memory store** — the initial `Store` implementation backed by a Go map; no persistence to disk. Replaceable later without touching CLI code.
- **File store** — a `Store` implementation backed by a JSON file (`store.json`) inside a `--data-dir`. Loads existing data on construction and flushes the full snapshot to disk synchronously on every `Set`/`Delete` (no batching — `kvs` is a one-shot CLI process, so there's no later point to defer a flush to). Auto-creates `--data-dir` if missing; fails loudly (returns an error rather than silently discarding data) if `store.json` contains invalid JSON, or if `--data-dir` names an existing regular file instead of a directory.
- **--data-dir** — CLI flag selecting on-disk persistence. `main.go` pre-scans `os.Args` for it (before Cobra's own flag parsing runs, since the `Store` has to exist before `cli.NewRootCmd` is called) to choose between an in-memory store and a file store; it's also registered as a persistent flag on the root command purely so it's documented in `--help`. Omitting it preserves the original in-memory-only behavior.
- **kvs** — the CLI binary name (`cmd/kvs/main.go`).
