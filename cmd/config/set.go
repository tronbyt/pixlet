package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

var SetCmd = &cobra.Command{
	Use:   "set",
	Short: "Saves a value to the config file.",
	Example: `  pixlet config set ` + URLKey + ` <tronbyt_url>
  pixlet config set ` + TokenKey + ` <user_token>`,
	Long:      `This command saves a value to the config file for use in subsequent runs.`,
	Args:      cobra.ExactArgs(2),
	ValidArgs: []string{URLKey, TokenKey},
	RunE:      setRun,
}

func setRun(cmd *cobra.Command, args []string) error {
	key, val := args[0], args[1]

	Config.Set(key, val)
	if err := Config.WriteConfig(); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}
