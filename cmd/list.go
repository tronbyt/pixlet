package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/cmd/groups"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

type listOptions struct {
	creds *flags.APICredentials
}

func NewListCmd() *cobra.Command {
	opts := &listOptions{
		creds: flags.NewAPICredentials(),
	}

	cmd := &cobra.Command{
		Use:     "list DEVICE_ID",
		GroupID: groups.Tronbyt,
		Short:   "Lists all apps installed on a Tronbyt",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listRun(cmd, args, opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeDevices(cmd, opts.creds)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	opts.creds.Register(cmd)
	return cmd
}

func listRun(cmd *cobra.Command, args []string, opts *listOptions) error {
	deviceID := args[0]

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	defer func() { _ = w.Flush() }()

	client, err := tronbytapi.NewClient(opts.creds.URL, opts.creds.APIToken)
	if err != nil {
		return err
	}

	for inst, err := range client.GetInstallations(cmd.Context(), deviceID) {
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\n", inst.ID, inst.AppID)
	}

	return nil
}
