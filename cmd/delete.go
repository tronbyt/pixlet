package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
)

type deleteOptions struct {
	apiToken string
	baseURL  string
}

func NewDeleteCmd() *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete [device ID] [installation ID]",
		Short: "Delete a Pixlet script from a Tronbyt",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteRun(cmd, args, opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return completeDevices(cmd)
			case 1:
				return completeInstallations(cmd, args[0])
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVarP(&opts.apiToken, "api-token", "t", opts.apiToken, "Tronbyt API token")
	_ = cmd.RegisterFlagCompletionFunc("api-token", cobra.NoFileCompletions)
	cmd.Flags().StringVarP(&opts.baseURL, "url", "u", opts.baseURL, "base URL of Tronbyt API")
	_ = cmd.RegisterFlagCompletionFunc("url", cobra.NoFileCompletions)

	return cmd
}

func deleteRun(_ *cobra.Command, args []string, opts *deleteOptions) error {
	deviceID := args[0]
	installationID := args[1]

	creds, err := resolveAPICredentials(opts.baseURL, opts.apiToken)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/v0/devices/%s/installations/%s", creds.baseURL, deviceID, installationID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("creating DELETE request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", creds.token))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("deleting via API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Tronbyt API returned an error", "status", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return fmt.Errorf("tronbyt API returned status: %s", resp.Status)
	}

	return nil
}
