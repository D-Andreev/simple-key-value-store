package cli

import (
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

			for _, k := range keys {
				cmd.Printf("%s=%s\n", k, items[k])
			}
			return nil
		},
	}
}
