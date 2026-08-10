package cli

import (
	"github.com/spf13/cobra"

	"github.com/D-Andreev/simple-key-value-store/internal/store"
)

func newSetCmd(s store.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a key to a value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s.Set(args[0], args[1])
			return nil
		},
	}
}
