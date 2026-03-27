package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/cmd/groups"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

type deleteOptions struct {
	creds *flags.APICredentials
}

func NewDeleteCmd() *cobra.Command {
	opts := &deleteOptions{
		creds: flags.NewAPICredentials(),
	}

	cmd := &cobra.Command{
		Use:     "delete DEVICE_ID INSTALLATION_ID",
		GroupID: groups.Tronbyt,
		Short:   "Delete a Pixlet script from a Tronbyt",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteRun(cmd, args, opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return completeDevices(cmd, opts.creds)
			case 1:
				return completeInstallations(cmd, opts.creds, args[0])
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	opts.creds.Register(cmd)
	return cmd
}

func deleteRun(cmd *cobra.Command, args []string, opts *deleteOptions) error {
	deviceID := args[0]
	installationID := args[1]

	client, err := tronbytapi.NewClient(opts.creds.URL, opts.creds.APIToken)
	if err != nil {
		return err
	}

	return client.Delete(cmd.Context(), deviceID, installationID)
}
