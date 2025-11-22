package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/cmd/config"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

var devicesURL string

func NewDevicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "devices",
		Short:             "List devices in your Tronbyt account",
		RunE:              devicesRun,
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	cmd.Flags().StringVarP(&apiToken, "api-token", "t", "", "Tronbyt API token")
	_ = cmd.RegisterFlagCompletionFunc("api-token", cobra.NoFileCompletions)
	cmd.Flags().StringVarP(&devicesURL, "url", "u", "", "base URL of Tronbyt API")
	_ = cmd.RegisterFlagCompletionFunc("url", cobra.NoFileCompletions)

	return cmd
}

func devicesRun(_ *cobra.Command, _ []string) error {
	for d, err := range getDevices() {
		if err != nil {
			return err
		}

		fmt.Printf("%s (%s)\n", d.ID, d.DisplayName)
	}

	return nil
}

func getDevices() iter.Seq2[*tronbytapi.Device, error] {
	return func(yield func(*tronbytapi.Device, error) bool) {
		if devicesURL == "" {
			var err error
			if devicesURL, err = config.GetURL(); err != nil {
				yield(nil, err)
				return
			}
		}

		if apiToken == "" {
			var err error
			if apiToken, err = config.GetToken(); err != nil {
				yield(nil, err)
				return
			}
		}

		client := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/v0/devices", devicesURL), nil)
		if err != nil {
			yield(nil, fmt.Errorf("creating request: %w", err))
			return
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))

		resp, err := client.Do(req)
		if err != nil {
			yield(nil, fmt.Errorf("listing devices: %w", err))
			return
		}

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			yield(nil, fmt.Errorf("tronbyt api error %s: %s", resp.Status, body))
			return
		}

		var devices tronbytapi.Devices
		if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
			yield(nil, fmt.Errorf("decoding json: %w", err))
			return
		}

		for _, device := range devices.Devices {
			yield(&device, nil)
		}
	}
}
