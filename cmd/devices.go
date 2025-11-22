package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/config"
)

var devicesURL string

func init() {
	DevicesCmd.Flags().StringVarP(&devicesURL, "url", "u", "https://api.tidbyt.com", "base URL of Tidbyt API")
}

var DevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List devices in your Tidbyt account",
	Run:   devices,
}

func devices(cmd *cobra.Command, args []string) {
	apiToken = config.OAuthTokenFromConfig(cmd.Context())
	if apiToken == "" {
		slog.Error("Login with `pixlet login`")
		os.Exit(1)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v0/devices", devicesURL), nil)
	if err != nil {
		slog.Error("Creating GET request", "error", err)
		os.Exit(1)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Listing devices from API", "error", err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		slog.Error("Tidbyt API returned an error", "status", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		os.Exit(1)
	}

	body := struct {
		Devices []struct {
			ID          string `json:"id"`
			DisplayName string `json:"displayName"`
		} `json:"devices"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		slog.Error("Decoding API response", "error", err)
		os.Exit(1)
	}

	for _, d := range body.Devices {
		fmt.Printf("%s (%s)\n", d.ID, d.DisplayName)
	}
}
