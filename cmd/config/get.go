package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets a value from the config file.",
	Example: `  pixlet config get ` + URLKey + `
  pixlet config get ` + TokenKey,
	Long:      `This command gets a value from the config file.`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{URLKey, TokenKey},
	RunE:      getRun,
}

var ErrKeyNotFound = fmt.Errorf("key not found")

func getRun(cmd *cobra.Command, args []string) error {
	key := args[0]

	if !Config.IsSet(key) {
		return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	fmt.Println(Config.Get(key))
	return nil
}
