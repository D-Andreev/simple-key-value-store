package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/D-Andreev/simple-key-value-store/internal/store"
)

func newGetCmd(s store.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get the value stored under a key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			value, err := s.Get(args[0])
			if err != nil {
				if errors.Is(err, store.ErrKeyNotFound) {
					cmd.PrintErrln("key not found")
				}
				return err
			}
			cmd.Println(value)
			return nil
		},
	}
}
