package community

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/encode"
)

var ListColorFiltersCmd = &cobra.Command{
	Use:     "list-color-filters",
	Short:   "List supported color filters.",
	Example: `  pixlet community list-color-filters`,
	Long:    `This command lists all color filters.`,
	RunE:    listColorFilters,
}

func listColorFilters(cmd *cobra.Command, _ []string) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	if _, err := w.Write([]byte("NAME\tDESCRIPTION\n")); err != nil {
		return err
	}

	for _, f := range encode.ColorFilterValues() {
		desc, _ := f.Description()
		if _, err := fmt.Fprintf(
			w, "%s\t%s\n",
			f.String(), desc,
		); err != nil {
			return err
		}
	}

	return w.Flush()
}
