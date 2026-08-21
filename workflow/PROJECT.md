# Language

- **Store** — the interface abstracting key/value storage (`Get`, `Set`, `Delete`, `List`), decoupled from the CLI layer. String keys, string values.
- **In-memory store** — the initial `Store` implementation backed by a Go map; no persistence to disk. Replaceable later without touching CLI code.
- **kvs** — the CLI binary name (`cmd/kvs/main.go`).
