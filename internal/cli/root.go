// Package cli wires the kvs command-line interface: a root command with
// set/get/delete/list subcommands, each operating on a store.Store.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/D-Andreev/simple-key-value-store/internal/store"
)

// NewRootCmd builds the root "kvs" command with all subcommands wired
// against the given store.
func NewRootCmd(s store.Store) *cobra.Command {
	root := &cobra.Command{
		Use:           "kvs",
		Short:         "kvs is a simple key/value store CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// --data-dir selects on-disk persistence (a FileStore backed by
	// <data-dir>/store.json) instead of the default in-memory-only
	// store. Cobra only parses flags once root.Execute() runs, but the
	// Store passed into NewRootCmd already has to exist by then — so
	// main.go pre-scans os.Args for --data-dir and constructs the store
	// before calling NewRootCmd. Registering it here doesn't drive that
	// choice; it exists purely so --data-dir shows up in --help/usage.
	root.PersistentFlags().String("data-dir", "", "directory for on-disk persistence (default: in-memory only)")

	root.AddCommand(newSetCmd(s))
	root.AddCommand(newGetCmd(s))
	root.AddCommand(newDeleteCmd(s))
	root.AddCommand(newListCmd(s))

	return root
}
