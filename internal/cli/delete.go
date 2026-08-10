package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/D-Andreev/simple-key-value-store/internal/store"
)

func newDeleteCmd(s store.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a key from the store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := s.Delete(args[0]); err != nil {
				if errors.Is(err, store.ErrKeyNotFound) {
					cmd.PrintErrln("key not found")
				}
				return err
			}
			return nil
		},
	}
}
