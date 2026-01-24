package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

type devicesOptions struct {
	apiToken string
	baseURL  string
}

func NewDevicesCmd() *cobra.Command {
	opts := &devicesOptions{}

	cmd := &cobra.Command{
		Use:   "devices",
		Short: "List devices in your Tronbyt account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return devicesRun(opts)
		},
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	cmd.Flags().StringVarP(&opts.apiToken, "api-token", "t", opts.apiToken, "Tronbyt API token")
	_ = cmd.RegisterFlagCompletionFunc("api-token", cobra.NoFileCompletions)
	cmd.Flags().StringVarP(&opts.baseURL, "url", "u", opts.baseURL, "base URL of Tronbyt API")
	_ = cmd.RegisterFlagCompletionFunc("url", cobra.NoFileCompletions)

	return cmd
}

func devicesRun(opts *devicesOptions) error {
	creds, err := resolveAPICredentials(opts.baseURL, opts.apiToken)
	if err != nil {
		return err
	}

	for d, err := range getDevices(creds) {
		if err != nil {
			return err
		}

		fmt.Printf("%s (%s)\n", d.ID, d.DisplayName)
	}

	return nil
}

func getDevices(creds apiCredentials) iter.Seq2[*tronbytapi.Device, error] {
	return func(yield func(*tronbytapi.Device, error) bool) {
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v0/devices", creds.baseURL), nil)
		if err != nil {
			yield(nil, fmt.Errorf("creating request: %w", err))
			return
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", creds.token))

		resp, err := client.Do(req)
		if err != nil {
			yield(nil, fmt.Errorf("listing devices: %w", err))
			return
		}

		if resp.StatusCode != http.StatusOK {
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
