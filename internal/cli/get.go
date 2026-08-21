package cli

import (
	"errors"
	"fmt"

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
			// cmd.Println falls back to stderr (not stdout) when no
			// explicit output writer is set, which is the case in the
			// real binary — write to OutOrStdout() directly so the
			// value actually lands on stdout.
			_, err = fmt.Fprintln(cmd.OutOrStdout(), value)
			return err
		},
	}
}
