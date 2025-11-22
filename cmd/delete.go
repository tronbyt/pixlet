package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/config"
)

var deleteURL string

func init() {
	DeleteCmd.Flags().StringVarP(&apiToken, "api-token", "t", "", "Tronbyt API token")
	_ = DeleteCmd.RegisterFlagCompletionFunc("api-token", cobra.NoFileCompletions)
	DeleteCmd.Flags().StringVarP(&deleteURL, "url", "u", "", "base URL of Tronbyt API")
	_ = DeleteCmd.RegisterFlagCompletionFunc("url", cobra.NoFileCompletions)
}

var DeleteCmd = &cobra.Command{
	Use:   "delete [device ID] [installation ID]",
	Short: "Delete a Pixlet script from a Tronbyt",
	Args:  cobra.MinimumNArgs(2),
	RunE:  delete,
	ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return completeDevices()
		case 1:
			return completeInstallations(args[0])
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func delete(cmd *cobra.Command, args []string) error {
	deviceID := args[0]
	installationID := args[1]

	if deleteURL == "" {
		var err error
		if deleteURL, err = config.GetURL(); err != nil {
			return err
		}
	}

	if apiToken == "" {
		var err error
		if apiToken, err = config.GetToken(); err != nil {
			return err
		}
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/v0/devices/%s/installations/%s", deleteURL, deviceID, installationID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("creating DELETE request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("deleting via API: %w", err)
	}

	if resp.StatusCode != 200 {
		slog.Error("Tronbyt API returned an error", "status", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return fmt.Errorf("Tronbyt API returned status: %s", resp.Status)
	}

	return nil
}
