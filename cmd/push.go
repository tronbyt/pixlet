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
	"github.com/tronbyt/pixlet/cmd/config"
)

var (
	apiToken       string
	installationID string
	background     bool
	pushURL        string
)

type TidbytPushJSON struct {
	DeviceID       string `json:"deviceID"`
	Image          string `json:"image"`
	InstallationID string `json:"installationID"`
	Background     bool   `json:"background"`
}

func init() {
	PushCmd.Flags().StringVarP(&apiToken, "api-token", "t", "", "Tronbyt API token")
	_ = PushCmd.RegisterFlagCompletionFunc("api-token", cobra.NoFileCompletions)
	PushCmd.Flags().StringVarP(&installationID, "installation-id", "i", "", "Give your installation an ID to keep it in the rotation")
	_ = PushCmd.RegisterFlagCompletionFunc("installation-id", cobra.NoFileCompletions)
	PushCmd.Flags().BoolVarP(&background, "background", "b", false, "Don't immediately show the image on the device")
	PushCmd.Flags().StringVarP(&pushURL, "url", "u", "", "base URL of Tronbyt API")
	_ = PushCmd.RegisterFlagCompletionFunc("url", cobra.NoFileCompletions)
}

var PushCmd = &cobra.Command{
	Use:   "push [device ID] [webp image]",
	Short: "Push a WebP to a Tronbyt",
	Args:  cobra.MinimumNArgs(2),
	RunE:  push,
	ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return completeDevices()
		case 1:
			return []string{"webp"}, cobra.ShellCompDirectiveFilterFileExt
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
}

func push(cmd *cobra.Command, args []string) error {
	deviceID := args[0]
	image := args[1]

	// TODO (mark): This is better served as a flag, but I don't want to break
	// folks in the short term. We should consider dropping this as an argument
	// in a future release.
	if len(args) == 3 {
		installationID = args[2]
	}

	if background && len(installationID) == 0 {
		return fmt.Errorf("background push won't do anything unless you also specify an installation ID")
	}

	if pushURL == "" {
		var err error
		if pushURL, err = config.GetURL(); err != nil {
			return err
		}
	}

	if apiToken == "" {
		var err error
		if apiToken, err = config.GetToken(); err != nil {
			return err
		}
	}

	imageData, err := os.ReadFile(image)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", image, err)
	}

	payload, err := json.Marshal(
		TidbytPushJSON{
			DeviceID:       deviceID,
			Image:          base64.StdEncoding.EncodeToString(imageData),
			InstallationID: installationID,
			Background:     background,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v0/devices/%s/push", pushURL, deviceID),
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("creating POST request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("pushing to API: %w", err)
	}

	if resp.StatusCode != 200 {
		slog.Error("Tronbyt API returned an error", "status", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return fmt.Errorf("Tronbyt API returned status: %s", resp.Status)
	}

	return nil
}
