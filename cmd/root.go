package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/community"
	"github.com/tronbyt/pixlet/cmd/config"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "pixlet",
		Short:        "pixel graphics rendering",
		Long:         "Pixlet renders graphics for pixel devices, like Tronbyt",
		SilenceUsage: true,
	}

	cmd.AddCommand(
		NewAPICmd(),
		NewCheckCmd(),
		NewCreateCmd(),
		NewDeleteCmd(),
		NewDevicesCmd(),
		NewFormatCmd(),
		NewLintCmd(),
		NewListCmd(),
		NewProfileCmd(),
		NewPushCmd(),
		NewRenderCmd(),
		NewSchemaCmd(),
		NewServeCmd(),
		NewVersionCmd(),
		community.NewCommunityCmd(),
		config.NewConfigCmd(),
	)

	return cmd
}
