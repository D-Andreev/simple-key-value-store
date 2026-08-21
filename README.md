# simple-key-value-store

A CLI for a simple key/value store. By default it's backed by an in-memory
`Store` (data does not survive process restarts); pass `--data-dir` to
persist data to disk instead.

## Build

```
go build -o kvs ./cmd/kvs
```

## Usage

```
kvs set <key> <value>
kvs get <key>
kvs delete <key>
kvs list
```

### On-disk persistence

Pass `--data-dir` to persist data to `<data-dir>/store.json`, so it survives
process restarts:

```
kvs --data-dir ~/.kvs set foo bar
kvs --data-dir ~/.kvs get foo   # -> bar, even after the process above exited
```

`--data-dir` is created automatically if it doesn't exist yet. Without
`--data-dir`, `kvs` behaves exactly as before: each invocation is its own
in-memory store, and nothing is written to disk.

## Test

```
go test ./...
```

