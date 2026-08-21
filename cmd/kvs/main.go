// Command kvs is a CLI for a key/value store, with optional on-disk
// persistence via --data-dir.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/D-Andreev/simple-key-value-store/internal/cli"
	"github.com/D-Andreev/simple-key-value-store/internal/store"
)

func main() {
	s, err := newStore(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "kvs: %v\n", err)
		os.Exit(1)
	}

	root := cli.NewRootCmd(s)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// newStore builds the Store the CLI runs against, based on --data-dir: a
// FileStore rooted at that directory if it's set, otherwise the default
// in-memory store (so existing scripts/tests that don't pass --data-dir
// see unchanged, in-memory-only behavior).
func newStore(args []string) (store.Store, error) {
	dataDir := dataDirFromArgs(args)
	if dataDir == "" {
		return store.NewMemoryStore(), nil
	}
	return store.NewFileStore(dataDir)
}

// dataDirFromArgs scans args for --data-dir ahead of Cobra's own flag
// parsing. Cobra only parses flags during root.Execute(), but the Store
// passed into cli.NewRootCmd has to be built before that call — so this
// does a manual pre-scan to find --data-dir early. --data-dir is also
// registered as a persistent flag on the root command (see
// cli.NewRootCmd) purely so it's documented in --help/usage text; this
// scan is the actual source of truth for which store gets used.
func dataDirFromArgs(args []string) string {
	const flag = "--data-dir"
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return args[i+1]
		}
		if v, ok := strings.CutPrefix(arg, flag+"="); ok {
			return v
		}
	}
	return ""
}
