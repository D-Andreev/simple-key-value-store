# simple-key-value-store

A CLI for a simple key/value store, backed by an in-memory `Store`.

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

## Test

```
go test ./...
```

