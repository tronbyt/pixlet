package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/runtime"
)

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version of Pixlet",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Pixlet version: %s\n", runtime.Version)
		},
		ValidArgsFunction: cobra.NoFileCompletions,
	}
	return cmd
}
