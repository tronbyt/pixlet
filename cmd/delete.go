package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"tidbyt.dev/pixlet/cmd/config"
)

var deleteURL string

func init() {
	DeleteCmd.Flags().StringVarP(&apiToken, "api-token", "t", "", "Tidbyt API token")
	DeleteCmd.Flags().StringVarP(&deleteURL, "url", "u", "https://api.tidbyt.com", "base URL of Tidbyt API")
}

var DeleteCmd = &cobra.Command{
	Use:   "delete [device ID] [installation ID]",
	Short: "Delete a pixlet script from a Tidbyt",
	Args:  cobra.MinimumNArgs(2),
	RunE:  delete,
}

func delete(cmd *cobra.Command, args []string) error {
	deviceID := args[0]
	installationID := args[1]

	if apiToken == "" {
		apiToken = os.Getenv(APITokenEnv)
	}

	if apiToken == "" {
		apiToken = config.OAuthTokenFromConfig(cmd.Context())
	}

	if apiToken == "" {
		return fmt.Errorf("blank Tidbyt API token (use `pixlet login`, set $%s or pass with --api-token)", APITokenEnv)
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
		fmt.Printf("Tidbyt API returned status %s\n", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return fmt.Errorf("Tidbyt API returned status: %s", resp.Status)
	}

	return nil
}
