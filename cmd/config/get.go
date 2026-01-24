package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Gets a value from the config file.",
		Example: `  pixlet config get ` + URLKey + `
  pixlet config get ` + TokenKey,
		Long: `This command gets a value from the config file.`,
		Args: cobra.ExactArgs(1),
		RunE: getRun,
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return []string{URLKey, TokenKey}, cobra.ShellCompDirectiveNoFileComp
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}
	return cmd
}

var ErrKeyNotFound = fmt.Errorf("key not found")

func getRun(_ *cobra.Command, args []string) error {
	InitConfig()
	key := args[0]

	if !Config.IsSet(key) {
		return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	fmt.Println(Config.Get(key))
	return nil
}
