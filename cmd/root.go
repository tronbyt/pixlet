package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/community"
	"github.com/tronbyt/pixlet/cmd/config"
	"github.com/tronbyt/pixlet/cmd/groups"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "pixlet",
		Short:        "pixel graphics rendering",
		Long:         "Pixlet renders graphics for pixel devices, like Tronbyt.",
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

	cmd.AddGroup(
		&cobra.Group{
			ID:    groups.Applet,
			Title: "Applet Commands:",
		},
		&cobra.Group{
			ID:    groups.Validate,
			Title: "Validation Commands:",
		},
		&cobra.Group{
			ID:    groups.Tronbyt,
			Title: "Tronbyt Commands:",
		},
	)

	return cmd
}
