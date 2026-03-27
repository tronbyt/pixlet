package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/flags"
	"github.com/tronbyt/pixlet/cmd/groups"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

type pushOptions struct {
	creds          *flags.APICredentials
	installationID string
	background     bool
}

func NewPushCmd() *cobra.Command {
	opts := &pushOptions{
		creds: flags.NewAPICredentials(),
	}

	cmd := &cobra.Command{
		Use:     "push DEVICE_ID WEBP",
		GroupID: groups.Tronbyt,
		Short:   "Push a WebP to a Tronbyt",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pushRun(cmd, args, opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return completeDevices(cmd, opts.creds)
			case 1:
				return []string{"webp"}, cobra.ShellCompDirectiveFilterFileExt
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVarP(&opts.installationID, "installation-id", "i", opts.installationID, "Give your installation an ID to keep it in the rotation")
	_ = cmd.RegisterFlagCompletionFunc("installation-id", cobra.NoFileCompletions)
	cmd.Flags().BoolVarP(&opts.background, "background", "b", opts.background, "Don't immediately show the image on the device")

	opts.creds.Register(cmd)
	return cmd
}

func pushRun(cmd *cobra.Command, args []string, opts *pushOptions) error {
	deviceID := args[0]
	image := args[1]

	// TODO (mark): This is better served as a flag, but I don't want to break
	// folks in the short term. We should consider dropping this as an argument
	// in a future release.
	if len(args) == 3 {
		opts.installationID = args[2]
	}

	client, err := tronbytapi.NewClient(opts.creds.URL, opts.creds.APIToken)
	if err != nil {
		return err
	}

	f, err := os.Open(image)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", image, err)
	}
	defer func() { _ = f.Close() }()

	return client.Push(cmd.Context(), deviceID, f, &tronbytapi.PushOptions{
		InstallationID: opts.installationID,
		Background:     opts.background,
	})
}
