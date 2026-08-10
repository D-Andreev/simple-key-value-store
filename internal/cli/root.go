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

	root.AddCommand(newSetCmd(s))
	root.AddCommand(newGetCmd(s))
	root.AddCommand(newDeleteCmd(s))
	root.AddCommand(newListCmd(s))

	return root
}
