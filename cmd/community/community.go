package community

import (
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/groups"
)

func NewCommunityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "community",
		GroupID: groups.Applet,
		Short:   "Utilities to manage the community repo",
		Long: `The community subcommand provides a set of utilities for managing the
community repo. This subcommand should be considered slightly unstable in that
we may determine a utility here should move to a more generalizable tool.`,
	}

	cmd.AddCommand(
		NewCreateManifestCmd(),
		NewListColorFiltersCmd(),
		NewListFontsCmd(),
		NewListIconsCmd(),
		NewLoadAppCmd(),
		NewValidateIconsCmd(),
		NewValidateManifestCmd(),
	)

	return cmd
}
