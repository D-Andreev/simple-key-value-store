// Command kvs is a CLI for a simple in-memory key/value store.
package main

import (
	"os"

	"github.com/D-Andreev/simple-key-value-store/internal/cli"
	"github.com/D-Andreev/simple-key-value-store/internal/store"
)

func main() {
	s := store.NewMemoryStore()
	root := cli.NewRootCmd(s)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
