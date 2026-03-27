package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/cmd/groups"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

type devicesOptions struct {
	creds *flags.APICredentials
}

func NewDevicesCmd() *cobra.Command {
	opts := &devicesOptions{
		creds: flags.NewAPICredentials(),
	}

	cmd := &cobra.Command{
		Use:     "devices",
		GroupID: groups.Tronbyt,
		Short:   "List devices in your Tronbyt account",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return devicesRun(cmd, opts)
		},
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	opts.creds.Register(cmd)
	return cmd
}

func devicesRun(cmd *cobra.Command, opts *devicesOptions) error {
	client, err := tronbytapi.NewClient(opts.creds.URL, opts.creds.APIToken)
	if err != nil {
		return err
	}

	for d, err := range client.GetDevices(cmd.Context()) {
		if err != nil {
			return err
		}

		fmt.Printf("%s (%s)\n", d.ID, d.DisplayName)
	}

	return nil
}
