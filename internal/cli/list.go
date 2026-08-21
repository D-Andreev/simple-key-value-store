package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/D-Andreev/simple-key-value-store/internal/store"
)

func newListCmd(s store.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all key/value pairs, sorted by key",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			items := s.List()
			keys := make([]string, 0, len(items))
			for k := range items {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// cmd.Printf falls back to stderr (not stdout) when no
			// explicit output writer is set, which is the case in the
			// real binary — write to OutOrStdout() directly so listed
			// pairs actually land on stdout.
			out := cmd.OutOrStdout()
			for _, k := range keys {
				if _, err := fmt.Fprintf(out, "%s=%s\n", k, items[k]); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
