# Requirements: issue-11

## Original ask
## Summary
Add full Create, Read, Update, and Delete operations for the key-value store, so callers can persist, retrieve, modify, and remove key/value pairs through the store's public interface.

## Motivation
The store currently doesn't expose a complete set of operations for managing data. Without Update and Delete, callers have no way to correct or remove a value once it's written, which blocks any real usage beyond initial inserts. CRUD is the minimum surface area needed for the store to be usable as an actual key-value store rather than an append-only log.

## Proposed solution
Expose four operations against the underlying store:

- **Create/Update (`Set`)** — `Set(key string, value []byte) error`: upserts a key, overwriting the value if the key already exists. A single upsert operation covers both create and update, since key-value stores don't typically distinguish the two at the API level.
- **Read (`Get`)** — `Get(key string) ([]byte, error)`: returns the value for a key, or a distinct "not found" error/sentinel if the key doesn't exist.
- **Delete (`Delete`)** — `Delete(key string) error`: removes a key. Deleting a non-existent key should not error (idempotent).
- **Exists (`Exists`)** — `Exists(key string) (bool, error)`: optional convenience to check presence without paying for a full value read.

## Scope
- [ ] `Set(key, value)` — create or overwrite a key's value
- [ ] `Get(key)` — retrieve a value, with a clear not-found signal
- [ ] `Delete(key)` — remove a key, idempotent on missing keys
- [ ] Concurrency-safe access if the store is used from multiple goroutines/requests

## Out of scope
- Range queries / listing all keys
- TTL / expiry
- Transactions or batch operations across multiple keys
- Persistence-format changes or a new storage backend

## Acceptance criteria
- [ ] Setting a key that doesn't exist creates it; getting it back returns the same value
- [ ] Setting a key that already exists overwrites the previous value
- [ ] Getting a key that was never set (or was deleted) returns a distinct not-found result, not a silent empty value
- [ ] Deleting a key removes it (subsequent `Get` returns not-found); deleting a missing key does not error
- [ ] Unit tests cover all four operations, including the not-found and overwrite cases

## Additional context
_(none)_

## Clarifications
| # | Question | Answer | Recommended |
|---|----------|--------|-------------|
| 1 | Should `Delete` on a missing key become idempotent (return `nil`) per the issue's literal wording, or keep the existing `ErrKeyNotFound` behavior the CLI already relies on for its "key not found" message? | Keep `Delete` erroring with `ErrKeyNotFound` on a missing key — no behavior change. | Keep `Delete` erroring with `ErrKeyNotFound` on a missing key (preserves existing CLI UX and test suite; "idempotent" read as end-state, not "always returns nil"). |

## Acceptance criteria
- [ ] Deleting a key removes it (subsequent `Get` returns `ErrKeyNotFound`); deleting a missing key also returns `ErrKeyNotFound` (unchanged from current behavior) — not a breaking change

## Approved by human
- [ ] Pending — say `approve requirements` in the session when ready
