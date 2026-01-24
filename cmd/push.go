package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/internal/tronbytapi"
)

type pushOptions struct {
	apiToken       string
	installationID string
	background     bool
	baseURL        string
}

func NewPushCmd() *cobra.Command {
	opts := &pushOptions{}

	cmd := &cobra.Command{
		Use:   "push [device ID] [webp image]",
		Short: "Push a WebP to a Tronbyt",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pushRun(cmd, args, opts)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
			switch len(args) {
			case 0:
				return completeDevices(cmd)
			case 1:
				return []string{"webp"}, cobra.ShellCompDirectiveFilterFileExt
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringVarP(&opts.apiToken, "api-token", "t", opts.apiToken, "Tronbyt API token")
	_ = cmd.RegisterFlagCompletionFunc("api-token", cobra.NoFileCompletions)
	cmd.Flags().StringVarP(&opts.installationID, "installation-id", "i", opts.installationID, "Give your installation an ID to keep it in the rotation")
	_ = cmd.RegisterFlagCompletionFunc("installation-id", cobra.NoFileCompletions)
	cmd.Flags().BoolVarP(&opts.background, "background", "b", opts.background, "Don't immediately show the image on the device")
	cmd.Flags().StringVarP(&opts.baseURL, "url", "u", opts.baseURL, "base URL of Tronbyt API")
	_ = cmd.RegisterFlagCompletionFunc("url", cobra.NoFileCompletions)

	return cmd
}

func pushRun(_ *cobra.Command, args []string, opts *pushOptions) error {
	deviceID := args[0]
	image := args[1]

	// TODO (mark): This is better served as a flag, but I don't want to break
	// folks in the short term. We should consider dropping this as an argument
	// in a future release.
	if len(args) == 3 {
		opts.installationID = args[2]
	}

	if opts.background && len(opts.installationID) == 0 {
		return fmt.Errorf("background push won't do anything unless you also specify an installation ID")
	}

	creds, err := resolveAPICredentials(opts.baseURL, opts.apiToken)
	if err != nil {
		return err
	}

	imageData, err := os.ReadFile(image)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", image, err)
	}

	payload, err := json.Marshal(
		tronbytapi.PushPayload{
			DeviceID:       deviceID,
			Image:          base64.StdEncoding.EncodeToString(imageData),
			InstallationID: opts.installationID,
			Background:     opts.background,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v0/devices/%s/push", creds.baseURL, deviceID),
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("creating POST request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", creds.token))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("pushing to API: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Tronbyt API returned an error", "status", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return fmt.Errorf("tronbyt API returned status: %s", resp.Status)
	}

	return nil
}
