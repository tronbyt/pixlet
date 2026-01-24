package community

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/icons"
)

func NewListIconsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list-icons",
		Short:             "List icons that are available in our mobile app.",
		Example:           `  pixlet community list-icons`,
		Long:              `This command lists all in your icons that are supported by our mobile app.`,
		RunE:              listIconsRun,
		ValidArgsFunction: cobra.NoFileCompletions,
	}
	return cmd
}

func listIconsRun(_ *cobra.Command, _ []string) error {
	iconSet := make([]string, 0, len(icons.IconsMap))
	for icon := range icons.IconsMap {
		iconSet = append(iconSet, icon)
	}

	sort.Strings(iconSet)
	for _, icon := range iconSet {
		fmt.Println(icon)
	}

	return nil
}
